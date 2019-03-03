package notification

import "strings"
import "runtime/debug"

type Level uint32

const (
	// UNSPECIFIED error level = 0
	UNSPECIFIED Level = iota
	TRACE
	DEBUG
	INFO
	WARNING
	ERROR
	CRITICAL
)

// Error ...
type Error struct {
	Stack   string
	Message string
	Level   string
}

func FormatError(level, msg string) Error {
	stack := string(debug.Stack())

	return Error{
		Stack:   stack,
		Message: msg,
		Level:   parseLevel(level).String(),
	}
}

func parseLevel(level string) Level {
	level = strings.ToUpper(level)
	switch level {
	case "UNSPECIFIED":
		return UNSPECIFIED
	case "TRACE":
		return TRACE
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARNING
	case "ERROR":
		return ERROR
	case "CRITICAL":
		return CRITICAL
	default:
		return UNSPECIFIED
	}
}

// String implements Stringer.
func (level Level) String() string {
	switch level {
	case UNSPECIFIED:
		return "UNSPECIFIED"
	case TRACE:
		return "TRACE"
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case CRITICAL:
		return "CRITICAL"
	default:
		return "<unknown>"
	}
}
