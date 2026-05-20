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
	QueuedTasks() ([]task.Task, error)
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

func (r *TaskRunner) pushTasksToQueue(ctx context.Context, tasksChan chan<- task.Task, queued []task.Task) {
	defer close(tasksChan)

	for _, current := range queued {
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
			fmt.Println(fmt.Errorf("runner readProcessed: update processed result for task id %d: %w", result.ID, err))
			continue
		}
	}
}

func validateRunnersCount(count int) error {
	if count <= 0 {
		return fmt.Errorf("runner validateRunnersCount: invalid count %d: %w", count, ErrInvalidRunnersCount)
	}
	return nil
}

func (r *TaskRunner) changeTaskToRunning(current task.Task) error {
	result := task.Result{ID: current.ID, Status: task.Running}
	_, err := r.service.UpdateTask(result)
	if err != nil {
		return fmt.Errorf("runner changeTaskToRunning: task id %d: %w", current.ID, err)
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

			err := r.changeTaskToRunning(current)
			if err != nil {
				fmt.Println(err)
				continue
			}

			resultedTask, err := r.processor.Process(ctx, current)
			if err != nil {
				fmt.Println(fmt.Errorf("runner startRunner: process task id %d: %w", current.ID, err))
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

	queued, err := r.service.QueuedTasks()
	if err != nil {
		return fmt.Errorf("runner Run: get queued tasks: %w", err)
	}
	queueCount := len(queued)

	if queueCount <= 0 {
		return nil
	}

	go r.startProgressReporter(ctx, stopProgress)

	go func() {
		r.readProcessed(processed)
		close(collectorDone)
	}()
	r.startRunnerPool(ctx, runnerCount, tasks, processed)

	r.pushTasksToQueue(ctx, tasks, queued)
	<-collectorDone
	close(stopProgress)
	fmt.Println("run completed")
	r.printer.PrintStats(r.service.GetStats())
	return nil
}
