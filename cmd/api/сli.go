package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"todo-runner-cli/internal/command"
	"todo-runner-cli/internal/config"
	"todo-runner-cli/internal/logger"
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
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to load config:", err)
		os.Exit(1)
	}
	fmt.Println("Type a command or 'exit' to quit:")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scanner := bufio.NewScanner(os.Stdin)
	log := logger.NewLogger(cfg.LogLevel, cfg.LogFormat)
	storage := task.NewInMemoryTaskStorage()
	taskService := task.NewTaskService(storage)
	taskProcessor := processor.SimpleTaskProcessor{}
	printer := Printer{}
	taskRunner := runner.NewTaskRunner(taskService, taskProcessor, log)
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
					log.Error("failed to add task", "task_input", input, "error", err)
					fmt.Fprintln(os.Stderr, "Failed to add task:", err)
					return
				}

				fmt.Println("queued ID:", created.ID)
			},
			Run: func() {
				err := taskRunner.Run(ctx, cfg.RunnerCount)
				if err != nil {
					log.Error("failed to run tasks", "error", err)
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
			log.Warn("invalid command input", "input", text, "error", err)
			fmt.Fprintln(os.Stderr, "Invalid command:", err)
			continue
		}

		shouldCancel := commandModule.Execute()

		if shouldCancel {
			fmt.Println("..exiting")
			return
		}
	}
}
