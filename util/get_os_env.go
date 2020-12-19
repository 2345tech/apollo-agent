package util

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

var lastErr error //nolint:gochecknoglobals // Exported.

// LastErr returns last error happens while parsing environment variable
// by any function of this package. Following calls will return nil until
// new error happens.
func LastErr() error {
	err := lastErr
	lastErr = nil
	return err
}

// Bool returns value of environment variable name as a bool.
// It will return def if environment variable is missing, empty or
// failed to parse as bool.
func Bool(name string, def bool) bool {
	value := os.Getenv(name)
	if value == "" {
		return def
	}
	b, err := strconv.ParseBool(value)
	if err != nil {
		lastErr = fmt.Errorf("parse $%s=%q as bool: %w", name, value, err)
		return def
	}
	return b
}

// Dur returns value of environment variable name as a time.Duration.
// It will return def if environment variable is missing, empty or
// failed to parse as a time.Duration.
func Dur(name string, def time.Duration) time.Duration {
	value := os.Getenv(name)
	if value == "" {
		return def
	}
	dur, err := time.ParseDuration(value)
	if err != nil {
		lastErr = fmt.Errorf("parse $%s=%q as duration: %w", name, value, err)
		return def
	}
	return dur
}

// Float returns value of environment variable name as a float64.
// It will return def if environment variable is missing, empty or
// failed to parse as a float64.
func Float(name string, def float64) float64 {
	value := os.Getenv(name)
	if value == "" {
		return def
	}
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		lastErr = fmt.Errorf("parse $%s=%q as float: %w", name, value, err)
		return def
	}
	return v
}

// Int returns value of environment variable name as an int.
// It will return def if environment variable is missing, empty or
// failed to parse as an int.
func Int(name string, def int) int {
	value := os.Getenv(name)
	if value == "" {
		return def
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		lastErr = fmt.Errorf("parse $%s=%q as int: %w", name, value, err)
		return def
	}
	return i
}

// Str returns value of environment variable name as a string.
// It will return def if environment variable is missing or empty.
func Str(name, def string) string {
	value := os.Getenv(name)
	if value == "" {
		return def
	}
	return value
}
