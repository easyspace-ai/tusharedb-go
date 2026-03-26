package logging

import "log/slog"

func New(level string) *slog.Logger {
	return slog.Default()
}
