package log

import (
    "log/slog"
)

// Init sets up the global logger.
func Init() {
    logger := slog.New(slog.NewTextHandler(&slog.HandlerOptions{AddSource: false}))
    slog.SetDefault(logger)
}

// L returns the default logger.
func L() *slog.Logger {
    return slog.Default()
}
