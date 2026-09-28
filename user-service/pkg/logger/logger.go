package logger

import (
	"log/slog"
	"os"
)

func New(enviournment string) *slog.Logger {
	var handler slog.Handler

	if enviournment == "profuction" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	return slog.New(handler)
}
