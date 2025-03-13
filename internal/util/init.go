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

	err := ko.Load(env.ProviderWithValue("TRACKER_", ".", func(s string, v string) (string, interface{}) {
		key := strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(s, "TRACKER_")), "__", ".")
		if strings.Contains(v, " ") {
			return key, strings.Split(v, " ")
		}
		return key, v
	}), nil)

	if err != nil {
		lo.Error("could not override config from env vars", "error", err)
		os.Exit(1)
	}

	if os.Getenv("DEBUG") != "" {
		ko.Print()
	}

	return ko
}
