// Package logging provides an internal logger that can be configured by the client.
package logging

import (
	"log/slog"
	"sync/atomic"
)

var defaultLogger atomic.Value // *slog.Logger

func init() {
	defaultLogger.Store(slog.Default())
}

// SetDefault sets the internal default logger. If l is nil, slog.Default() is used.
func SetDefault(l *slog.Logger) {
	if l == nil {
		l = slog.Default()
	}
	defaultLogger.Store(l)
}

// L returns the current internal logger.
func L() *slog.Logger {
	if v := defaultLogger.Load(); v != nil {
		if l, ok := v.(*slog.Logger); ok {
			return l
		}
	}
	return slog.Default()
}

// Debug logs a debug message.
func Debug(msg string, args ...any) { L().Debug(msg, args...) }

// Info logs an info message.
func Info(msg string, args ...any) { L().Info(msg, args...) }

// Warn logs a warning message.
func Warn(msg string, args ...any) { L().Warn(msg, args...) }

// Error logs an error message.
func Error(msg string, args ...any) { L().Error(msg, args...) }
