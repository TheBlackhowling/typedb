package typedb

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestExecContextLoggingOverrides verifies that context-based logging overrides work for Exec
func TestExecContextLoggingOverrides(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	logger := &testLogger{}
	ctx := context.Background()
	typedbDB := NewDBWithLoggerAndFlags(db, "test", 5*time.Second, logger, true, true)
	query := "INSERT INTO users (name, email) VALUES ($1, $2)"
	args := []any{"John", "john@example.com"}

	t.Run("WithNoLogging disables all logging", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		mock.ExpectExec("INSERT INTO users").
			WithArgs("John", "john@example.com").
			WillReturnResult(sqlmock.NewResult(1, 1))

		newCtx := WithNoLogging(ctx)
		_, err := typedbDB.Exec(newCtx, query, args...)
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		// Should log message but without query/args
		if len(logger.debugs) == 0 {
			t.Fatal("Expected Debug log even when logging disabled")
		}
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

	t.Run("WithNoQueryLogging disables query logging only", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		mock.ExpectExec("INSERT INTO users").
			WithArgs("John", "john@example.com").
			WillReturnResult(sqlmock.NewResult(1, 1))

		newCtx := WithNoQueryLogging(ctx)
		_, err := typedbDB.Exec(newCtx, query, args...)
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

	t.Run("WithNoArgLogging disables argument logging only", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		mock.ExpectExec("INSERT INTO users").
			WithArgs("John", "john@example.com").
			WillReturnResult(sqlmock.NewResult(1, 1))

		newCtx := WithNoArgLogging(ctx)
		_, err := typedbDB.Exec(newCtx, query, args...)
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		// Args should not be logged, but query should be
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

// TestQueryAllContextLoggingOverrides verifies that context-based logging overrides work for QueryAll
func TestQueryAllContextLoggingOverrides(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	logger := &testLogger{}
	ctx := context.Background()
	typedbDB := NewDBWithLoggerAndFlags(db, "test", 5*time.Second, logger, true, true)
	query := "SELECT id, name, email FROM users WHERE email = $1"
	args := []any{"john@example.com"}

	t.Run("WithNoLogging disables all logging", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs("john@example.com").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(1, "John", "john@example.com"))

		newCtx := WithNoLogging(ctx)
		_, err := typedbDB.QueryAll(newCtx, query, args...)
		if err != nil {
			t.Fatalf("QueryAll failed: %v", err)
		}

		// Should log message but without query/args
		if len(logger.debugs) == 0 {
			t.Fatal("Expected Debug log even when logging disabled")
		}
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

	t.Run("WithNoQueryLogging disables query logging only", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs("john@example.com").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(1, "John", "john@example.com"))

		newCtx := WithNoQueryLogging(ctx)
		_, err := typedbDB.QueryAll(newCtx, query, args...)
		if err != nil {
			t.Fatalf("QueryAll failed: %v", err)
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

	t.Run("WithNoArgLogging disables argument logging only", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs("john@example.com").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(1, "John", "john@example.com"))

		newCtx := WithNoArgLogging(ctx)
		_, err := typedbDB.QueryAll(newCtx, query, args...)
		if err != nil {
			t.Fatalf("QueryAll failed: %v", err)
		}

		// Args should not be logged, but query should be
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

// TestQueryRowMapContextLoggingOverrides verifies that context-based logging overrides work for QueryRowMap
func TestQueryRowMapContextLoggingOverrides(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	logger := &testLogger{}
	ctx := context.Background()
	typedbDB := NewDBWithLoggerAndFlags(db, "test", 5*time.Second, logger, true, true)
	query := "SELECT id, name, email FROM users WHERE id = $1"
	args := []any{123}

	t.Run("WithNoLogging disables all logging", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs(123).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(123, "John", "john@example.com"))

		newCtx := WithNoLogging(ctx)
		_, err := typedbDB.QueryRowMap(newCtx, query, args...)
		if err != nil {
			t.Fatalf("QueryRowMap failed: %v", err)
		}

		// Should log message but without query/args
		if len(logger.debugs) == 0 {
			t.Fatal("Expected Debug log even when logging disabled")
		}
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

	t.Run("WithNoQueryLogging disables query logging only", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs(123).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(123, "John", "john@example.com"))

		newCtx := WithNoQueryLogging(ctx)
		_, err := typedbDB.QueryRowMap(newCtx, query, args...)
		if err != nil {
			t.Fatalf("QueryRowMap failed: %v", err)
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

	t.Run("WithNoArgLogging disables argument logging only", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs(123).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(123, "John", "john@example.com"))

		newCtx := WithNoArgLogging(ctx)
		_, err := typedbDB.QueryRowMap(newCtx, query, args...)
		if err != nil {
			t.Fatalf("QueryRowMap failed: %v", err)
		}

		// Args should not be logged, but query should be
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

// TestGetIntoContextLoggingOverrides verifies that context-based logging overrides work for GetInto
func TestGetIntoContextLoggingOverrides(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	logger := &testLogger{}
	ctx := context.Background()
	typedbDB := NewDBWithLoggerAndFlags(db, "test", 5*time.Second, logger, true, true)
	query := "SELECT id, name, email FROM users WHERE id = $1"
	args := []any{123}
	var id int
	var name, email string

	t.Run("WithNoLogging disables all logging", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs(123).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(123, "John", "john@example.com"))

		newCtx := WithNoLogging(ctx)
		err := typedbDB.GetInto(newCtx, query, args, &id, &name, &email)
		if err != nil {
			t.Fatalf("GetInto failed: %v", err)
		}

		// Should log message but without query/args
		if len(logger.debugs) == 0 {
			t.Fatal("Expected Debug log even when logging disabled")
		}
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

	t.Run("WithNoQueryLogging disables query logging only", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs(123).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(123, "John", "john@example.com"))

		newCtx := WithNoQueryLogging(ctx)
		err := typedbDB.GetInto(newCtx, query, args, &id, &name, &email)
		if err != nil {
			t.Fatalf("GetInto failed: %v", err)
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

	t.Run("WithNoArgLogging disables argument logging only", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs(123).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(123, "John", "john@example.com"))

		newCtx := WithNoArgLogging(ctx)
		err := typedbDB.GetInto(newCtx, query, args, &id, &name, &email)
		if err != nil {
			t.Fatalf("GetInto failed: %v", err)
		}

		// Args should not be logged, but query should be
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

// TestQueryDoContextLoggingOverrides verifies that context-based logging overrides work for QueryDo
func TestQueryDoContextLoggingOverrides(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	logger := &testLogger{}
	ctx := context.Background()
	typedbDB := NewDBWithLoggerAndFlags(db, "test", 5*time.Second, logger, true, true)
	query := "SELECT id, name, email FROM users WHERE email = $1"
	args := []any{"john@example.com"}

	t.Run("WithNoLogging disables all logging", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs("john@example.com").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(1, "John", "john@example.com"))

		newCtx := WithNoLogging(ctx)
		err := typedbDB.QueryDo(newCtx, query, args, func(rows *sql.Rows) error {
			return nil
		})
		if err != nil {
			t.Fatalf("QueryDo failed: %v", err)
		}

		// Should log message but without query/args
		if len(logger.debugs) == 0 {
			t.Fatal("Expected Debug log even when logging disabled")
		}
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

	t.Run("WithNoQueryLogging disables query logging only", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs("john@example.com").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(1, "John", "john@example.com"))

		newCtx := WithNoQueryLogging(ctx)
		err := typedbDB.QueryDo(newCtx, query, args, func(rows *sql.Rows) error {
			return nil
		})
		if err != nil {
			t.Fatalf("QueryDo failed: %v", err)
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

	t.Run("WithNoArgLogging disables argument logging only", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs("john@example.com").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(1, "John", "john@example.com"))

		newCtx := WithNoArgLogging(ctx)
		err := typedbDB.QueryDo(newCtx, query, args, func(rows *sql.Rows) error {
			return nil
		})
		if err != nil {
			t.Fatalf("QueryDo failed: %v", err)
		}

		// Args should not be logged, but query should be
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

// TestContextLoggingOverrides verifies that context-based logging overrides work correctly
func TestContextLoggingOverrides(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	logger := &testLogger{}
	ctx := context.Background()
	typedbDB := NewDBWithLoggerAndFlags(db, "test", 5*time.Second, logger, true, true)

	t.Run("WithNoLogging - disables all logging", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		mock.ExpectExec("INSERT INTO users").
			WithArgs("test", "password123").
			WillReturnResult(sqlmock.NewResult(1, 1))

		newCtx := WithNoLogging(ctx)
		_, err := typedbDB.Exec(newCtx, "INSERT INTO users (name, password) VALUES ($1, $2)", "test", "password123")
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

		newCtx := WithNoQueryLogging(ctx)
		_, err := typedbDB.Exec(newCtx, "INSERT INTO users (name, password) VALUES ($1, $2)", "test", "password123")
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

		newCtx := WithNoArgLogging(ctx)
		_, err := typedbDB.Exec(newCtx, "INSERT INTO users (name, password) VALUES ($1, $2)", "test", "password123")
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
