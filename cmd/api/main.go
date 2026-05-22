package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"todo-runner-cli/internal/config"
	"todo-runner-cli/internal/processor"
	"todo-runner-cli/internal/runner"
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

type StatusResponse struct {
	Status string `json:"status"`
}

func ValidateHeaders(header http.Header) error {
	var headerErrors error
	contentType := strings.ToLower(strings.TrimSpace(header.Get("Content-Type")))
	if !strings.HasPrefix(contentType, "application/json") {
		headerErrors = errors.New("Content-Type header is not application/json")
	}

	return headerErrors
}

func WriteJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		fmt.Println("failed to write json response:", err)
	}
}

func NewLogger(level string, format string) *slog.Logger {
	var lvl slog.Level

	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "info":
		lvl = slog.LevelInfo
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	var handler slog.Handler
	switch format {
	case "text":
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: lvl,
		})
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: lvl,
		})
	default:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: lvl,
		})
	}

	return slog.New(handler)
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("failed to load config:", err)
		return
	}
	mux := http.NewServeMux()
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	ctx, cancel := context.WithCancel(context.Background())

	storage := task.NewInMemoryTaskStorage()
	service := task.NewTaskService(storage)
	taskProcessor := processor.SimpleTaskProcessor{}
	logger := NewLogger(cfg.LogLevel, cfg.LogFormat)
	worker := runner.NewTaskRunner(service, taskProcessor, logger)

	mux.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			WriteJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method: " + r.Method + " is not allowed"})
			return
		}

		headerErr := ValidateHeaders(r.Header)
		if headerErr != nil {
			WriteJSON(w, http.StatusNotAcceptable, ErrorResponse{Error: headerErr.Error()})
			return
		}

		var payload AddTaskRequest
		if decodeErr := json.NewDecoder(r.Body).Decode(&payload); decodeErr != nil {
			WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: decodeErr.Error()})
			return
		}

		result, addErr := service.Add(task.AddTaskInput{Name: payload.Name, Duration: payload.DurationInSeconds})
		if addErr != nil {
			WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: addErr.Error()})
			return
		}

		WriteJSON(w, http.StatusCreated, AddTaskResponse{ID: result.ID, Name: result.Name, Status: string(result.Status), DurationInSeconds: result.DurationInSeconds})
	})

	mux.HandleFunc("/tasks/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			WriteJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method: " + r.Method + " is not allowed"})
			return
		}

		result := service.GetStats()

		WriteJSON(w, http.StatusOK, TaskStats{Queued: result.Queued, Running: result.Running, Done: result.Done, Failed: result.Failed, Canceled: result.Cancelled})
		return
	})

	mux.HandleFunc("/tasks/run", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			WriteJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method: " + r.Method + " is not allowed"})
			return
		}

		go func() {
			runErr := worker.Run(ctx, cfg.RunnerCount)
			if runErr != nil {
				logger.Error("failed to run tasks", "error", runErr)
				return
			}
		}()

		WriteJSON(w, http.StatusAccepted, StatusResponse{Status: "started"})
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			WriteJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method: " + r.Method + " is not allowed"})
			return
		}
		WriteJSON(w, http.StatusOK, StatusResponse{Status: "ok"})
	})

	go func() {
		logger.Info(fmt.Sprintf("Listening on %s", server.Addr))

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err.Error())
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	sig := <-signals
	logger.Info("shutdown signal received:", "signal:", sig)
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	shutdownErr := server.Shutdown(shutdownCtx)
	if shutdownErr != nil {
		logger.Error("shutdown failed", "error", shutdownErr)
	}
	logger.Info("shutdown complete, server stopped")
}
