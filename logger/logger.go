package logger

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

// Enabled returns true if the given level is at or above this level.
func (l Level) Enabled(level Level) bool {
	return level >= l
}

// GetLevel converts a level string into a logger Level value.
// returns info level if the input string does not match known values.
func GetLevel(levelStr string) Level {
	switch levelStr {
	case TraceLevel.String():
		return TraceLevel
	case DebugLevel.String():
		return DebugLevel
	case InfoLevel.String():
		return InfoLevel
	case WarnLevel.String():
		return WarnLevel
	case ErrorLevel.String():
		return ErrorLevel
	case PanicLevel.String():
		return PanicLevel
	case FatalLevel.String():
		return FatalLevel
	}
	return InfoLevel
}

const (
	HelperFlag int8 = 0x10
)
