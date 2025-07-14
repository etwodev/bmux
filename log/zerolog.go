package log

import (
	"time"

	"github.com/rs/zerolog"
)

// ZeroLogger is an adapter that implements the Logger interface
// using the zerolog.Logger from the zerolog package.
//
// Example usage:
//
//	import (
//	  "github.com/rs/zerolog"
//	  "os"
//	)
//
//	func main() {
//	  zlogger := zerolog.New(os.Stdout).With().Timestamp().Logger()
//	  logger := NewZeroLogger(zlogger)
//
//	  logger.Info().Str("user", "alice").Msg("User logged in")
//	}
type ZeroLogger struct {
	z zerolog.Logger
}

// NewZeroLogger creates a new ZeroLogger wrapping a zerolog.Logger instance.
func NewZeroLogger(z zerolog.Logger) *ZeroLogger {
	return &ZeroLogger{z}
}

// Debug starts a debug-level log entry.
func (zl *ZeroLogger) Debug() Entry { return &zeroEntry{zl.z.Debug()} }

// Info starts an info-level log entry.
func (zl *ZeroLogger) Info() Entry { return &zeroEntry{zl.z.Info()} }

// Warn starts a warn-level log entry.
func (zl *ZeroLogger) Warn() Entry { return &zeroEntry{zl.z.Warn()} }

// Error starts an error-level log entry.
func (zl *ZeroLogger) Error() Entry { return &zeroEntry{zl.z.Error()} }

// Fatal starts a fatal-level log entry.
func (zl *ZeroLogger) Fatal() Entry { return &zeroEntry{zl.z.Fatal()} }

// zeroEntry wraps zerolog.Event to implement the Entry interface.
type zeroEntry struct {
	e *zerolog.Event
}

// Str adds a string key-value pair to the log entry.
func (z *zeroEntry) Str(k, v string) Entry {
	z.e.Str(k, v)
	return z
}

// Dur adds a time.Duration key-value pair to the log entry.
func (z *zeroEntry) Dur(k string, v time.Duration) Entry {
	z.e.Dur(k, v)
	return z
}

// Int adds an int key-value pair to the log entry.
func (z *zeroEntry) Int(k string, v int) Entry {
	z.e.Int(k, v)
	return z
}

// Bool adds a bool key-value pair to the log entry.
func (z *zeroEntry) Bool(k string, v bool) Entry {
	z.e.Bool(k, v)
	return z
}

// Err adds an error to the log entry.
func (z *zeroEntry) Err(e error) Entry {
	z.e.Err(e)
	return z
}

// Msg sends the log entry with the specified message.
func (z *zeroEntry) Msg(m string) {
	z.e.Msg(m)
}
