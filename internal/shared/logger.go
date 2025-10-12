package shared

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Logger provides structured logging capabilities
type Logger struct {
	*log.Logger
	level LogLevel
}

// LogLevel represents logging level
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarning
	LogLevelError
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarning:
		return "WARNING"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// NewLogger creates a new logger instance
func NewLogger(level LogLevel) *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds),
		level:  level,
	}
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// Debug logs a debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level <= LogLevelDebug {
		l.Printf("[DEBUG] "+format, v...)
	}
}

// Info logs an info message
func (l *Logger) Info(format string, v ...interface{}) {
	if l.level <= LogLevelInfo {
		l.Printf("[INFO] "+format, v...)
	}
}

// Warning logs a warning message
func (l *Logger) Warning(format string, v ...interface{}) {
	if l.level <= LogLevelWarning {
		l.Printf("[WARNING] "+format, v...)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	if l.level <= LogLevelError {
		l.Printf("[ERROR] "+format, v...)
	}
}

// WithFields logs a message with additional fields
func (l *Logger) WithFields(level LogLevel, format string, fields map[string]interface{}, v ...interface{}) {
	if l.level <= level {
		fieldStr := ""
		for k, val := range fields {
			if fieldStr != "" {
				fieldStr += ", "
			}
			fieldStr += k + "=" + formatValue(val)
		}
		if fieldStr != "" {
			l.Printf("[%s] "+format+" | %s", append([]interface{}{level.String()}, append(v, fieldStr)...)...)
		} else {
			l.Printf("[%s] "+format, append([]interface{}{level.String()}, v...)...)
		}
	}
}

// formatValue formats a value for logging
func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return formatInt(int64(val))
	case int64:
		return formatInt(val)
	case int32:
		return formatInt(int64(val))
	case float64:
		return formatFloat(val)
	case float32:
		return formatFloat(float64(val))
	case bool:
		if val {
			return "true"
		}
		return "false"
	case time.Time:
		return val.Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func formatInt(v int64) string {
	return fmt.Sprintf("%d", v)
}

func formatFloat(v float64) string {
	return fmt.Sprintf("%f", v)
}

// Default logger instance
var defaultLogger = NewLogger(LogLevelInfo)

// SetDefaultLogger sets the default logger
func SetDefaultLogger(logger *Logger) {
	defaultLogger = logger
}

// Debug logs a debug message using the default logger
func Debug(format string, v ...interface{}) {
	defaultLogger.Debug(format, v...)
}

// Info logs an info message using the default logger
func Info(format string, v ...interface{}) {
	defaultLogger.Info(format, v...)
}

// Warning logs a warning message using the default logger
func Warning(format string, v ...interface{}) {
	defaultLogger.Warning(format, v...)
}

// Error logs an error message using the default logger
func Error(format string, v ...interface{}) {
	defaultLogger.Error(format, v...)
}
