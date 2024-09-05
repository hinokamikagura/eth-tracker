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

func InitLogger() *slog.Logger {
	loggOpts := logg.LoggOpts{
		FormatType: logg.Logfmt,
		LogLevel:   slog.LevelInfo,
	}

	if os.Getenv("DEBUG") != "" {
		loggOpts.LogLevel = slog.LevelDebug
	}

	if os.Getenv("DEV") != "" {
		loggOpts.LogLevel = slog.LevelDebug
		loggOpts.FormatType = logg.Human
	}

	return logg.NewLogg(loggOpts)
}

func InitConfig(lo *slog.Logger, confFilePath string) *koanf.Koanf {
	var (
		ko = koanf.New(".")
	)

	confFile := file.Provider(confFilePath)
	if err := ko.Load(confFile, toml.Parser()); err != nil {
		lo.Error("could not parse configuration file", "error", err)
		os.Exit(1)
	}

	if err := ko.Load(env.Provider("TRACKER_", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(
			strings.TrimPrefix(s, "TRACKER_")), "__", ".")
	}), nil); err != nil {
		lo.Error("could not override config from env vars", "error", err)
		os.Exit(1)
	}

	return ko
}
