package main

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"todo-runner-cli/internal/command"
	"todo-runner-cli/internal/processor"
	"todo-runner-cli/internal/runner"
	"todo-runner-cli/internal/task"
)

type Printer struct {
}

func (f Printer) PrintStats(stats task.TaskStats) {
	fmt.Println("Queued:", stats.Queued, "Done:", stats.Done, "Failed:", stats.Failed, "Running:", stats.Running, "Cancelled:", stats.Cancelled)
}

func main() {
	fmt.Println("Type a command or 'exit' to quit:")
	ctx, cancel := context.WithCancel(context.Background())
	runnerCount := 3

	scanner := bufio.NewScanner(os.Stdin)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	storage := task.NewInMemoryTaskStorage()
	taskService := task.NewTaskService(storage)
	taskProcessor := processor.SimpleTaskProcessor{}
	printer := Printer{}
	taskRunner := runner.NewTaskRunner(taskService, taskProcessor, logger)
	for scanner.Scan() {
		text := scanner.Text()
		commandModule, err := command.NewCommand(text, command.Actions{
			Add: func(newTask command.ParsedCommand) {
				input := task.AddTaskInput{
					Name:     newTask.Name,
					Duration: newTask.Duration,
				}
				created, err := taskService.Add(input)

				if err != nil {
					logger.Error("failed to add task", "task_input", input, "error", err)
					fmt.Fprintln(os.Stderr, "Failed to add task:", err)
					return
				}

				fmt.Println("queued ID:", created.ID)
			},
			Run: func() {
				err := taskRunner.Run(ctx, runnerCount)
				if err != nil {
					logger.Error("failed to run tasks", "error", err)
					fmt.Fprintln(os.Stderr, "Failed to run tasks:", err)
					return
				}

				printer.PrintStats(taskService.GetStats())
			},
			Stats: func() {
				printer.PrintStats(taskService.GetStats())
			},
		})

		if err != nil {
			logger.Warn("invalid command input", "input", text, "error", err)
			fmt.Fprintln(os.Stderr, "Invalid command:", err)
			continue
		}

		shouldCancel := commandModule.Execute()

		if shouldCancel {
			fmt.Println("..exiting")
			cancel()
			return
		}
	}
}
