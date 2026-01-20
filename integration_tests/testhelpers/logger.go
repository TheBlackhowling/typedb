package testhelpers

// TestLogger is a simple test logger that captures log messages.
type TestLogger struct {
	Debugs []LogEntry
	Infos  []LogEntry
	Warns  []LogEntry
	Errors []LogEntry
}

// LogEntry represents a single log entry with message and key-value pairs.
type LogEntry struct {
	Msg     string
	Keyvals []any
}

// Debug appends a debug log entry.
func (t *TestLogger) Debug(msg string, keyvals ...any) {
	t.Debugs = append(t.Debugs, LogEntry{Msg: msg, Keyvals: keyvals})
}

// Info appends an info log entry.
func (t *TestLogger) Info(msg string, keyvals ...any) {
	t.Infos = append(t.Infos, LogEntry{Msg: msg, Keyvals: keyvals})
}

// Warn appends a warn log entry.
func (t *TestLogger) Warn(msg string, keyvals ...any) {
	t.Warns = append(t.Warns, LogEntry{Msg: msg, Keyvals: keyvals})
}

// Error appends an error log entry.
func (t *TestLogger) Error(msg string, keyvals ...any) {
	t.Errors = append(t.Errors, LogEntry{Msg: msg, Keyvals: keyvals})
}

// Reset clears all captured log entries.
func (t *TestLogger) Reset() {
	t.Debugs = nil
	t.Infos = nil
	t.Warns = nil
	t.Errors = nil
}
