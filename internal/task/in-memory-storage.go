package task

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrTaskNotFound        = errors.New("task not found")
	ErrInvalidStatus       = errors.New("invalid status, must be one of: 'queued' 'running' 'done' 'failed' 'cancelled'")
	ErrInvalidTaskName     = errors.New("invalid name")
	ErrInvalidTaskDuration = errors.New("invalid Duration in seconds, must be grater then 0")
	ErrInvalidTaskID       = errors.New("invalid task ID, must be grater then 0")
)

const (
	Queued    Status = "queued"
	Running   Status = "running"
	Done      Status = "done"
	Failed    Status = "failed"
	Cancelled Status = "cancelled"
)

func NewInMemoryTaskStorage() *Storage {
	return &Storage{
		nextID: 1,
		tasks:  make(map[int]Task),
	}
}

func validateTaskInput(input AddTaskInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("in-memory storage invalid task name: %w, got: %q", ErrInvalidTaskName, input.Name)
	}

	if input.Duration <= 0 {
		return fmt.Errorf("in-memory storage invalid task duration: %w, got: %d", ErrInvalidTaskDuration, input.Duration)
	}

	return nil
}

func (t *Storage) Add(task AddTaskInput) (Task, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	err := validateTaskInput(task)

	if err != nil {
		return Task{}, fmt.Errorf("in-memory storage Add, invalid task input: %w", err)
	}

	created := Task{ID: t.nextID, Name: task.Name, DurationInSeconds: task.Duration, Status: Queued, CreatedAt: time.Now()}

	t.tasks[t.nextID] = created

	t.nextID++

	return created, nil
}

func (t *Storage) GetByID(id int) (Task, error) {
	err := ValidateId(id)
	if err != nil {
		return Task{}, fmt.Errorf("in-memory storage GetByID, invalid task id %d, err : %w", id, err)
	}

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	task, ok := t.tasks[id]

	if !ok {
		return task, fmt.Errorf("in-memory storage GetByID, task with id: %d not found: %w", id, ErrTaskNotFound)
	}

	return task, nil
}

func ValidateStatus(status Status) error {
	if status != Queued && status != Running && status != Failed && status != Cancelled && status != Done {
		return ErrInvalidStatus
	}

	return nil
}

func ValidateId(id int) error {
	if id <= 0 {
		return ErrInvalidTaskID
	}
	return nil
}

func validateIds(ids []int) error {
	for _, id := range ids {
		err := ValidateId(id)
		if err != nil {
			return fmt.Errorf("in-memory storage, invalid task ID: %d: %w", id, err)
		}
	}

	return nil
}

func validateStatuses(statuses []Status) error {
	var err error

	for _, status := range statuses {
		statusErr := ValidateStatus(status)
		if statusErr != nil {
			err = fmt.Errorf("in-memory storage, invalid task status %q: %w", status, statusErr)
			break
		}
	}

	return err
}

func getByIdsAndFilters(tasks map[int]Task, ids []int, statuses []Status) []Task {
	var result []Task

	for _, id := range ids {
		for _, status := range statuses {
			task, ok := tasks[id]
			if ok && task.Status == status {
				result = append(result, task)
			}
		}
	}

	return result
}

func getByStatuses(tasks map[int]Task, statuses []Status) []Task {
	var result []Task

	for _, task := range tasks {
		for _, status := range statuses {
			if task.Status == status {
				result = append(result, task)
			}
		}
	}

	return result
}

func getByIds(tasks map[int]Task, ids []int) []Task {
	var result []Task

	for _, task := range tasks {
		for _, id := range ids {
			if task.ID == id {
				result = append(result, task)
			}
		}
	}

	return result
}

func getSlice(tasks map[int]Task) []Task {
	var result []Task
	for _, task := range tasks {
		result = append(result, task)
	}

	return result
}

func (t *Storage) GetList(filter GetListFilterParams) ([]Task, error) {
	// валидируем переданные статусы
	if len(filter.Statuses) > 0 {
		err := validateStatuses(filter.Statuses)
		if err != nil {
			return []Task{}, fmt.Errorf("in-memory storage GetList: %w", err)
		}
	}

	// валидируем переданные айдишникиь
	if len(filter.IDs) > 0 {
		err := validateIds(filter.IDs)
		if err != nil {
			return []Task{}, fmt.Errorf("in-memory storage GetList: %w", err)
		}
	}

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	// если параметров нет, то собираем слайс из всех элементов в мап
	if len(filter.IDs) == 0 && len(filter.Statuses) == 0 {
		return getSlice(t.tasks), nil
	}

	// кейс 1: переданы оба параметра
	if len(filter.IDs) > 0 && len(filter.Statuses) > 0 {
		return getByIdsAndFilters(t.tasks, filter.IDs, filter.Statuses), nil
	}

	// кейс 2: переданы только статусы
	if len(filter.IDs) == 0 && len(filter.Statuses) > 0 {
		return getByStatuses(t.tasks, filter.Statuses), nil
	}

	// кейс 3: переданы только айдишники
	if len(filter.IDs) > 0 && len(filter.Statuses) == 0 {
		return getByIds(t.tasks, filter.IDs), nil
	}

	return []Task{}, nil
}

func (t *Storage) GetStats() TaskStats {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	queued := 0
	done := 0
	failed := 0
	running := 0
	cancelled := 0

	for _, item := range t.tasks {
		switch item.Status {
		case Queued:
			queued++
		case Done:
			done++
		case Failed:
			failed++
		case Running:
			running++
		case Cancelled:
			cancelled++

		}
	}

	return TaskStats{Queued: queued, Done: done, Failed: failed, Cancelled: cancelled, Running: running}
}

func (t *Storage) UpdateTask(taskResult Result) (Task, error) {
	idErr := ValidateId(taskResult.ID)
	if idErr != nil {
		return Task{}, fmt.Errorf("in-memory storage UpdateTask: %w, Id: %d is incorrect", idErr, taskResult.ID)
	}

	statusErr := ValidateStatus(taskResult.Status)
	if statusErr != nil {
		return Task{}, fmt.Errorf("in-memory storage: invalid status: %w, got: %q", statusErr, taskResult.Status)
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	task, ok := t.tasks[taskResult.ID]
	if !ok {
		return Task{}, fmt.Errorf("in-memory storage: task with id: %d, not found: %w", taskResult.ID, ErrTaskNotFound)
	}

	task.Status = taskResult.Status

	if taskResult.Error != nil {
		task.Error = taskResult.Error.Error()
	}

	if taskResult.Status == Running {
		now := time.Now()
		task.StartedAt = &now
	}

	if taskResult.Status == Done || taskResult.Status == Failed || taskResult.Status == Cancelled {
		now := time.Now()
		task.FinishedAt = &now
	}

	t.tasks[taskResult.ID] = task

	return task, nil
}
