package typedb

// Logger defines the interface for logging in typedb.
// Users can implement this interface to integrate with their preferred logging library.
type Logger interface {
	// Debug logs a debug message with optional key-value pairs.
	Debug(msg string, keyvals ...any)

	// Info logs an informational message with optional key-value pairs.
	Info(msg string, keyvals ...any)

	// Warn logs a warning message with optional key-value pairs.
	Warn(msg string, keyvals ...any)

	// Error logs an error message with optional key-value pairs.
	Error(msg string, keyvals ...any)
}

// noOpLogger is a no-op logger that discards all log messages.
// This is the default logger when none is provided.
type noOpLogger struct{}

func (n *noOpLogger) Debug(msg string, keyvals ...any) {}
func (n *noOpLogger) Info(msg string, keyvals ...any)  {}
func (n *noOpLogger) Warn(msg string, keyvals ...any)  {}
func (n *noOpLogger) Error(msg string, keyvals ...any) {}

var defaultLogger Logger = &noOpLogger{}

// SetLogger sets the global logger for typedb.
// This logger will be used by all DB instances unless overridden with WithLogger.
func SetLogger(logger Logger) {
	if logger == nil {
		defaultLogger = &noOpLogger{}
		return
	}
	defaultLogger = logger
}

// GetLogger returns the current global logger.
func GetLogger() Logger {
	return defaultLogger
}
