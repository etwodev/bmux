package log

import (
	"context"
	"time"
)

// ctxKey is a private type used as a key for storing values in context.
// This helps avoid collisions with other context keys.
type ctxKey string

// LoggerCtxKey is the context key used to store and retrieve
// the Logger instance from a context.Context.
var LoggerCtxKey = ctxKey("logger")

// Logger defines a structured logging interface with
// methods for different log levels.
type Logger interface {
	// Debug starts a log entry with Debug level.
	Debug() Entry
	// Info starts a log entry with Info level.
	Info() Entry
	// Warn starts a log entry with Warn level.
	Warn() Entry
	// Error starts a log entry with Error level.
	Error() Entry
	// Fatal starts a log entry with Fatal level.
	Fatal() Entry
}

// Entry represents a single log event builder.
// It supports chaining of key-value pairs and message emission.
type Entry interface {
	// Str adds a string key-value pair to the log entry.
	Str(key, value string) Entry
	// Dur adds a time.Duration key-value pair to the log entry.
	Dur(key string, value time.Duration) Entry
	// Int adds an int key-value pair to the log entry.
	Int(key string, value int) Entry
	// Bool adds a bool key-value pair to the log entry.
	Bool(key string, value bool) Entry
	// Err adds an error to the log entry.
	Err(error) Entry
	// Msg sends the log entry with the given message.
	Msg(msg string)
}

// FromContext attempts to retrieve a Logger from the provided context.Context.
// If no logger is found, it returns nil.
//
// Example usage:
//
//	logger := log.FromContext(ctx)
//	if logger != nil {
//	    logger.Info().Str("user", "alice").Msg("User logged in")
//	}
func FromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(LoggerCtxKey).(Logger); ok {
		return logger
	}
	return nil
}
