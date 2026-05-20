package runner

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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

type TaskRunner struct {
	service   TaskService
	processor TaskProcessor
	logger    *slog.Logger
}

var (
	ErrInvalidRunnersCount = errors.New("runner count must be positive")
)

func NewTaskRunner(service TaskService, processor TaskProcessor, logger *slog.Logger) *TaskRunner {
	if logger == nil {
		logger = slog.Default()
	}

	return &TaskRunner{
		service:   service,
		processor: processor,
		logger:    logger,
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
			r.logger.Error("runner readProcessed: update processed result", "result", result, "error", err)
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

func (r *TaskRunner) startRunner(ctx context.Context, tasks <-chan task.Task, processed chan<- task.Result, runnerID int) {
	r.logger.Debug("runner startRunner", "worker_id", runnerID)
	for {
		select {
		case <-ctx.Done():
			return
		case current, ok := <-tasks:
			if !ok {
				return
			}
			r.logger.Debug(
				"task received",
				"worker_id", runnerID,
				"task_id", current.ID,
				"task_name", current.Name,
				"duration_seconds", current.DurationInSeconds,
			)

			err := r.changeTaskToRunning(current)
			if err != nil {
				r.logger.Error("failed to change task status to running", "task", current, "error", err)
				continue
			}

			resultedTask, err := r.processor.Process(ctx, current)
			if err != nil {
				r.logger.Warn(
					"task processing failed",
					"worker_id", runnerID,
					"task", current,
					"error", err,
				)
			}

			processed <- resultedTask
			r.logger.Debug("task processed", "task", current)
		}
	}
}

func (r *TaskRunner) startRunnerPool(ctx context.Context, runnerCount int, tasks <-chan task.Task, processed chan<- task.Result) {
	r.logger.Info("worker pool started", "runner_count", runnerCount)
	wg := &sync.WaitGroup{}
	wg.Add(runnerCount)

	for i := 0; i < runnerCount; i++ {
		workerID := i + 1
		go func(workerID int) {
			defer wg.Done()
			r.startRunner(ctx, tasks, processed, workerID)
		}(workerID)
	}

	go func() {
		wg.Wait()
		close(processed)
	}()
}

func (r *TaskRunner) startProgressReporter(ctx context.Context, stop <-chan struct{}) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	r.logger.Info("runner reporter started")
	for {
		select {
		case <-ticker.C:
			r.logger.Info("statistics", "stats", r.service.GetStats())
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
		r.logger.Error("invalid runner count", "runner_count", runnerCount, "error", err)
		return err
	}
	r.logger.Info("run started", "runner_count", runnerCount)

	tasks := make(chan task.Task, runnerCount)
	processed := make(chan task.Result, runnerCount)
	collectorDone := make(chan struct{})
	stopProgress := make(chan struct{})

	queued, err := r.service.QueuedTasks()
	if err != nil {
		r.logger.Error("failed to get queued tasks", "error", err)
		return fmt.Errorf("runner Run: get queued tasks: %w", err)
	}
	queueCount := len(queued)
	r.logger.Info("queued tasks loaded", "queued_count", queueCount)
	if queueCount <= 0 {
		r.logger.Info("no queued tasks to run")
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
	r.logger.Info("run completed", "stats", r.service.GetStats())
	return nil
}
