package task

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidTransaction = errors.New("invalid transaction")
)

func NewTaskService(storage TaskStorage) *TaskServiceStruct {
	return &TaskServiceStruct{
		storage: storage,
	}
}

func (s *TaskServiceStruct) Add(task AddTaskInput) (Task, error) {
	result, err := s.storage.Add(task)
	if err != nil {
		return result, fmt.Errorf("service Add: invalid input %+v: %w", task, err)
	}
	return result, nil
}

func (s *TaskServiceStruct) GetByID(taskID int) (Task, error) {
	result, err := s.storage.GetByID(taskID)
	if err != nil {
		return result, fmt.Errorf("service GetByID: task id %d: %w", taskID, err)
	}
	return result, nil
}

func (s *TaskServiceStruct) GetList(filter GetListFilterParams) ([]Task, error) {
	result, err := s.storage.GetList(filter)
	if err != nil {
		return nil, fmt.Errorf("service GetList: filter %+v: %w", filter, err)
	}

	return result, nil
}

func validateTransactionStatus(task Task, status Status) error {
	if task.Status == Queued {
		if status != Running && status != Cancelled {
			return ErrInvalidTransaction
		}
	}

	if task.Status == Running {
		if status != Done && status != Failed && status != Cancelled {
			return ErrInvalidTransaction
		}
	}

	if task.Status == Done || task.Status == Failed || task.Status == Cancelled {
		return ErrInvalidTransaction
	}

	return nil
}

func (s *TaskServiceStruct) UpdateTask(taskResult Result) (Task, error) {
	task, err := s.storage.GetByID(taskResult.ID)

	if err != nil {
		return task, fmt.Errorf("service UpdateTask: get current task by id %d: %w", taskResult.ID, err)
	}

	err = ValidateStatus(taskResult.Status)
	if err != nil {
		return task, fmt.Errorf("service UpdateTask: validate status %q: %w", taskResult.Status, err)
	}

	err = validateTransactionStatus(task, taskResult.Status)
	if err != nil {
		return task, fmt.Errorf(
			"service UpdateTask: invalid transition from %q to %q: %w",
			task.Status,
			taskResult.Status,
			err,
		)
	}

	updatedTask, err := s.storage.UpdateTask(taskResult)
	if err != nil {
		return task, fmt.Errorf("service UpdateTask: save result for task id %d: %w", taskResult.ID, err)
	}

	return updatedTask, nil
}

func (s *TaskServiceStruct) GetStats() TaskStats {
	return s.storage.GetStats()
}

func (s *TaskServiceStruct) QueuedTasks() ([]Task, error) {
	var params GetListFilterParams
	params.Statuses = []Status{Queued}
	tasks, err := s.storage.GetList(params)

	if err != nil {
		return nil, fmt.Errorf("service QueuedTasks: get queued tasks: %w", err)
	}

	return tasks, nil
}
