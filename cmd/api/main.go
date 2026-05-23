package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"todo-runner-cli/internal/baseapi"
	"todo-runner-cli/internal/config"
	"todo-runner-cli/internal/httpresponse"
	"todo-runner-cli/internal/logger"
	"todo-runner-cli/internal/processor"
	"todo-runner-cli/internal/runner"
	"todo-runner-cli/internal/task"
	"todo-runner-cli/internal/taskapi"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load config:", err)
		os.Exit(1)
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
	log := logger.NewLogger(cfg.LogLevel, cfg.LogFormat)
	worker := runner.NewTaskRunner(service, taskProcessor, log)
	jsonWriter := httpresponse.NewJSONWriter(log)
	baseHandler := baseapi.NewBaseHandler(log, jsonWriter)
	baseHandler.RegisterRoutes(mux)

	apiHandler := taskapi.NewTaskHandler(ctx, service, log, worker, taskapi.Config{WorkersCount: cfg.RunnerCount}, jsonWriter)
	apiHandler.RegisterRoutes(mux)

	go func() {
		log.Info(fmt.Sprintf("Listening on %s", server.Addr))

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", "error", err.Error())
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	sig := <-signals
	log.Info("shutdown signal received:", "signal:", sig)
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	shutdownErr := server.Shutdown(shutdownCtx)
	if shutdownErr != nil {
		log.Error("shutdown failed", "error", shutdownErr)
	}
	log.Info("shutdown complete, server stopped")
}
