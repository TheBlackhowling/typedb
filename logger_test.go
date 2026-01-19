package typedb

import (
	"testing"
)

// testLogger is a simple test logger that captures log messages.
type testLogger struct {
	debugs []logEntry
	infos  []logEntry
	warns  []logEntry
	errors []logEntry
}

type logEntry struct {
	msg     string
	keyvals []any
}

func (t *testLogger) Debug(msg string, keyvals ...any) {
	t.debugs = append(t.debugs, logEntry{msg: msg, keyvals: keyvals})
}

func (t *testLogger) Info(msg string, keyvals ...any) {
	t.infos = append(t.infos, logEntry{msg: msg, keyvals: keyvals})
}

func (t *testLogger) Warn(msg string, keyvals ...any) {
	t.warns = append(t.warns, logEntry{msg: msg, keyvals: keyvals})
}

func (t *testLogger) Error(msg string, keyvals ...any) {
	t.errors = append(t.errors, logEntry{msg: msg, keyvals: keyvals})
}

func TestLoggerInterface(t *testing.T) {
	logger := &testLogger{}

	// Test that logger can be set globally
	SetLogger(logger)
	if GetLogger() != logger {
		t.Error("GetLogger() should return the logger set by SetLogger()")
	}

	// Test that logger methods work
	logger.Debug("test debug", "key", "value")
	logger.Info("test info", "key", "value")
	logger.Warn("test warn", "key", "value")
	logger.Error("test error", "key", "value")

	if len(logger.debugs) != 1 {
		t.Errorf("Expected 1 debug log, got %d", len(logger.debugs))
	}
	if len(logger.infos) != 1 {
		t.Errorf("Expected 1 info log, got %d", len(logger.infos))
	}
	if len(logger.warns) != 1 {
		t.Errorf("Expected 1 warn log, got %d", len(logger.warns))
	}
	if len(logger.errors) != 1 {
		t.Errorf("Expected 1 error log, got %d", len(logger.errors))
	}

	// Test that nil logger uses no-op logger
	SetLogger(nil)
	if GetLogger() == nil {
		t.Error("GetLogger() should return a no-op logger, not nil")
	}
}

func TestNoOpLogger(t *testing.T) {
	logger := &noOpLogger{}

	// These should not panic
	logger.Debug("test")
	logger.Info("test")
	logger.Warn("test")
	logger.Error("test")
}
