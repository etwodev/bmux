package log

import "time"

// NoOpLogger is a Logger implementation that performs no operations.
// It can be used as a default or placeholder logger to avoid nil checks.
//
// Example usage:
//
//	var logger Logger = &NoOpLogger{}
//	logger.Info().Str("key", "value").Msg("This message is ignored")
type NoOpLogger struct{}

// Debug returns a no-op Entry for debug-level logs.
func (l *NoOpLogger) Debug() Entry { return &noopEntry{} }

// Info returns a no-op Entry for info-level logs.
func (l *NoOpLogger) Info() Entry { return &noopEntry{} }

// Warn returns a no-op Entry for warn-level logs.
func (l *NoOpLogger) Warn() Entry { return &noopEntry{} }

// Error returns a no-op Entry for error-level logs.
func (l *NoOpLogger) Error() Entry { return &noopEntry{} }

// Fatal returns a no-op Entry for fatal-level logs.
func (l *NoOpLogger) Fatal() Entry { return &noopEntry{} }

// noopEntry is an Entry implementation that discards all log data.
type noopEntry struct{}

// Str is a no-op for adding a string field.
func (n *noopEntry) Str(string, string) Entry { return n }

// Dur is a no-op for adding a time.Duration field.
func (n *noopEntry) Dur(string, time.Duration) Entry { return n }

// Int is a no-op for adding an int field.
func (n *noopEntry) Int(string, int) Entry { return n }

// Bool is a no-op for adding a bool field.
func (n *noopEntry) Bool(string, bool) Entry { return n }

// Err is a no-op for adding an error field.
func (n *noopEntry) Err(error) Entry { return n }

// Msg is a no-op for sending the log message.
func (n *noopEntry) Msg(string) {}
