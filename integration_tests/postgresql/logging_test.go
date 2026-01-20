package main

import (
	"context"
	"strings"
	"testing"

	"github.com/TheBlackHowling/typedb"
	_ "github.com/lib/pq" // PostgreSQL driver
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

func TestPostgreSQL_Logging_Exec(t *testing.T) {
	logger := &testLogger{}
	db, err := typedb.OpenWithoutValidation("postgres", getTestDSN(), typedb.WithLogger(logger))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("success logs debug", func(t *testing.T) {
		logger.debugs = nil // Reset logs
		_, err := db.Exec(ctx, "INSERT INTO users (name, email) VALUES ($1, $2)", "Test User", "test@example.com")
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
		// Verify query is in keyvals
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
		// Use invalid SQL to trigger an error
		_, err := db.Exec(ctx, "INSERT INTO nonexistent_table (name) VALUES ($1)", "test")
		if err == nil {
			t.Fatal("Expected error for invalid SQL, got nil")
		}

		// Verify Error log was emitted
		if len(logger.errors) == 0 {
			t.Fatal("Expected Error log for Exec failure, got none")
		}
		if logger.errors[0].msg != "Query execution failed" {
			t.Errorf("Expected Error log message 'Query execution failed', got %q", logger.errors[0].msg)
		}
	})
}

func TestPostgreSQL_Logging_QueryAll(t *testing.T) {
	logger := &testLogger{}
	db, err := typedb.OpenWithoutValidation("postgres", getTestDSN(), typedb.WithLogger(logger))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("success logs debug", func(t *testing.T) {
		logger.debugs = nil // Reset logs
		_, err := db.QueryAll(ctx, "SELECT id, name, email, created_at FROM users ORDER BY id LIMIT 1")
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
		// Use invalid SQL to trigger an error
		_, err := db.QueryAll(ctx, "SELECT invalid_column FROM users")
		if err == nil {
			t.Fatal("Expected error for invalid SQL, got nil")
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

func TestPostgreSQL_Logging_Begin_Commit_Rollback(t *testing.T) {
	logger := &testLogger{}
	db, err := typedb.OpenWithoutValidation("postgres", getTestDSN(), typedb.WithLogger(logger))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("begin logs debug", func(t *testing.T) {
		logger.debugs = nil // Reset logs
		tx, err := db.Begin(ctx, nil)
		if err != nil {
			t.Fatalf("Begin failed: %v", err)
		}
		defer tx.Rollback()

		// Verify Debug log was emitted
		if len(logger.debugs) == 0 {
			t.Fatal("Expected Debug log for Begin, got none")
		}
		if logger.debugs[0].msg != "Beginning transaction" {
			t.Errorf("Expected Debug log message 'Beginning transaction', got %q", logger.debugs[0].msg)
		}
	})

	t.Run("commit logs info", func(t *testing.T) {
		logger.infos = nil // Reset logs
		tx, err := db.Begin(ctx, nil)
		if err != nil {
			t.Fatalf("Begin failed: %v", err)
		}

		err = tx.Commit()
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

	t.Run("rollback logs info", func(t *testing.T) {
		logger.infos = nil // Reset logs
		tx, err := db.Begin(ctx, nil)
		if err != nil {
			t.Fatalf("Begin failed: %v", err)
		}

		err = tx.Rollback()
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

func TestPostgreSQL_Logging_Close(t *testing.T) {
	logger := &testLogger{}
	db, err := typedb.OpenWithoutValidation("postgres", getTestDSN(), typedb.WithLogger(logger))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	t.Run("close logs info", func(t *testing.T) {
		logger.infos = nil // Reset logs
		err := db.Close()
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

func TestPostgreSQL_Logging_PerInstanceLogger(t *testing.T) {
	globalLogger := &testLogger{}
	instanceLogger := &testLogger{}

	// Set global logger
	typedb.SetLogger(globalLogger)

	// Create DB with per-instance logger
	db, err := typedb.OpenWithoutValidation("postgres", getTestDSN(), typedb.WithLogger(instanceLogger))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	_, err = db.Exec(ctx, "INSERT INTO users (name, email) VALUES ($1, $2)", "Test User", "test2@example.com")
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
