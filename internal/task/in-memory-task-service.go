package task

import (
	"errors"
	"sync"
)

type Status string

const (
	Queued Status = "queued"
	Done   Status = "done"
	Failed Status = "failed"
)

type Task struct {
	ID                int
	Name              string
	DurationInSeconds int
	Status            Status
}

var (
	ErrInvalidQueuedTasks = errors.New("no queued tasks")
	ErrInvalidTaskID      = errors.New("invalid task id")
)

type TaskStats struct {
	Queued int
	Done   int
	Failed int
}

type AddTaskInput struct {
	Name     string
	Duration int
}

type InMemoryTaskService struct {
	tasks  []Task
	nextID int
	mutex  sync.RWMutex
}

func NewInMemoryTaskService() *InMemoryTaskService {
	return &InMemoryTaskService{nextID: 1}
}

type Result struct {
	ID    int
	Name  string
	Error error
}

func (t *InMemoryTaskService) Add(task AddTaskInput) Task {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	created := Task{ID: t.nextID, Name: task.Name, DurationInSeconds: task.Duration, Status: Queued}

	t.tasks = append(t.tasks, created)
	t.nextID++

	return created
}

func (t *InMemoryTaskService) Stats() TaskStats {
	queued := 0
	done := 0
	failed := 0

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	for _, item := range t.tasks {
		switch item.Status {
		case Queued:
			queued++
		case Done:
			done++
		case Failed:
			failed++
		}
	}

	return TaskStats{Queued: queued, Done: done, Failed: failed}
}

func (t *InMemoryTaskService) QueuedTasks() []Task {
	var result []Task

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	for _, task := range t.tasks {
		if task.Status == Queued {
			result = append(result, task)
		}
	}

	return result
}

func (t *InMemoryTaskService) MarkProcessed(taskResult Result) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for index, task := range t.tasks {
		if task.ID == taskResult.ID {
			if taskResult.Error != nil {
				t.tasks[index].Status = Failed
			} else {
				t.tasks[index].Status = Done
			}
			return nil
		}
	}

	return ErrInvalidTaskID
}

func (t *InMemoryTaskService) ValidateQueuedTaskCount() error {
	count := 0

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	for _, task := range t.tasks {
		if task.Status == Queued {
			count++
		}
	}
	if count == 0 {
		return ErrInvalidQueuedTasks
	}
	return nil
}
