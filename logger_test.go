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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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

// TestLogQueriesConfig verifies that WithLogQueries option controls query logging
func TestLogQueriesConfig(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	logger := &testLogger{}
	ctx := context.Background()

	t.Run("LogQueries=true logs queries", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		typedbDB, err := OpenWithoutValidation("test", "test", WithLogger(logger), WithLogQueries(true))
		if err != nil {
			// OpenWithoutValidation will fail with invalid DSN, but we can use NewDBWithLoggerAndFlags directly
			typedbDB = NewDBWithLoggerAndFlags(db, "test", 5*time.Second, logger, true, true)
		} else {
			typedbDB.Close()
			typedbDB = NewDBWithLoggerAndFlags(db, "test", 5*time.Second, logger, true, true)
		}

		mock.ExpectExec("INSERT INTO users").
			WithArgs("test").
			WillReturnResult(sqlmock.NewResult(1, 1))

		_, err = typedbDB.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "test")
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		// Verify query was logged
		if len(logger.debugs) == 0 {
			t.Fatal("Expected Debug log with query when LogQueries=true")
		}
		foundQuery := false
		for i := 0; i < len(logger.debugs[0].keyvals)-1; i += 2 {
			if logger.debugs[0].keyvals[i] == "query" {
				foundQuery = true
				break
			}
		}
		if !foundQuery {
			t.Error("Expected 'query' key in log when LogQueries=true")
		}
	})

	t.Run("LogQueries=false does not log queries", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		typedbDB := NewDBWithLoggerAndFlags(db, "test", 5*time.Second, logger, false, true)

		mock.ExpectExec("INSERT INTO users").
			WithArgs("test").
			WillReturnResult(sqlmock.NewResult(1, 1))

		_, err := typedbDB.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "test")
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		// Verify query was NOT logged
		if len(logger.debugs) == 0 {
			t.Fatal("Expected Debug log even when LogQueries=false (should log without query)")
		}
		foundQuery := false
		for i := 0; i < len(logger.debugs[0].keyvals)-1; i += 2 {
			if logger.debugs[0].keyvals[i] == "query" {
				foundQuery = true
				break
			}
		}
		if foundQuery {
			t.Error("Expected 'query' key to be absent when LogQueries=false")
		}
	})
}

// TestLogArgsConfig verifies that WithLogArgs option controls argument logging
func TestLogArgsConfig(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	logger := &testLogger{}
	ctx := context.Background()

	t.Run("LogArgs=true logs arguments", func(t *testing.T) {
		logger.debugs = nil

		typedbDB := NewDBWithLoggerAndFlags(db, "test", 5*time.Second, logger, true, true)

		mock.ExpectExec("INSERT INTO users").
			WithArgs("test", "password123").
			WillReturnResult(sqlmock.NewResult(1, 1))

		_, err := typedbDB.Exec(ctx, "INSERT INTO users (name, password) VALUES ($1, $2)", "test", "password123")
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		// Verify args were logged
		if len(logger.debugs) == 0 {
			t.Fatal("Expected Debug log")
		}
		foundArgs := false
		for i := 0; i < len(logger.debugs[0].keyvals)-1; i += 2 {
			if logger.debugs[0].keyvals[i] == "args" {
				foundArgs = true
				break
			}
		}
		if !foundArgs {
			t.Error("Expected 'args' key in log when LogArgs=true")
		}
	})

	t.Run("LogArgs=false does not log arguments", func(t *testing.T) {
		logger.debugs = nil

		typedbDB := NewDBWithLoggerAndFlags(db, "test", 5*time.Second, logger, true, false)

		mock.ExpectExec("INSERT INTO users").
			WithArgs("test", "password123").
			WillReturnResult(sqlmock.NewResult(1, 1))

		_, err := typedbDB.Exec(ctx, "INSERT INTO users (name, password) VALUES ($1, $2)", "test", "password123")
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		// Verify args were NOT logged
		if len(logger.debugs) == 0 {
			t.Fatal("Expected Debug log")
		}
		foundArgs := false
		for i := 0; i < len(logger.debugs[0].keyvals)-1; i += 2 {
			if logger.debugs[0].keyvals[i] == "args" {
				foundArgs = true
				break
			}
		}
		if foundArgs {
			t.Error("Expected 'args' key to be absent when LogArgs=false")
		}
		// But query should still be logged
		foundQuery := false
		for i := 0; i < len(logger.debugs[0].keyvals)-1; i += 2 {
			if logger.debugs[0].keyvals[i] == "query" {
				foundQuery = true
				break
			}
		}
		if !foundQuery {
			t.Error("Expected 'query' key in log when LogQueries=true")
		}
	})
}

// TestLogConfigOptions verifies that WithLogQueries and WithLogArgs options work correctly
func TestLogConfigOptions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	logger := &testLogger{}
	ctx := context.Background()

	t.Run("both disabled - no query or args logged", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		typedbDB := NewDBWithLoggerAndFlags(db, "test", 5*time.Second, logger, false, false)

		mock.ExpectExec("INSERT INTO users").
			WithArgs("test", "password123").
			WillReturnResult(sqlmock.NewResult(1, 1))

		_, err := typedbDB.Exec(ctx, "INSERT INTO users (name, password) VALUES ($1, $2)", "test", "password123")
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		// Verify neither query nor args were logged
		if len(logger.debugs) == 0 {
			t.Fatal("Expected Debug log even when both disabled (should log without query/args)")
		}
		foundQuery := false
		foundArgs := false
		for i := 0; i < len(logger.debugs[0].keyvals)-1; i += 2 {
			if logger.debugs[0].keyvals[i] == "query" {
				foundQuery = true
			}
			if logger.debugs[0].keyvals[i] == "args" {
				foundArgs = true
			}
		}
		if foundQuery {
			t.Error("Expected 'query' key to be absent when LogQueries=false")
		}
		if foundArgs {
			t.Error("Expected 'args' key to be absent when LogArgs=false")
		}
	})

	t.Run("error logging respects flags", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		typedbDB := NewDBWithLoggerAndFlags(db, "test", 5*time.Second, logger, false, false)

		expectedErr := errors.New("database error")
		mock.ExpectExec("INSERT INTO users").
			WithArgs("test").
			WillReturnError(expectedErr)

		_, err := typedbDB.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "test")
		if err != expectedErr {
			t.Fatalf("Expected error %v, got %v", expectedErr, err)
		}

		// Verify error log does not contain query or args
		if len(logger.errors) == 0 {
			t.Fatal("Expected Error log")
		}
		foundQuery := false
		foundArgs := false
		for i := 0; i < len(logger.errors[0].keyvals)-1; i += 2 {
			if logger.errors[0].keyvals[i] == "query" {
				foundQuery = true
			}
			if logger.errors[0].keyvals[i] == "args" {
				foundArgs = true
			}
		}
		if foundQuery {
			t.Error("Expected 'query' key to be absent in error log when LogQueries=false")
		}
		if foundArgs {
			t.Error("Expected 'args' key to be absent in error log when LogArgs=false")
		}
		// But error should still be logged
		foundError := false
		for i := 0; i < len(logger.errors[0].keyvals)-1; i += 2 {
			if logger.errors[0].keyvals[i] == "error" {
				foundError = true
				break
			}
		}
		if !foundError {
			t.Error("Expected 'error' key in error log")
		}
	})
}

// TestContextLoggingOverrides verifies that context-based logging overrides work correctly
func TestContextLoggingOverrides(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	logger := &testLogger{}
	ctx := context.Background()
	typedbDB := NewDBWithLoggerAndFlags(db, "test", 5*time.Second, logger, true, true)

	t.Run("WithNoLogging - disables all logging", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		mock.ExpectExec("INSERT INTO users").
			WithArgs("test", "password123").
			WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := WithNoLogging(ctx)
		_, err := typedbDB.Exec(ctx, "INSERT INTO users (name, password) VALUES ($1, $2)", "test", "password123")
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		// Should still log basic message but without query/args
		if len(logger.debugs) == 0 {
			t.Fatal("Expected Debug log even when logging disabled")
		}
		// Check that query and args are not in the log
		foundQuery := false
		foundArgs := false
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "query" {
					foundQuery = true
				}
				if entry.keyvals[i] == "args" {
					foundArgs = true
				}
			}
		}
		if foundQuery {
			t.Error("Expected 'query' key to be absent when WithNoLogging is used")
		}
		if foundArgs {
			t.Error("Expected 'args' key to be absent when WithNoLogging is used")
		}
	})

	t.Run("WithNoQueryLogging - disables query logging only", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		mock.ExpectExec("INSERT INTO users").
			WithArgs("test", "password123").
			WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := WithNoQueryLogging(ctx)
		_, err := typedbDB.Exec(ctx, "INSERT INTO users (name, password) VALUES ($1, $2)", "test", "password123")
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		// Query should not be logged, but args should be
		foundQuery := false
		foundArgs := false
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "query" {
					foundQuery = true
				}
				if entry.keyvals[i] == "args" {
					foundArgs = true
				}
			}
		}
		if foundQuery {
			t.Error("Expected 'query' key to be absent when WithNoQueryLogging is used")
		}
		if !foundArgs {
			t.Error("Expected 'args' key to be present when WithNoQueryLogging is used (only query disabled)")
		}
	})

	t.Run("WithNoArgLogging - disables argument logging only", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		mock.ExpectExec("INSERT INTO users").
			WithArgs("test", "password123").
			WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := WithNoArgLogging(ctx)
		_, err := typedbDB.Exec(ctx, "INSERT INTO users (name, password) VALUES ($1, $2)", "test", "password123")
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		// Query should be logged, but args should not
		foundQuery := false
		foundArgs := false
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "query" {
					foundQuery = true
				}
				if entry.keyvals[i] == "args" {
					foundArgs = true
				}
			}
		}
		if !foundQuery {
			t.Error("Expected 'query' key to be present when WithNoArgLogging is used (only args disabled)")
		}
		if foundArgs {
			t.Error("Expected 'args' key to be absent when WithNoArgLogging is used")
		}
	})
}

// UserWithNolog is a test model with nolog tag
type UserWithNolog struct {
	Model
	ID       int    `db:"id" load:"primary"`
	Name     string `db:"name"`
	Password string `db:"password" nolog:"true"`
	Email    string `db:"email"`
}

func (u *UserWithNolog) TableName() string {
	return "users"
}

func (u *UserWithNolog) QueryByID() string {
	return "SELECT id, name, email FROM users WHERE id = $1"
}

// TestNologTagMasking verifies that nolog struct tags mask arguments in logs
func TestNologTagMasking(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	logger := &testLogger{}
	ctx := context.Background()

	RegisterModel[*UserWithNolog]()

	user := &UserWithNolog{
		Name:     "John",
		Password: "secret123",
		Email:    "john@example.com",
	}

	typedbDB := NewDBWithLoggerAndFlags(db, "postgres", 5*time.Second, logger, true, true)

	t.Run("Insert masks nolog fields", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		mock.ExpectQuery(`INSERT INTO "users" \("name", "password", "email"\) VALUES \(\$1, \$2, \$3\) RETURNING "id"`).
			WithArgs("John", "secret123", "john@example.com").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		err := Insert(ctx, typedbDB, user)
		if err != nil {
			t.Fatalf("Insert failed: %v", err)
		}

		// Check that password argument is masked in logs
		foundArgs := false
		var loggedArgs []any
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "args" {
					foundArgs = true
					if args, ok := entry.keyvals[i+1].([]any); ok {
						loggedArgs = args
					}
					break
				}
			}
			if foundArgs {
				break
			}
		}

		if !foundArgs {
			t.Fatal("Expected 'args' key in log")
		}

		// Password should be masked (at index 1: Name, Password, Email)
		if len(loggedArgs) < 3 {
			t.Fatalf("Expected at least 3 arguments, got %d", len(loggedArgs))
		}
		if loggedArgs[1] != "[REDACTED]" {
			t.Errorf("Expected password to be masked, got %v", loggedArgs[1])
		}
		// Other fields should not be masked
		if loggedArgs[0] != "John" {
			t.Errorf("Expected name to be 'John', got %v", loggedArgs[0])
		}
		if loggedArgs[2] != "john@example.com" {
			t.Errorf("Expected email to be 'john@example.com', got %v", loggedArgs[2])
		}
	})

	t.Run("Update masks nolog fields", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		user.ID = 1
		mock.ExpectExec(`UPDATE "users" SET "name" = \$1, "password" = \$2, "email" = \$3 WHERE "id" = \$4`).
			WithArgs("John", "secret123", "john@example.com", 1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := Update(ctx, typedbDB, user)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		// Check that password argument is masked in logs
		foundArgs := false
		var loggedArgs []any
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "args" {
					foundArgs = true
					if args, ok := entry.keyvals[i+1].([]any); ok {
						loggedArgs = args
					}
					break
				}
			}
			if foundArgs {
				break
			}
		}

		if !foundArgs {
			t.Fatal("Expected 'args' key in log")
		}

		// Password should be masked (at index 1: Name, Password, Email, then ID)
		if len(loggedArgs) < 3 {
			t.Fatalf("Expected at least 3 arguments, got %d", len(loggedArgs))
		}
		if loggedArgs[1] != "[REDACTED]" {
			t.Errorf("Expected password to be masked, got %v", loggedArgs[1])
		}
	})
}

// UserWithNologPK is a test model with nolog tag on primary key
type UserWithNologPK struct {
	Model
	ID       int    `db:"id" load:"primary" nolog:"true"`
	Name     string `db:"name"`
	Email    string `db:"email"`
}

func (u *UserWithNologPK) TableName() string {
	return "users"
}

func (u *UserWithNologPK) QueryByID() string {
	return "SELECT id, name, email FROM users WHERE id = $1"
}

// UserWithNologEmail is a test model with nolog tag on email field
type UserWithNologEmail struct {
	Model
	ID       int    `db:"id" load:"primary"`
	Name     string `db:"name"`
	Email    string `db:"email" nolog:"true"`
}

func (u *UserWithNologEmail) TableName() string {
	return "users"
}

func (u *UserWithNologEmail) QueryByID() string {
	return "SELECT id, name, email FROM users WHERE id = $1"
}

func (u *UserWithNologEmail) QueryByEmail() string {
	return "SELECT id, name, email FROM users WHERE email = $1"
}

// UserPostWithNolog is a test model with composite key where one field has nolog tag
type UserPostWithNolog struct {
	Model
	UserID   int    `db:"user_id" load:"composite:userpost"`
	PostID   int    `db:"post_id" load:"composite:userpost" nolog:"true"`
	Title    string `db:"title"`
}

func (u *UserPostWithNolog) TableName() string {
	return "user_posts"
}

func (u *UserPostWithNolog) QueryByPostIDUserID() string {
	return "SELECT user_id, post_id, title FROM user_posts WHERE post_id = $1 AND user_id = $2"
}

// TestNologTagMaskingInLoad verifies that nolog struct tags mask arguments in Load operations
func TestNologTagMaskingInLoad(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	logger := &testLogger{}
	ctx := context.Background()

	RegisterModel[*UserWithNologPK]()

	typedbDB := NewDBWithLoggerAndFlags(db, "postgres", 5*time.Second, logger, true, true)

	t.Run("Load masks nolog primary key field", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		user := &UserWithNologPK{ID: 123}

		mock.ExpectQuery("SELECT id, name, email FROM users WHERE id = \\$1").
			WithArgs(123).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(123, "John", "john@example.com"))

		err := Load(ctx, typedbDB, user)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		// Check that ID argument is masked in logs
		foundArgs := false
		var loggedArgs []any
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "args" {
					foundArgs = true
					if args, ok := entry.keyvals[i+1].([]any); ok {
						loggedArgs = args
					}
					break
				}
			}
			if foundArgs {
				break
			}
		}

		if !foundArgs {
			t.Fatal("Expected 'args' key in log")
		}

		// ID should be masked (at index 0)
		if len(loggedArgs) < 1 {
			t.Fatalf("Expected at least 1 argument, got %d", len(loggedArgs))
		}
		if loggedArgs[0] != "[REDACTED]" {
			t.Errorf("Expected ID to be masked, got %v", loggedArgs[0])
		}
	})

	t.Run("LoadByField masks nolog field", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		RegisterModel[*UserWithNologEmail]()

		user := &UserWithNologEmail{Email: "secret@example.com"}

		mock.ExpectQuery("SELECT id, name, email FROM users WHERE email = \\$1").
			WithArgs("secret@example.com").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(1, "John", "secret@example.com"))

		err := LoadByField(ctx, typedbDB, user, "Email")
		if err != nil {
			t.Fatalf("LoadByField failed: %v", err)
		}

		// Check that Email argument is masked in logs
		foundArgs := false
		var loggedArgs []any
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "args" {
					foundArgs = true
					if args, ok := entry.keyvals[i+1].([]any); ok {
						loggedArgs = args
					}
					break
				}
			}
			if foundArgs {
				break
			}
		}

		if !foundArgs {
			t.Fatal("Expected 'args' key in log")
		}

		// Email should be masked (at index 0)
		if len(loggedArgs) < 1 {
			t.Fatalf("Expected at least 1 argument, got %d", len(loggedArgs))
		}
		if loggedArgs[0] != "[REDACTED]" {
			t.Errorf("Expected Email to be masked, got %v", loggedArgs[0])
		}
	})

	t.Run("LoadByComposite masks nolog fields", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		RegisterModel[*UserPostWithNolog]()

		userPost := &UserPostWithNolog{UserID: 1, PostID: 2}

		mock.ExpectQuery("SELECT user_id, post_id, title FROM user_posts WHERE post_id = \\$1 AND user_id = \\$2").
			WithArgs(2, 1). // PostID first (alphabetically sorted), then UserID
			WillReturnRows(sqlmock.NewRows([]string{"user_id", "post_id", "title"}).
				AddRow(1, 2, "Test Post"))

		err := LoadByComposite(ctx, typedbDB, userPost, "userpost")
		if err != nil {
			t.Fatalf("LoadByComposite failed: %v", err)
		}

		// Check that PostID argument is masked in logs (at index 0, since PostID comes before UserID alphabetically)
		foundArgs := false
		var loggedArgs []any
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "args" {
					foundArgs = true
					if args, ok := entry.keyvals[i+1].([]any); ok {
						loggedArgs = args
					}
					break
				}
			}
			if foundArgs {
				break
			}
		}

		if !foundArgs {
			t.Fatal("Expected 'args' key in log")
		}

		// PostID should be masked (at index 0), UserID should not be masked (at index 1)
		if len(loggedArgs) < 2 {
			t.Fatalf("Expected at least 2 arguments, got %d", len(loggedArgs))
		}
		if loggedArgs[0] != "[REDACTED]" {
			t.Errorf("Expected PostID to be masked, got %v", loggedArgs[0])
		}
		if loggedArgs[1] == "[REDACTED]" {
			t.Errorf("Expected UserID to NOT be masked, got %v", loggedArgs[1])
		}
		if loggedArgs[1] != 1 {
			t.Errorf("Expected UserID to be 1, got %v", loggedArgs[1])
		}
	})
}