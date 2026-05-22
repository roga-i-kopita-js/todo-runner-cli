package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	RunnerCount int
	LogLevel    string
	LogFormat   string
}

var (
	ErrInvalidRunnerCount = errors.New("RUNNER_COUNT environment variable is not valid, must be a positive integer")
	ErrInvalidLogLevel    = errors.New("LOG_LEVEL environment can be only 'debug' | 'info' | 'warn' | 'error'")
	ErrInvalidLogFormat   = errors.New("LOG_FORMAT environment can be only 'text' | 'json'")
)

func Load() (Config, error) {
	result := Config{LogLevel: "info", RunnerCount: 3, LogFormat: "text"}

	runnerCountEnv, ok := os.LookupEnv("RUNNER_COUNT")
	if ok {
		runnerCount, err := strconv.Atoi(runnerCountEnv)
		if err != nil {
			return Config{}, fmt.Errorf("config load: RUNNER_COUNT: %w, %q", ErrInvalidRunnerCount, runnerCountEnv)
		}
		if runnerCount <= 0 {
			return Config{}, fmt.Errorf("config load: RUNNER_COUNT: %w, %q", ErrInvalidRunnerCount, runnerCountEnv)
		}
		result.RunnerCount = runnerCount
	}

	logFormatEnv, ok := os.LookupEnv("LOG_FORMAT")
	if ok {
		trimmedLogFormat := strings.TrimSpace(logFormatEnv)
		if trimmedLogFormat == "" {
			return Config{}, fmt.Errorf("config load: LOG_FORMAT: %w, %q", ErrInvalidLogFormat, trimmedLogFormat)
		}

		if trimmedLogFormat != "json" && trimmedLogFormat != "text" {
			return Config{}, fmt.Errorf("config load: LOG_FORMAT: %w, %q", ErrInvalidLogFormat, trimmedLogFormat)
		}
		result.LogFormat = trimmedLogFormat
	}

	logLevelEnv, ok := os.LookupEnv("LOG_LEVEL")
	if ok {
		trimmedLogLevel := strings.TrimSpace(logLevelEnv)
		if trimmedLogLevel == "" {
			return Config{}, fmt.Errorf("config load: LOG_LEVEL: %w, %q", ErrInvalidLogLevel, trimmedLogLevel)
		}
		if trimmedLogLevel != "debug" && trimmedLogLevel != "info" && trimmedLogLevel != "warn" && trimmedLogLevel != "error" {
			return Config{}, fmt.Errorf("config load: LOG_LEVEL: %w, %q", ErrInvalidLogLevel, trimmedLogLevel)
		}
		result.LogLevel = trimmedLogLevel
	}

	return result, nil
}
