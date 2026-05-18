package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"todo-runner-cli/internal/command"
	"todo-runner-cli/internal/processor"
	"todo-runner-cli/internal/runner"
	"todo-runner-cli/internal/task"
)

func PrintStatsInfo(stats task.TaskStats) {
	fmt.Println("Queued:", stats.Queued, "Done:", stats.Done, "Failed:", stats.Failed, "Running:", stats.Running, "Cancelled:", stats.Cancelled)
}

func main() {
	fmt.Println("Type a command or 'exit' to quit:")
	ctx, cancel := context.WithCancel(context.Background())
	runnerCount := 3

	scanner := bufio.NewScanner(os.Stdin)

	storage := task.NewInMemoryTaskStorage()
	taskService := task.NewTaskService(storage)
	taskProcessor := processor.SimpleTaskProcessor{}

	for scanner.Scan() {
		text := scanner.Text()
		commandModule, err := command.NewCommand(text, command.Actions{
			Add: func(newTask command.ParsedCommand) {
				created, err := taskService.Add(task.AddTaskInput{
					Name:     newTask.Name,
					Duration: newTask.Duration,
				})

				if err != nil {
					fmt.Println("Err:", err.Error())
					return
				}

				fmt.Println("queued ID:", created.ID)
			},
			Run: func() {
				runner.Run(ctx, taskService, runnerCount, taskProcessor, PrintStatsInfo)
			},
			Stats: func() {
				PrintStatsInfo(taskService.GetStats())
			},
		})

		if err != nil {
			fmt.Println(err)
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
