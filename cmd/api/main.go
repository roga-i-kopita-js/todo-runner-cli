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

type TaskAddService interface {
	Add(task task.AddTaskInput) task.Task
}

func Add(taskService TaskAddService, newTask command.ParsedCommand) {
	created := taskService.Add(task.AddTaskInput{
		Name:     newTask.Name,
		Duration: newTask.Duration,
	})
	fmt.Println("queued:", created.ID)
}

func PrintStatsInfo(stats task.TaskStats) {
	fmt.Println("Queued:", stats.Queued, "Done:", stats.Done, "Failed:", stats.Failed)
}

func main() {
	fmt.Println("Type a command or 'exit' to quit:")
	ctx, cancel := context.WithCancel(context.Background())
	runnerCount := 3

	scanner := bufio.NewScanner(os.Stdin)
	taskService := task.NewInMemoryTaskService()
	taskProcessor := processor.SimpleTaskProcessor{}

	for scanner.Scan() {
		text := scanner.Text()
		commandModule, err := command.NewCommand(text, command.Actions{
			Add: func(newTask command.ParsedCommand) {
				Add(taskService, newTask)
			},
			Run: func() {
				runner.Run(ctx, taskService, runnerCount, taskProcessor, PrintStatsInfo)
			},
			Stats: func() {
				PrintStatsInfo(taskService.Stats())
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
