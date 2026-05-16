package processor

import (
	"context"
	"errors"
	"time"
	"todo-runner-cli/internal/task"
)

var (
	ErrInvalidName = errors.New("invalid name")
	ErrClosed      = errors.New("connection closed")
)

type SimpleTaskProcessor struct{}

func (p SimpleTaskProcessor) Process(ctx context.Context, current task.Task) (task.Result, error) {
	resultedTask := task.Result{ID: current.ID, Name: current.Name}
	select {
	case <-ctx.Done():
		return resultedTask, ErrClosed
	case <-time.After(time.Duration(current.DurationInSeconds) * time.Second):
		if current.Name == "fail" {
			resultedTask.Error = ErrInvalidName
		}

		return resultedTask, nil
	}

}
