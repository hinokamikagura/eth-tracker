// Package util provides utility functions for initialization and configuration.
package util

import (
	"log/slog"
	"os"
	"strings"

	"github.com/kamikazechaser/common/logg"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const (
	// envDebug enables debug logging when set
	envDebug = "DEBUG"
	// envDev enables development mode (human-readable logs) when set
	envDev = "DEV"
	// envPrefix is the prefix for environment variables that override config
	envPrefix = "TRACKER_"
	// envSeparator is used to split array values in environment variables
	envSeparator = " "
	// envNestedSeparator is used to represent nested config keys in environment variables
	envNestedSeparator = "__"
)

// InitLogger initializes and returns a structured logger based on environment variables.
// DEBUG: enables debug level logging
// DEV: enables debug level logging with human-readable format
func InitLogger() *slog.Logger {
	loggOpts := logg.LoggOpts{
		FormatType: logg.Logfmt,
		LogLevel:   slog.LevelInfo,
	}

	if os.Getenv(envDebug) != "" {
		loggOpts.LogLevel = slog.LevelDebug
	}

	if os.Getenv(envDev) != "" {
		loggOpts.LogLevel = slog.LevelDebug
		loggOpts.FormatType = logg.Human
	}

	return logg.NewLogg(loggOpts)
}

// InitConfig loads configuration from a TOML file and environment variables.
// Environment variables prefixed with TRACKER_ override file values.
// Nested keys can be specified using double underscores (e.g., TRACKER_CORE__POOL_SIZE).
// Array values can be specified as space-separated strings.
func InitConfig(lo *slog.Logger, confFilePath string) *koanf.Koanf {
	ko := koanf.New(".")

	// Load configuration file
	confFile := file.Provider(confFilePath)
	if err := ko.Load(confFile, toml.Parser()); err != nil {
		lo.Error("failed to load configuration file", "file", confFilePath, "error", err)
		os.Exit(1)
	}

	// Load environment variable overrides
	err := ko.Load(env.ProviderWithValue(envPrefix, ".", func(s string, v string) (string, interface{}) {
		// Convert TRACKER_KEY__NESTED to key.nested
		key := strings.ReplaceAll(
			strings.ToLower(strings.TrimPrefix(s, envPrefix)),
			envNestedSeparator,
			".",
		)

		// Handle array values (space-separated)
		if strings.Contains(v, envSeparator) {
			return key, strings.Split(v, envSeparator)
		}

		return key, v
	}), nil)

	if err != nil {
		lo.Error("failed to load environment variable overrides", "error", err)
		os.Exit(1)
	}

	// Print configuration in debug mode
	if os.Getenv(envDebug) != "" {
		ko.Print()
	}

	return ko
}
