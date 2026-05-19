package runner

import (
	"context"
	"errors"
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

type Printer interface {
	PrintStats(stats task.TaskStats)
}

type TaskRunner struct {
	service   TaskService
	processor TaskProcessor
	printer   Printer
}

var (
	ErrInvalidRunnersCount = errors.New("runner count must be positive")
)

func NewTaskRunner(service TaskService, processor TaskProcessor, printer Printer) *TaskRunner {
	return &TaskRunner{
		service:   service,
		processor: processor,
		printer:   printer,
	}
}

func (r *TaskRunner) pushTasksToQueue(ctx context.Context, tasksChan chan<- task.Task) {
	defer close(tasksChan)

	for _, current := range r.service.QueuedTasks() {
		select {
		case tasksChan <- current:
			continue
		case <-ctx.Done():
			return
		}
	}
}

func (r *TaskRunner) readProcessed(processed <-chan task.Result) {
	for result := range processed {
		_, err := r.service.UpdateTask(result)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}

func validateRunnersCount(count int) error {
	if count <= 0 {
		return ErrInvalidRunnersCount
	}
	return nil
}

func (r *TaskRunner) startRunner(ctx context.Context, tasks <-chan task.Task, processed chan<- task.Result) {
	fmt.Println("Runner started")
	for {
		select {
		case <-ctx.Done():
			return
		case current, ok := <-tasks:
			if !ok {
				return
			}
			result := task.Result{ID: current.ID, Status: task.Running}
			_, err := r.service.UpdateTask(result)

			if err != nil {
				continue
			}

			resultedTask, err := r.processor.Process(ctx, current)

			if err != nil {
				fmt.Println(err)
			}

			processed <- resultedTask
			fmt.Println("Task", current.Name, "processed")
		}
	}
}

func (r *TaskRunner) startRunnerPool(ctx context.Context, runnerCount int, tasks <-chan task.Task, processed chan<- task.Result) {
	fmt.Println("Starting runner pool, count is:", runnerCount)
	wg := &sync.WaitGroup{}
	wg.Add(runnerCount)

	for i := 0; i < runnerCount; i++ {
		go func() {
			defer wg.Done()
			r.startRunner(ctx, tasks, processed)
		}()
	}

	go func() {
		wg.Wait()
		close(processed)
	}()
}

func (r *TaskRunner) startProgressReporter(ctx context.Context, stop <-chan struct{}) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	fmt.Println("Progress reporter started")
	for {
		select {
		case <-ticker.C:
			r.printer.PrintStats(r.service.GetStats())
		case <-ctx.Done():
			return
		case <-stop:
			return
		}
	}
}

func (r *TaskRunner) Run(ctx context.Context, runnerCount int) error {
	err := validateRunnersCount(runnerCount)
	if err != nil {
		return err
	}

	tasks := make(chan task.Task, runnerCount)
	processed := make(chan task.Result, runnerCount)
	collectorDone := make(chan struct{})
	stopProgress := make(chan struct{})

	queueCount := len(r.service.QueuedTasks())

	if queueCount <= 0 {
		fmt.Println(queueCount)
		return nil
	}

	go r.startProgressReporter(ctx, stopProgress)

	go func() {
		r.readProcessed(processed)
		close(collectorDone)
	}()
	r.startRunnerPool(ctx, runnerCount, tasks, processed)

	r.pushTasksToQueue(ctx, tasks)
	<-collectorDone
	close(stopProgress)
	fmt.Println("run completed")
	r.printer.PrintStats(r.service.GetStats())
	return nil
}
