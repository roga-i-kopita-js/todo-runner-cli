package config

import (
	"errors"
	"os"
	"testing"
)

type EntryStruct struct {
	LogLevel    string
	RunnerCount string
	LogFormat   string
}

func TestConfig_Load(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		err := os.Unsetenv("LOG_LEVEL")
		if err != nil {
			t.Fatal(err)
		}
		err1 := os.Unsetenv("LOG_FORMAT")
		if err1 != nil {
			t.Fatal(err1)
		}
		err2 := os.Unsetenv("RUNNER_COUNT")
		if err2 != nil {
			t.Fatal(err2)
		}

		cfg, err := Load()
		if err != nil {
			t.Fatal(err)
		}

		if cfg.LogLevel != "info" {
			t.Fatalf("wrong log level,expected 'info', got %q", cfg.LogLevel)
		}

		if cfg.LogFormat != "text" {
			t.Fatalf("wrong log level,expected 'text', got %q", cfg.LogFormat)
		}

		if cfg.RunnerCount != 3 {
			t.Fatalf("wrong runner count,expected 3, got %d", cfg.RunnerCount)
		}
	})

	fieldsValidationTests := []struct {
		Name     string
		Value    EntryStruct
		Expected error
	}{
		{
			Name:     "valid fields",
			Value:    EntryStruct{LogLevel: "info", RunnerCount: "3", LogFormat: "text"},
			Expected: nil,
		},

		// log level field
		{
			Name:     "empty log level",
			Value:    EntryStruct{LogLevel: "", RunnerCount: "3", LogFormat: "text"},
			Expected: ErrInvalidLogLevel,
		},
		{
			Name:     "empty log level with spaces",
			Value:    EntryStruct{LogLevel: " ", RunnerCount: "3", LogFormat: "text"},
			Expected: ErrInvalidLogLevel,
		},
		{
			Name:     "invalid log level",
			Value:    EntryStruct{LogLevel: "asd", RunnerCount: "3", LogFormat: "text"},
			Expected: ErrInvalidLogLevel,
		},

		//runnerCount field
		{
			Name:     "zero count",
			Value:    EntryStruct{LogLevel: "info", RunnerCount: "0", LogFormat: "text"},
			Expected: ErrInvalidRunnerCount,
		},
		{
			Name:     "empty count",
			Value:    EntryStruct{LogLevel: "info", RunnerCount: "", LogFormat: "text"},
			Expected: ErrInvalidRunnerCount,
		},
		{
			Name:     "empty with spaces count",
			Value:    EntryStruct{LogLevel: "info", RunnerCount: "  ", LogFormat: "text"},
			Expected: ErrInvalidRunnerCount,
		},
		{
			Name:     "negative count",
			Value:    EntryStruct{LogLevel: "info", RunnerCount: "-5", LogFormat: "text"},
			Expected: ErrInvalidRunnerCount,
		},
		{
			Name:     "incorrect count format",
			Value:    EntryStruct{LogLevel: "info", RunnerCount: "asd3", LogFormat: "text"},
			Expected: ErrInvalidRunnerCount,
		},

		// log format field
		{
			Name:     "empty LogFormat with spaces",
			Value:    EntryStruct{LogLevel: "info", RunnerCount: "2", LogFormat: " "},
			Expected: ErrInvalidLogFormat,
		},
		{
			Name:     "empty LogFormat",
			Value:    EntryStruct{LogLevel: "info", RunnerCount: "3", LogFormat: ""},
			Expected: ErrInvalidLogFormat,
		},
		{
			Name:     "invalid LogFormat",
			Value:    EntryStruct{LogLevel: "info", RunnerCount: "3", LogFormat: "asdsd"},
			Expected: ErrInvalidLogFormat,
		},
	}

	for _, test := range fieldsValidationTests {
		t.Run(test.Name, func(t *testing.T) {
			t.Setenv("LOG_LEVEL", test.Value.LogLevel)
			t.Setenv("LOG_FORMAT", test.Value.LogFormat)
			t.Setenv("RUNNER_COUNT", test.Value.RunnerCount)

			_, err := Load()
			if test.Expected == nil && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}

			if !errors.Is(err, test.Expected) {
				t.Fatalf("expected error: %v, got: %v", test.Expected, err)
			}
		})
	}

	t.Run("Happy case", func(t *testing.T) {
		t.Setenv("LOG_LEVEL", "error")
		t.Setenv("LOG_FORMAT", "json")
		t.Setenv("RUNNER_COUNT", "15")

		cfg, err := Load()
		if err != nil {
			t.Fatal(err)
		}

		if cfg.LogLevel != "error" {
			t.Fatalf("wrong log level,expected 'error', got %q", cfg.LogLevel)
		}
		if cfg.RunnerCount != 15 {
			t.Fatalf("wrong runner count,expected 15, got %d", cfg.RunnerCount)
		}
		if cfg.LogFormat != "json" {
			t.Fatalf("wrong log format,expected 'json', got %q", cfg.LogFormat)
		}
	})
}
