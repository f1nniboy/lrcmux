package logging

import (
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Level  string `toml:"level" comment:"debug, info, warn, error"`
	Format string `toml:"format" comment:"text or json"`
}

func Init(cfg Config) {
	var level slog.Level
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	if strings.ToLower(cfg.Format) == "json" {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}
	slog.SetDefault(slog.New(handler))
}

func New(component string) *slog.Logger {
	return slog.Default().With("component", component)
}
