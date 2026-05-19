package command

import (
	"errors"
	"testing"
)

type FakeActions struct {
	AddCommand ParsedCommand
	AddCount   int
	RunCount   int
	StatsCount int
}

func (fa *FakeActions) Add(command ParsedCommand) {
	fa.AddCount++
	fa.AddCommand = command
}
func (fa *FakeActions) Run() {
	fa.RunCount++
}
func (fa *FakeActions) Stats() {
	fa.StatsCount++
}

func TestNewCommand(t *testing.T) {
	fakeActions := &FakeActions{}

	tests := []struct {
		Name     string
		expected error
		Value    string
	}{
		{
			Name:     "Valid command add",
			expected: nil,
			Value:    "add task 1",
		},
		{
			Name:     "invalid command add",
			expected: ErrInvalidAddTask,
			Value:    "add d",
		},
		{
			Name:     "invalid command add",
			expected: ErrInvalidTaskDuration,
			Value:    "add d asd",
		},
		{
			Name:     "invalid command add",
			expected: ErrInvalidTaskDurationValue,
			Value:    "add d -5",
		},
		{
			Name:     "Invalid command run",
			expected: ErrUnknownArguments,
			Value:    "run kek",
		},
		{
			Name:     "Valid command run",
			expected: nil,
			Value:    "run",
		},
		{
			Name:     "Valid command exit",
			expected: nil,
			Value:    "exit",
		},
		{
			Name:     "Valid command stats",
			expected: nil,
			Value:    "stats",
		},
		{
			Name:     "invalid command stats",
			expected: ErrUnknownArguments,
			Value:    "stats asd",
		},
		{
			Name:     "invalid command",
			expected: ErrInvalidTaskCommand,
			Value:    "asd",
		},
		{
			Name:     "invalid empty command",
			expected: ErrInvalidTaskCommand,
			Value:    "",
		},
		{
			Name:     "invalid blank command",
			expected: ErrInvalidTaskCommand,
			Value:    " ",
		},
		{
			Name:     "invalid Exit command",
			expected: ErrUnknownArguments,
			Value:    "exit asd",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			_, err := NewCommand(test.Value, Actions{Add: fakeActions.Add, Run: fakeActions.Run, Stats: fakeActions.Stats})
			if !errors.Is(err, test.expected) {
				t.Errorf("Expected error %v, got %v", test.expected, err)
			}
		})
	}
}

func TestCommand_Execute(t *testing.T) {
	t.Run("Happy execute add", func(t *testing.T) {
		fakeActions := &FakeActions{}
		cmd, err := NewCommand("add task 1", Actions{Add: fakeActions.Add, Run: fakeActions.Run, Stats: fakeActions.Stats})

		if err != nil {
			t.Fatal(err)
		}

		exit := cmd.Execute()

		if exit {
			t.Fatal("Not expected command to exit, but it did")
		}

		if fakeActions.AddCount != 1 {
			t.Fatal("Invalid counts of calls, expected:", 1, "got:", fakeActions.AddCount)
		}

		if fakeActions.RunCount != 0 || fakeActions.StatsCount != 0 {
			t.Fatal("Invalid counts of calls, expected:", 0, "got:", fakeActions.RunCount, "and", fakeActions.StatsCount)
		}

		if fakeActions.AddCommand.Command != "add" || fakeActions.AddCommand.Name != "task" || fakeActions.AddCommand.Duration != 1 {
			t.Fatal("Invalid command arguments, expected: add task 1, got:", fakeActions.AddCommand)
		}
	})

	t.Run("Happy execute run", func(t *testing.T) {
		fakeActions := &FakeActions{}
		cmd, err := NewCommand("run", Actions{Add: fakeActions.Add, Run: fakeActions.Run, Stats: fakeActions.Stats})

		if err != nil {
			t.Fatal(err)
		}

		exit := cmd.Execute()

		if exit {
			t.Fatal("Not expected command to exit, but it did")
		}

		if fakeActions.RunCount != 1 {
			t.Fatal("Invalid counts of calls, expected:", 1, "got:", fakeActions.RunCount)
		}

		if fakeActions.AddCount != 0 || fakeActions.StatsCount != 0 {
			t.Fatal("Invalid counts of calls, expected:", 0, "got:", fakeActions.AddCount, "and", fakeActions.StatsCount)
		}
	})

	t.Run("Happy execute stats", func(t *testing.T) {
		fakeActions := &FakeActions{}
		cmd, err := NewCommand("stats", Actions{Add: fakeActions.Add, Run: fakeActions.Run, Stats: fakeActions.Stats})

		if err != nil {
			t.Fatal(err)
		}

		exit := cmd.Execute()

		if exit {
			t.Fatal("Not expected command to exit, but it did")
		}

		if fakeActions.StatsCount != 1 {
			t.Fatal("Invalid counts of calls, expected:", 1, "got:", fakeActions.StatsCount)
		}

		if fakeActions.AddCount != 0 || fakeActions.RunCount != 0 {
			t.Fatal("Invalid counts of calls, expected:", 0, "got:", fakeActions.RunCount, "and", fakeActions.StatsCount)
		}
	})

	t.Run("Happy execute exit", func(t *testing.T) {
		fakeActions := &FakeActions{}
		cmd, err := NewCommand("exit", Actions{Add: fakeActions.Add, Run: fakeActions.Run, Stats: fakeActions.Stats})

		if err != nil {
			t.Fatal(err)
		}

		exit := cmd.Execute()

		if !exit {
			t.Fatal("Expected exit action, but it did not")
		}

		if fakeActions.AddCount != 0 || fakeActions.RunCount != 0 || fakeActions.StatsCount != 0 {
			t.Fatal("Invalid counts of calls, expected:", 0, "got:", fakeActions.RunCount, "and", fakeActions.StatsCount, "and", fakeActions.AddCount)
		}
	})
}
