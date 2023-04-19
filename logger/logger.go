package logger

import (
	"errors"
	"fmt"
)

// Logger is a generic logging interface
type Logger interface {
	// Log writes a log entry, spaces are added between operands when neither is a string and a newline is appended
	Log(level Level, v ...interface{})
	// Logln writes a log entry, spaces are always added between operands and a newline is appended
	Logln(level Level, v ...interface{})
	// Logf writes a formatted log entry
	Logf(level Level, format string, v ...interface{})
}

type Level int8

const (
	// TraceLevel level. Designates finer-grained informational events than the Debug.
	TraceLevel Level = iota
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
	// InfoLevel is the default logging priority.
	// General operational entries about what's going on inside the application.
	InfoLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	ErrorLevel
	// PanicLevel level. Logs and call `panic()`.
	PanicLevel
	// FatalLevel level. Logs and then calls `logger.Exit(1)`. highest level of severity.
	FatalLevel
)

func (l Level) String() string {
	l &= 0x0f

	switch l {
	case TraceLevel:
		return "trace"
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case PanicLevel:
		return "panic"
	case FatalLevel:
		return "fatal"
	}

	return "unknown"
}

// Set converts a level string into a logger Level value.
// returns error if the input string does not match known values.
func (l *Level) Set(str string) error {
	if l == nil {
		return errors.New("can't set a nil *Level")
	}

	switch str {
	case TraceLevel.String():
		*l = TraceLevel
	case DebugLevel.String():
		*l = DebugLevel
	case InfoLevel.String():
		*l = InfoLevel
	case WarnLevel.String():
		*l = WarnLevel
	case ErrorLevel.String():
		*l = ErrorLevel
	case PanicLevel.String():
		*l = PanicLevel
	case FatalLevel.String():
		*l = FatalLevel
	}

	return fmt.Errorf("unrecognized level: %q", str)
}

// Enabled returns true if the given level is at or above this level.
func (l Level) Enabled(level Level) bool {
	l &= 0x0f
	return level >= l
}

// PackSkip returns a new Level value with an additional skip offset encoded in the high bits.
// The skip value indicates the number of additional stack frames to skip before logging.
// It is useful for providing more contextual information in the log.
func (l Level) PackSkip(skip int8) Level {
	return l | (Level(skip) << 4)
}

// UnpackSkip extracts the original Level value and the skip offset from a packed Level value.
// If the skip offset is not present in the packed value, a default value of 1 is used.
// It is useful for decoding the skip offset and recovering the original Level value.
func (l Level) UnpackSkip() (Level, int8) {
	skip := int8(l >> 4)
	return l & 0x0f, skip
}
