package task

import (
	"sync"
	"time"
)

type Status string

type Task struct {
	ID                int
	Name              string
	DurationInSeconds int
	Status            Status
	Error             string
	StartedAt         *time.Time
	CreatedAt         time.Time
	FinishedAt        *time.Time
}

type TaskStats struct {
	Queued    int
	Done      int
	Failed    int
	Running   int
	Cancelled int
}

type AddTaskInput struct {
	Name     string
	Duration int
}

type Storage struct {
	tasks  map[int]Task
	nextID int
	mutex  sync.RWMutex
}

type GetListFilterParams struct {
	IDs      []int
	Statuses []Status
}

type Result struct {
	ID     int
	Name   string
	Error  error
	Status Status
}

type TaskServiceStruct struct {
	storage TaskStorage
}

type TaskStorage interface {
	Add(task AddTaskInput) (Task, error)
	GetByID(taskID int) (Task, error)
	GetList(filter GetListFilterParams) ([]Task, error)
	UpdateTask(taskResult Result) (Task, error)
	GetStats() TaskStats
}

type TaskService interface {
	Add(task AddTaskInput) (Task, error)
	GetByID(taskID int) (Task, error)
	GetList(filter GetListFilterParams) ([]Task, error)
	UpdateTask(taskResult Result) (Task, error)
	GetStats() TaskStats
	QueuedTasks() []Task
}
