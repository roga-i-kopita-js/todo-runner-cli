package processor

import (
	"context"
	"errors"
	"testing"
	"todo-runner-cli/internal/task"
)

func TestSimpleTaskProcessor_Process(t *testing.T) {
	t.Run("Queued to done", func(t *testing.T) {
		ctx := context.Background()
		process := &SimpleTaskProcessor{}
		taskEl := task.Task{ID: 1, Status: task.Queued, Name: "task"}
		result, err := process.Process(ctx, taskEl)

		if err != nil {
			t.Fatal("Expected no error, got: ", err)
		}
		if result.ID != taskEl.ID || result.Error != nil || result.Status != task.Done {
			t.Fatal("invalid result, expected id", taskEl.ID, "error:", nil, "status:", task.Done, "got:", result)
		}
	})

	t.Run("Queued to Failed", func(t *testing.T) {
		ctx := context.Background()
		process := &SimpleTaskProcessor{}
		taskEl := task.Task{ID: 1, Status: task.Queued, Name: "fail"}
		result, err := process.Process(ctx, taskEl)

		if err != nil {
			t.Fatal("Expected no error, got", err)
		}
		if result.ID != taskEl.ID || !errors.Is(result.Error, ErrInvalidName) || result.Status != task.Failed {
			t.Fatal("invalid result, expected id", taskEl.ID, "error:", ErrInvalidName, "status:", task.Failed, "got:", result)
		}
	})

	t.Run("Cancellation test", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		process := &SimpleTaskProcessor{}
		taskEl := task.Task{ID: 1, Status: task.Queued, Name: "task", DurationInSeconds: 0}
		expected := task.Result{ID: taskEl.ID, Status: task.Cancelled}
		result, err := process.Process(ctx, taskEl)

		if result.ID != expected.ID {
			t.Fatal("Expected result ID to match expected ID, got:", result.ID, "expected:", expected.ID)
		}

		if result.Status != expected.Status {
			t.Fatal("Expected result status to match expected status, got:", result.Status, "expected:", expected.Status)
		}

		if !errors.Is(err, ErrClosed) {
			t.Fatal("Expected error:", ErrClosed, "got", err)
		}
	})
}
