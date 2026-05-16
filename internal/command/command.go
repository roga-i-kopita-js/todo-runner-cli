package command

import (
	"errors"
	"strconv"
	"strings"
)

type TypedCommand string

type Command struct {
	raw     TypedCommand
	actions Actions
	parsed  ParsedCommand
}
type ParsedCommand struct {
	Command  string
	Name     string
	Duration int
}

var (
	ErrInvalidAddTask           = errors.New("invalid add task, should be in the format: <command> <name> <duration in seconds>")
	ErrInvalidTaskCommand       = errors.New("invalid task command, should be 'add' | 'run' | 'stats' | 'exit'")
	ErrInvalidTaskDuration      = errors.New("invalid task duration, should be number")
	ErrInvalidTaskDurationValue = errors.New("invalid task duration, should be grater than 0")
	ErrUnknownArguments         = errors.New("unknown arguments")
)

type Actions struct {
	Add   func(command ParsedCommand)
	Run   func()
	Stats func()
}

func NewCommand(text string, actions Actions) (Command, error) {
	commandModule := Command{raw: TypedCommand(text), actions: actions}
	err := commandModule.validate()
	return commandModule, err
}

func (c *Command) Execute() bool {
	switch c.parsed.Command {
	case "add":
		c.actions.Add(c.parsed)
		return false
	case "run":
		c.actions.Run()
		return false
	case "stats":
		c.actions.Stats()
		return false
	case "exit":
		return true
	default:
		return false
	}
}

func (c *Command) validateAddTaskFormat() (ParsedCommand, error) {
	var t ParsedCommand
	parts := strings.Fields(string(c.raw))

	if len(parts) != 3 {
		return t, ErrInvalidAddTask
	}

	duration, err := strconv.Atoi(parts[2])
	if err != nil {
		return t, ErrInvalidTaskDuration
	}
	if duration <= 0 {
		return t, ErrInvalidTaskDurationValue
	}

	return ParsedCommand{
		Command:  parts[0],
		Name:     parts[1],
		Duration: duration,
	}, nil
}

func (c *Command) validateRunTaskFormat() (ParsedCommand, error) {
	var t ParsedCommand
	parts := strings.Fields(string(c.raw))

	if len(parts) > 1 {
		return t, ErrUnknownArguments
	}

	return ParsedCommand{Command: parts[0]}, nil
}

func (c *Command) validateStatsTaskFormat() (ParsedCommand, error) {
	var t ParsedCommand

	parts := strings.Fields(string(c.raw))

	if len(parts) > 1 {
		return t, ErrUnknownArguments
	}

	return ParsedCommand{Command: parts[0]}, nil
}

func (c *Command) validateExitTaskFormat() (ParsedCommand, error) {
	var t ParsedCommand
	parts := strings.Fields(string(c.raw))

	if len(parts) > 1 {
		return t, ErrUnknownArguments
	}

	return ParsedCommand{Command: parts[0]}, nil
}

func (c *Command) validate() error {
	parts := strings.Fields(string(c.raw))

	if len(parts) == 0 {
		return ErrInvalidTaskCommand
	}

	switch parts[0] {
	case "add":
		parsed, err := c.validateAddTaskFormat()
		if err != nil {
			return err
		}

		c.parsed = parsed
		return nil
	case "run":
		parsed, err := c.validateRunTaskFormat()

		if err != nil {
			return err
		}
		c.parsed = parsed

		return nil
	case "stats":
		parsed, err := c.validateStatsTaskFormat()

		if err != nil {
			return err
		}
		c.parsed = parsed

		return nil
	case "exit":
		parsed, err := c.validateExitTaskFormat()

		if err != nil {
			return err
		}
		c.parsed = parsed

		return nil
	default:
		return ErrInvalidTaskCommand
	}
}
