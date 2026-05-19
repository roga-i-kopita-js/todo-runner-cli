package runner

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"todo-runner-cli/internal/task"
)

type FakeProcessor struct {
	calls atomic.Int32
}

func (p *FakeProcessor) Process(ctx context.Context, current task.Task) (task.Result, error) {
	p.calls.Add(1)
	resultedTask := task.Result{ID: current.ID}
	select {
	case <-ctx.Done():
		resultedTask.Status = task.Cancelled
		return resultedTask, errors.New("context canceled")
	default:
		if current.Name == "fail" {
			resultedTask.Error = errors.New("invalid task name")
			resultedTask.Status = task.Failed
		} else {
			resultedTask.Status = task.Done
		}
		return resultedTask, nil
	}
}

type FakePrinter struct {
}

func (f FakePrinter) PrintStats(_stats task.TaskStats) {
}

func TestRunner_Run(t *testing.T) {
	t.Run("Happy", func(t *testing.T) {
		ctx := context.Background()
		storage := task.NewInMemoryTaskStorage()
		service := task.NewTaskService(storage)

		_, err := service.Add(task.AddTaskInput{Name: "task1", Duration: 1})
		if err != nil {
			t.Fatal(err)
		}
		_, err1 := service.Add(task.AddTaskInput{Name: "task2", Duration: 1})
		if err1 != nil {
			t.Fatal(err1)
		}
		_, err2 := service.Add(task.AddTaskInput{Name: "task3", Duration: 1})
		if err2 != nil {
			t.Fatal(err2)
		}
		_, err3 := service.Add(task.AddTaskInput{Name: "fail", Duration: 1})
		if err3 != nil {
			t.Fatal(err3)
		}
		processor := &FakeProcessor{}
		runner := NewTaskRunner(service, processor, FakePrinter{})
		err4 := runner.Run(ctx, 3)
		if err4 != nil {
			t.Fatal(err3)
		}
		stats := service.GetStats()
		if processor.calls.Load() != 4 {
			t.Fatal("Invalid count of processor calls, got:", processor.calls.Load(), "expected:", 4)
		}
		if stats.Queued != 0 || stats.Done != 3 || stats.Failed != 1 || stats.Running != 0 || stats.Cancelled != 0 {
			t.Fatalf("Expected 0 queued, 3 done, and 1 failed tasks, got %d queued, %d done, and %d failed, but ", stats.Queued, stats.Done, stats.Failed)
		}
	})

	t.Run("Only queued task processed", func(t *testing.T) {
		ctx := context.Background()
		storage := task.NewInMemoryTaskStorage()
		service := task.NewTaskService(storage)

		_, err := service.Add(task.AddTaskInput{Name: "task1", Duration: 1})
		if err != nil {
			t.Fatal(err)
		}
		task2, err1 := service.Add(task.AddTaskInput{Name: "task2", Duration: 1})
		if err1 != nil {
			t.Fatal(err1)
		}
		task3, err2 := service.Add(task.AddTaskInput{Name: "task3", Duration: 1})
		if err2 != nil {
			t.Fatal(err2)
		}
		_, err3 := service.Add(task.AddTaskInput{Name: "fail", Duration: 1})
		if err3 != nil {
			t.Fatal(err3)
		}

		_, err4 := service.UpdateTask(task.Result{ID: task2.ID, Status: task.Cancelled})
		if err4 != nil {
			t.Fatal(err4)
		}

		_, err5 := service.UpdateTask(task.Result{ID: task3.ID, Status: task.Running})
		if err5 != nil {
			t.Fatal(err5)
		}

		_, err6 := service.UpdateTask(task.Result{ID: task3.ID, Status: task.Done})
		if err6 != nil {
			t.Fatal(err6)
		}

		processor := &FakeProcessor{}
		runner := NewTaskRunner(service, processor, FakePrinter{})
		err7 := runner.Run(ctx, 3)
		if err7 != nil {
			t.Fatal(err7)
		}

		if processor.calls.Load() != 2 {
			t.Fatal("Invalid count of processor calls, got:", processor.calls.Load(), "expected:", 2)
		}

		stats := service.GetStats()

		if stats.Queued != 0 || stats.Done != 2 || stats.Failed != 1 || stats.Running != 0 || stats.Cancelled != 1 {
			t.Fatalf("unexpected stats: %+v", stats)
		}
	})

	t.Run("Should not started if no queued tasks", func(t *testing.T) {
		ctx := context.Background()
		storage := task.NewInMemoryTaskStorage()
		service := task.NewTaskService(storage)

		processor := &FakeProcessor{}
		runner := NewTaskRunner(service, processor, FakePrinter{})
		err := runner.Run(ctx, 3)
		if err != nil {
			t.Fatal(err)
		}

		if processor.calls.Load() > 0 {
			t.Fatalf("processor should not be called if no tasks")
		}

		stats := service.GetStats()

		if stats.Queued != 0 || stats.Done != 0 || stats.Failed != 0 || stats.Running != 0 || stats.Cancelled != 0 {
			t.Fatalf("unexpected stats: %+v", stats)
		}
	})
}
