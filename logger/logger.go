package logger

import (
	"bytes"
	"errors"
	"fmt"
)

// Logger is a generic logging interface
type Logger interface {
	// Log writes a log entry, spaces are added between operands when neither is a string and a newline is appended.
	Log(level Level, v ...interface{})
	// Logln writes a log entry, spaces are always added between operands and a newline is appended.
	Logln(level Level, v ...interface{})
	// Logf writes a formatted log entry.
	Logf(level Level, format string, v ...interface{})
}

// A Level is a logging priority. Higher levels are more important.
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
	// DPanicLevel level. Logs and call `panic()` in development mode.
	DPanicLevel
	// PanicLevel level. Logs and call `panic()`.
	PanicLevel
	// FatalLevel level. Logs and then calls `logger.Exit(1)`. highest level of severity.
	FatalLevel
)

// MarshalText marshals the Level to text. Note that the text representation
// drops the -Level suffix (see example).
func (l Level) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

// UnmarshalText unmarshals text to a level. Like MarshalText, UnmarshalText
// expects the text representation of a Level to drop the -Level suffix (see
// example).
//
// In particular, this makes it easy to configure logging levels using YAML,
// TOML, or JSON files.
func (l *Level) UnmarshalText(text []byte) error {
	if l == nil {
		return errors.New("can't unmarshal a nil *Level")
	}
	if !l.unmarshalText(text) && !l.unmarshalText(bytes.ToLower(text)) {
		return fmt.Errorf("unrecognized level: %q", text)
	}
	return nil
}

func (l *Level) unmarshalText(text []byte) bool {
	_, skip := l.UnpackSkip()

	switch string(text) {
	case "debug", "DEBUG":
		*l = DebugLevel
	case "info", "INFO", "": // make the zero value useful
		*l = InfoLevel
	case "warn", "WARN":
		*l = WarnLevel
	case "error", "ERROR":
		*l = ErrorLevel
	case "dpanic", "DPANIC":
		*l = DPanicLevel
	case "panic", "PANIC":
		*l = PanicLevel
	case "fatal", "FATAL":
		*l = FatalLevel
	default:
		return false
	}

	*l = l.PackSkip(skip)

	return true
}

// String returns a lower-case ASCII representation of the log level.
func (l Level) String() string {
	l, _ = l.UnpackSkip()

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
	case DPanicLevel:
		return "dpanic"
	case PanicLevel:
		return "panic"
	case FatalLevel:
		return "fatal"
	default:
		return fmt.Sprintf("LEVEL(%d)", l)
	}
}

// CapitalString returns an all-caps ASCII representation of the log level.
func (l Level) CapitalString() string {
	l, _ = l.UnpackSkip()

	// Printing levels in all-caps is common enough that we should export this
	// functionality.
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case DPanicLevel:
		return "DPANIC"
	case PanicLevel:
		return "PANIC"
	case FatalLevel:
		return "FATAL"
	default:
		return fmt.Sprintf("LEVEL(%d)", l)
	}
}

// Set converts a level string into a logger Level value.
// returns error if the input string does not match known values.
func (l *Level) Set(str string) error {
	return l.UnmarshalText([]byte(str))
}

// Get gets the level for the flag.Getter interface.
func (l Level) Get() interface{} {
	return l
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
	return l & 0x0f, int8(l >> 4)
}

// Enabled returns true if the given level is at or above this level.
func (l Level) Enabled(level Level) bool {
	l, _ = l.UnpackSkip()
	level, _ = level.UnpackSkip()

	return level >= l
}
