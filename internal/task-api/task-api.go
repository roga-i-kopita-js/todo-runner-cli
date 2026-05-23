package task_api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"todo-runner-cli/internal/task"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type AddTaskRequest struct {
	Name              string `json:"name"`
	DurationInSeconds int    `json:"duration_seconds"`
}

type AddTaskResponse struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	DurationInSeconds int    `json:"duration_seconds"`
	Status            string `json:"status"`
}

type TaskStats struct {
	Queued   int `json:"queued"`
	Done     int `json:"done"`
	Failed   int `json:"failed"`
	Running  int `json:"running"`
	Canceled int `json:"canceled"`
}

type TaskService interface {
	Add(task task.AddTaskInput) (task.Task, error)
	GetStats() task.TaskStats
}

type TaskWorker interface {
	Run(ctx context.Context, count int) error
}

type Config struct {
	WorkersCount int
}

type TaskHandler struct {
	service TaskService
	logger  *slog.Logger
	worker  TaskWorker
	ctx     context.Context
	cfg     Config
}

func NewTaskHandler(ctx context.Context, service TaskService, logger *slog.Logger, worker TaskWorker, cfg Config) *TaskHandler {
	return &TaskHandler{
		service: service,
		ctx:     ctx,
		cfg:     cfg,
		logger:  logger,
		worker:  worker,
	}
}

type StatusResponse struct {
	Status string `json:"status"`
}

func validateHeaders(header http.Header) error {
	var headerErrors error
	contentType := strings.ToLower(strings.TrimSpace(header.Get("Content-Type")))
	if !strings.HasPrefix(contentType, "application/json") {
		headerErrors = errors.New("Content-Type header is not application/json")
	}

	return headerErrors
}

func (t *TaskHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/tasks", t.addHandler)
	mux.HandleFunc("/tasks/stats", t.statsHandler)
	mux.HandleFunc("/tasks/run", t.runHandler)
}

func (t *TaskHandler) writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		t.logger.Error("failed to write json response", "error", err)
	}
}

func (t *TaskHandler) addHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		t.writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method: " + r.Method + " is not allowed"})
		return
	}

	headerErr := validateHeaders(r.Header)
	if headerErr != nil {
		t.writeJSON(w, http.StatusUnsupportedMediaType, ErrorResponse{Error: headerErr.Error()})
		return
	}

	var payload AddTaskRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&payload); decodeErr != nil {
		t.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: decodeErr.Error()})
		return
	}

	result, addErr := t.service.Add(task.AddTaskInput{Name: payload.Name, Duration: payload.DurationInSeconds})
	if addErr != nil {
		t.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: addErr.Error()})
		return
	}

	t.writeJSON(w, http.StatusCreated, AddTaskResponse{ID: result.ID, Name: result.Name, Status: string(result.Status), DurationInSeconds: result.DurationInSeconds})

}

func (t *TaskHandler) statsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		t.writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method: " + r.Method + " is not allowed"})
		return
	}

	result := t.service.GetStats()

	t.writeJSON(w, http.StatusOK, TaskStats{Queued: result.Queued, Running: result.Running, Done: result.Done, Failed: result.Failed, Canceled: result.Cancelled})
	return
}

func (t *TaskHandler) runHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		t.writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method: " + r.Method + " is not allowed"})
		return
	}

	go func() {
		runErr := t.worker.Run(t.ctx, t.cfg.WorkersCount)
		if runErr != nil {
			t.logger.Error("failed to run tasks", "error", runErr)
			return
		}
	}()

	t.writeJSON(w, http.StatusAccepted, StatusResponse{Status: "started"})
}
