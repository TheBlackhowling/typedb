package typedb

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
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

// TestDB_Exec_Logging verifies that Debug and Error logs are emitted during Exec operations
func TestDB_Exec_Logging(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	logger := &testLogger{}
	typedbDB := NewDBWithLogger(db, "test", 5*time.Second, logger)
	ctx := context.Background()

	t.Run("success logs debug", func(t *testing.T) {
		logger.debugs = nil // Reset logs
		mock.ExpectExec("INSERT INTO users").
			WithArgs("test").
			WillReturnResult(sqlmock.NewResult(1, 1))

		_, err := typedbDB.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "test")
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		// Verify Debug log was emitted
		if len(logger.debugs) == 0 {
			t.Fatal("Expected Debug log for Exec, got none")
		}
		if logger.debugs[0].msg != "Executing query" {
			t.Errorf("Expected Debug log message 'Executing query', got %q", logger.debugs[0].msg)
		}
		// Verify query and args are in keyvals
		foundQuery := false
		for i := 0; i < len(logger.debugs[0].keyvals)-1; i += 2 {
			if logger.debugs[0].keyvals[i] == "query" {
				foundQuery = true
				if !strings.Contains(logger.debugs[0].keyvals[i+1].(string), "INSERT INTO users") {
					t.Errorf("Expected query to contain 'INSERT INTO users', got %v", logger.debugs[0].keyvals[i+1])
				}
			}
		}
		if !foundQuery {
			t.Error("Expected 'query' key in Debug log keyvals")
		}
	})

	t.Run("error logs error", func(t *testing.T) {
		logger.errors = nil // Reset logs
		expectedErr := errors.New("database error")
		mock.ExpectExec("INSERT INTO users").
			WithArgs("test").
			WillReturnError(expectedErr)

		_, err := typedbDB.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "test")
		if err != expectedErr {
			t.Fatalf("Expected error %v, got %v", expectedErr, err)
		}

		// Verify Error log was emitted
		if len(logger.errors) == 0 {
			t.Fatal("Expected Error log for Exec failure, got none")
		}
		if logger.errors[0].msg != "Query execution failed" {
			t.Errorf("Expected Error log message 'Query execution failed', got %q", logger.errors[0].msg)
		}
		// Verify error is in keyvals
		foundError := false
		for i := 0; i < len(logger.errors[0].keyvals)-1; i += 2 {
			if logger.errors[0].keyvals[i] == "error" {
				foundError = true
				if logger.errors[0].keyvals[i+1] != expectedErr {
					t.Errorf("Expected error %v, got %v", expectedErr, logger.errors[0].keyvals[i+1])
				}
			}
		}
		if !foundError {
			t.Error("Expected 'error' key in Error log keyvals")
		}
	})
}

// TestDB_QueryAll_Logging verifies that Debug and Error logs are emitted during QueryAll operations
func TestDB_QueryAll_Logging(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	logger := &testLogger{}
	typedbDB := NewDBWithLogger(db, "test", 5*time.Second, logger)
	ctx := context.Background()

	t.Run("success logs debug", func(t *testing.T) {
		logger.debugs = nil // Reset logs
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Alice").
			AddRow(2, "Bob")
		mock.ExpectQuery("SELECT id, name FROM users").
			WillReturnRows(rows)

		_, err := typedbDB.QueryAll(ctx, "SELECT id, name FROM users")
		if err != nil {
			t.Fatalf("QueryAll failed: %v", err)
		}

		// Verify Debug log was emitted
		if len(logger.debugs) == 0 {
			t.Fatal("Expected Debug log for QueryAll, got none")
		}
		if logger.debugs[0].msg != "Querying all rows" {
			t.Errorf("Expected Debug log message 'Querying all rows', got %q", logger.debugs[0].msg)
		}
	})

	t.Run("error logs error", func(t *testing.T) {
		logger.errors = nil // Reset logs
		expectedErr := errors.New("query error")
		mock.ExpectQuery("SELECT id, name FROM users").
			WillReturnError(expectedErr)

		_, err := typedbDB.QueryAll(ctx, "SELECT id, name FROM users")
		if err != expectedErr {
			t.Fatalf("Expected error %v, got %v", expectedErr, err)
		}

		// Verify Error log was emitted
		if len(logger.errors) == 0 {
			t.Fatal("Expected Error log for QueryAll failure, got none")
		}
		if logger.errors[0].msg != "Query failed" {
			t.Errorf("Expected Error log message 'Query failed', got %q", logger.errors[0].msg)
		}
	})
}

// TestDB_Begin_Logging verifies that Debug logs are emitted during Begin operations
func TestDB_Begin_Logging(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	logger := &testLogger{}
	typedbDB := NewDBWithLogger(db, "test", 5*time.Second, logger)
	ctx := context.Background()

	t.Run("success logs debug", func(t *testing.T) {
		logger.debugs = nil // Reset logs
		mock.ExpectBegin()

		tx, err := typedbDB.Begin(ctx, nil)
		if err != nil {
			t.Fatalf("Begin failed: %v", err)
		}
		if tx == nil {
			t.Fatal("Expected transaction, got nil")
		}

		// Verify Debug log was emitted
		if len(logger.debugs) == 0 {
			t.Fatal("Expected Debug log for Begin, got none")
		}
		if logger.debugs[0].msg != "Beginning transaction" {
			t.Errorf("Expected Debug log message 'Beginning transaction', got %q", logger.debugs[0].msg)
		}
	})

	t.Run("error logs error", func(t *testing.T) {
		logger.errors = nil // Reset logs
		expectedErr := errors.New("begin error")
		mock.ExpectBegin().WillReturnError(expectedErr)

		_, err := typedbDB.Begin(ctx, nil)
		if err != expectedErr {
			t.Fatalf("Expected error %v, got %v", expectedErr, err)
		}

		// Verify Error log was emitted
		if len(logger.errors) == 0 {
			t.Fatal("Expected Error log for Begin failure, got none")
		}
		if logger.errors[0].msg != "Failed to begin transaction" {
			t.Errorf("Expected Error log message 'Failed to begin transaction', got %q", logger.errors[0].msg)
		}
	})
}

// TestTx_Commit_Logging verifies that Info logs are emitted during Commit operations
func TestTx_Commit_Logging(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	logger := &testLogger{}

	mock.ExpectBegin()
	mockTx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	typedbTx := &Tx{
		tx:      mockTx,
		timeout: 5 * time.Second,
		logger:  logger,
	}

	t.Run("success logs info", func(t *testing.T) {
		logger.infos = nil // Reset logs
		mock.ExpectCommit()

		err := typedbTx.Commit()
		if err != nil {
			t.Fatalf("Commit failed: %v", err)
		}

		// Verify Info log was emitted
		if len(logger.infos) == 0 {
			t.Fatal("Expected Info log for Commit, got none")
		}
		if logger.infos[0].msg != "Committing transaction" {
			t.Errorf("Expected Info log message 'Committing transaction', got %q", logger.infos[0].msg)
		}
	})

	t.Run("error logs error", func(t *testing.T) {
		// Need a new transaction for error test
		mock.ExpectBegin()
		mockTx2, err := db.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}
		typedbTx2 := &Tx{
			tx:      mockTx2,
			timeout: 5 * time.Second,
			logger:  logger,
		}

		logger.errors = nil // Reset logs
		expectedErr := errors.New("commit error")
		mock.ExpectCommit().WillReturnError(expectedErr)

		err = typedbTx2.Commit()
		if err != expectedErr {
			t.Fatalf("Expected error %v, got %v", expectedErr, err)
		}

		// Verify Error log was emitted
		if len(logger.errors) == 0 {
			t.Fatal("Expected Error log for Commit failure, got none")
		}
		if logger.errors[0].msg != "Transaction commit failed" {
			t.Errorf("Expected Error log message 'Transaction commit failed', got %q", logger.errors[0].msg)
		}
	})
}

// TestTx_Rollback_Logging verifies that Info logs are emitted during Rollback operations
func TestTx_Rollback_Logging(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	logger := &testLogger{}

	mock.ExpectBegin()
	mockTx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	typedbTx := &Tx{
		tx:      mockTx,
		timeout: 5 * time.Second,
		logger:  logger,
	}

	t.Run("success logs info", func(t *testing.T) {
		logger.infos = nil // Reset logs
		mock.ExpectRollback()

		err := typedbTx.Rollback()
		if err != nil {
			t.Fatalf("Rollback failed: %v", err)
		}

		// Verify Info log was emitted
		if len(logger.infos) == 0 {
			t.Fatal("Expected Info log for Rollback, got none")
		}
		if logger.infos[0].msg != "Rolling back transaction" {
			t.Errorf("Expected Info log message 'Rolling back transaction', got %q", logger.infos[0].msg)
		}
	})
}

// TestDB_Close_Logging verifies that Info logs are emitted during Close operations
func TestDB_Close_Logging(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	logger := &testLogger{}
	typedbDB := NewDBWithLogger(db, "test", 5*time.Second, logger)

	t.Run("success logs info", func(t *testing.T) {
		logger.infos = nil // Reset logs
		mock.ExpectClose()

		err := typedbDB.Close()
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}

		// Verify Info log was emitted
		if len(logger.infos) == 0 {
			t.Fatal("Expected Info log for Close, got none")
		}
		if logger.infos[0].msg != "Closing database connection" {
			t.Errorf("Expected Info log message 'Closing database connection', got %q", logger.infos[0].msg)
		}
	})
}

// TestPerInstanceLogger verifies that per-instance logger overrides global logger
func TestPerInstanceLogger(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	globalLogger := &testLogger{}
	instanceLogger := &testLogger{}

	// Set global logger
	SetLogger(globalLogger)

	// Create DB with per-instance logger
	typedbDB := NewDBWithLogger(db, "test", 5*time.Second, instanceLogger)
	ctx := context.Background()

	mock.ExpectExec("INSERT INTO users").
		WithArgs("test").
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err = typedbDB.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "test")
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	// Verify instance logger received the log, not global logger
	if len(instanceLogger.debugs) == 0 {
		t.Error("Expected instance logger to receive Debug log")
	}
	if len(globalLogger.debugs) != 0 {
		t.Error("Expected global logger to NOT receive log when per-instance logger is set")
	}
}
