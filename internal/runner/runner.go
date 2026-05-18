package runner

import (
	"context"
	"fmt"
	"sync"
	"time"
	"todo-runner-cli/internal/task"
)

type TaskService interface {
	UpdateTask(taskResult task.Result) (task.Task, error)
	QueuedTasks() []task.Task
	GetStats() task.TaskStats
}

type TaskProcessor interface {
	Process(ctx context.Context, task task.Task) (task.Result, error)
}

func PushTasksToQueue(ctx context.Context, tasksChan chan<- task.Task, service TaskService) {
	defer close(tasksChan)

	for _, current := range service.QueuedTasks() {
		select {
		case tasksChan <- current:
			continue
		case <-ctx.Done():
			return
		}
	}
}

func ReadProcessed(processed <-chan task.Result, service TaskService) {
	for result := range processed {
		_, err := service.UpdateTask(result)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}

func StartRunner(ctx context.Context, tasks <-chan task.Task, processed chan<- task.Result, processor TaskProcessor, service TaskService) {
	fmt.Println("Runner started")
	for {
		select {
		case <-ctx.Done():
			return
		case current, ok := <-tasks:
			if !ok {
				return
			}
			result := task.Result{ID: current.ID, Name: current.Name, Status: task.Running}
			_, err := service.UpdateTask(result)

			if err != nil {
				continue
			}

			resultedTask, err := processor.Process(ctx, current)

			if err != nil {
				fmt.Println(err)
			}

			processed <- resultedTask
			fmt.Println("Task", current.Name, "processed")
		}
	}
}

func StartRunnerPool(ctx context.Context, runnerCount int, tasks <-chan task.Task, processed chan<- task.Result, processor TaskProcessor, service TaskService) {
	fmt.Println("Starting runner pool, count is:", runnerCount)
	wg := &sync.WaitGroup{}
	wg.Add(runnerCount)

	for i := 0; i < runnerCount; i++ {
		go func() {
			defer wg.Done()
			StartRunner(ctx, tasks, processed, processor, service)
		}()
	}

	go func() {
		wg.Wait()
		close(processed)
	}()
}

func StartProgressReporter(ctx context.Context, service TaskService, stop <-chan struct{}, print func(stats task.TaskStats)) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	fmt.Println("Progress reporter started")
	for {
		select {
		case <-ticker.C:
			print(service.GetStats())
		case <-ctx.Done():
			return
		case <-stop:
			return
		}
	}
}

func Run(ctx context.Context, service TaskService, runnerCount int, processor TaskProcessor, print func(stats task.TaskStats)) {
	tasks := make(chan task.Task, runnerCount)
	processed := make(chan task.Result, runnerCount)
	collectorDone := make(chan struct{})
	stopProgress := make(chan struct{})

	queueCount := len(service.QueuedTasks())

	if queueCount <= 0 {
		fmt.Println(queueCount)
		return
	}

	go StartProgressReporter(ctx, service, stopProgress, print)

	go func() {
		ReadProcessed(processed, service)
		close(collectorDone)
	}()
	StartRunnerPool(ctx, runnerCount, tasks, processed, processor, service)

	PushTasksToQueue(ctx, tasks, service)
	<-collectorDone
	close(stopProgress)
	fmt.Println("run completed")
}
