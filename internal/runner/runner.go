package runner

import (
	"context"
	"fmt"
	"sync"
	"time"
	"todo-runner-cli/internal/task"
)

type TaskRunnerService interface {
	ValidateQueuedTaskCount() error
	QueuedTasks() []task.Task
	Stats() task.TaskStats
	MarkProcessed(result task.Result) error
}

type TaskProcessor interface {
	Process(ctx context.Context, task task.Task) (task.Result, error)
}

func PushTasksToQueue(ctx context.Context, tasksChan chan<- task.Task, service TaskRunnerService) {
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

func ReadProcessed(ctx context.Context, processed <-chan task.Result, service TaskRunnerService) {
	for {
		select {
		case result, ok := <-processed:
			if !ok {
				return
			}
			err := service.MarkProcessed(result)
			if err != nil {
				fmt.Println(err)
				continue
			}
		case <-ctx.Done():
			return
		}
	}
}

func StartRunner(ctx context.Context, tasks <-chan task.Task, processed chan<- task.Result, processor TaskProcessor) {
	fmt.Println("Runner started")
	for {
		select {
		case <-ctx.Done():
			return
		case current, ok := <-tasks:
			if !ok {
				return
			}
			resultedTask, err := processor.Process(ctx, current)
			if err != nil {
				return
			}
			select {
			case <-ctx.Done():
				return
			case processed <- resultedTask:
				fmt.Println("Task", current.Name, "processed")
			}
		}
	}
}

func StartRunnerPool(ctx context.Context, runnerCount int, tasks <-chan task.Task, processed chan<- task.Result, processor TaskProcessor) {
	fmt.Println("Starting runner pool, count is:", runnerCount)
	wg := &sync.WaitGroup{}
	wg.Add(runnerCount)

	for i := 0; i < runnerCount; i++ {
		go func() {
			defer wg.Done()
			StartRunner(ctx, tasks, processed, processor)
		}()
	}

	go func() {
		wg.Wait()
		close(processed)
	}()
}

func StartProgressReporter(ctx context.Context, service TaskRunnerService, stop <-chan struct{}, print func(stats task.TaskStats)) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	fmt.Println("Progress reporter started")
	for {
		select {
		case <-ticker.C:
			print(service.Stats())
		case <-ctx.Done():
			return
		case <-stop:
			return
		}
	}
}

func Run(ctx context.Context, service TaskRunnerService, runnerCount int, processor TaskProcessor, print func(stats task.TaskStats)) {
	tasks := make(chan task.Task, runnerCount)
	processed := make(chan task.Result, runnerCount)
	collectorDone := make(chan struct{})
	stopProgress := make(chan struct{})

	queueCount := service.ValidateQueuedTaskCount()
	if queueCount != nil {
		fmt.Println(queueCount)
		return
	}

	go StartProgressReporter(ctx, service, stopProgress, print)

	go func() {
		ReadProcessed(ctx, processed, service)
		close(collectorDone)
	}()
	StartRunnerPool(ctx, runnerCount, tasks, processed, processor)

	PushTasksToQueue(ctx, tasks, service)
	<-collectorDone
	close(stopProgress)
	fmt.Println("run completed")
}
