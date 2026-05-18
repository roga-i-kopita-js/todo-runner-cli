package task

import "errors"

var (
	ErrInvalidTransaction = errors.New("Invalid transaction")
)

func NewTaskService(storage TaskStorage) *TaskServiceStruct {
	return &TaskServiceStruct{
		storage: storage,
	}
}

func (s *TaskServiceStruct) Add(task AddTaskInput) (Task, error) {
	return s.storage.Add(task)
}

func (s *TaskServiceStruct) GetByID(taskID int) (Task, error) {
	return s.storage.GetByID(taskID)
}

func (s *TaskServiceStruct) GetList(filter GetListFilterParams) ([]Task, error) {
	return s.storage.GetList(filter)
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
		return task, err
	}

	err = ValidateStatus(taskResult.Status)
	if err != nil {
		return task, err
	}

	err = validateTransactionStatus(task, taskResult.Status)

	if err != nil {
		return task, err
	}

	return s.storage.UpdateTask(taskResult)
}

func (s *TaskServiceStruct) GetStats() TaskStats {
	return s.storage.GetStats()
}

func (s *TaskServiceStruct) QueuedTasks() []Task {
	var params GetListFilterParams
	params.Statuses = []Status{Queued}
	tasks, _ := s.storage.GetList(params)

	return tasks
}
