package typedb

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestExecGlobalLoggingConfig verifies that global logging config options work for Exec
func TestExecGlobalLoggingConfig(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	logger := &testLogger{}
	ctx := context.Background()
	query := "INSERT INTO users (name, email) VALUES ($1, $2)"
	args := []any{"John", "john@example.com"}

	t.Run("LogQueries=false disables query logging", func(t *testing.T) {
		logger.debugs = nil
		typedbDB := NewDBWithLoggerAndFlags(db, "test", 5*time.Second, logger, false, true)

		mock.ExpectExec("INSERT INTO users").
			WithArgs("John", "john@example.com").
			WillReturnResult(sqlmock.NewResult(1, 1))

		_, err := typedbDB.Exec(ctx, query, args...)
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		// Should log message but without query
		if len(logger.debugs) == 0 {
			t.Fatal("Expected Debug log even when LogQueries=false")
		}
		foundQuery := false
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "query" {
					foundQuery = true
				}
			}
		}
		if foundQuery {
			t.Error("Expected 'query' key to be absent when LogQueries=false")
		}
	})

	t.Run("LogArgs=false disables argument logging", func(t *testing.T) {
		logger.debugs = nil
		typedbDB := NewDBWithLoggerAndFlags(db, "test", 5*time.Second, logger, true, false)

		mock.ExpectExec("INSERT INTO users").
			WithArgs("John", "john@example.com").
			WillReturnResult(sqlmock.NewResult(1, 1))

		_, err := typedbDB.Exec(ctx, query, args...)
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		// Should log query but without args
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
			t.Error("Expected 'query' key to be present when LogQueries=true")
		}
		if foundArgs {
			t.Error("Expected 'args' key to be absent when LogArgs=false")
		}
	})

	t.Run("LogQueries=false and LogArgs=false disables both", func(t *testing.T) {
		logger.debugs = nil
		typedbDB := NewDBWithLoggerAndFlags(db, "test", 5*time.Second, logger, false, false)

		mock.ExpectExec("INSERT INTO users").
			WithArgs("John", "john@example.com").
			WillReturnResult(sqlmock.NewResult(1, 1))

		_, err := typedbDB.Exec(ctx, query, args...)
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		// Should log message but without query or args
		if len(logger.debugs) == 0 {
			t.Fatal("Expected Debug log even when both disabled")
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
			t.Error("Expected 'query' key to be absent when LogQueries=false")
		}
		if foundArgs {
			t.Error("Expected 'args' key to be absent when LogArgs=false")
		}
	})
}

// TestQueryAllGlobalLoggingConfig verifies that global logging config options work for QueryAll
func TestQueryAllGlobalLoggingConfig(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	logger := &testLogger{}
	ctx := context.Background()
	query := "SELECT id, name, email FROM users WHERE email = $1"
	args := []any{"john@example.com"}

	t.Run("LogQueries=false disables query logging", func(t *testing.T) {
		logger.debugs = nil
		typedbDB := NewDBWithLoggerAndFlags(db, "test", 5*time.Second, logger, false, true)

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs("john@example.com").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(1, "John", "john@example.com"))

		_, err := typedbDB.QueryAll(ctx, query, args...)
		if err != nil {
			t.Fatalf("QueryAll failed: %v", err)
		}

		// Should log message but without query
		if len(logger.debugs) == 0 {
			t.Fatal("Expected Debug log even when LogQueries=false")
		}
		foundQuery := false
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "query" {
					foundQuery = true
				}
			}
		}
		if foundQuery {
			t.Error("Expected 'query' key to be absent when LogQueries=false")
		}
	})

	t.Run("LogArgs=false disables argument logging", func(t *testing.T) {
		logger.debugs = nil
		typedbDB := NewDBWithLoggerAndFlags(db, "test", 5*time.Second, logger, true, false)

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs("john@example.com").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(1, "John", "john@example.com"))

		_, err := typedbDB.QueryAll(ctx, query, args...)
		if err != nil {
			t.Fatalf("QueryAll failed: %v", err)
		}

		// Should log query but without args
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
			t.Error("Expected 'query' key to be present when LogQueries=true")
		}
		if foundArgs {
			t.Error("Expected 'args' key to be absent when LogArgs=false")
		}
	})
}

// TestQueryRowMapGlobalLoggingConfig verifies that global logging config options work for QueryRowMap
func TestQueryRowMapGlobalLoggingConfig(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	logger := &testLogger{}
	ctx := context.Background()
	query := "SELECT id, name, email FROM users WHERE id = $1"
	args := []any{123}

	t.Run("LogQueries=false disables query logging", func(t *testing.T) {
		logger.debugs = nil
		typedbDB := NewDBWithLoggerAndFlags(db, "test", 5*time.Second, logger, false, true)

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs(123).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(123, "John", "john@example.com"))

		_, err := typedbDB.QueryRowMap(ctx, query, args...)
		if err != nil {
			t.Fatalf("QueryRowMap failed: %v", err)
		}

		// Should log message but without query
		if len(logger.debugs) == 0 {
			t.Fatal("Expected Debug log even when LogQueries=false")
		}
		foundQuery := false
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "query" {
					foundQuery = true
				}
			}
		}
		if foundQuery {
			t.Error("Expected 'query' key to be absent when LogQueries=false")
		}
	})

	t.Run("LogArgs=false disables argument logging", func(t *testing.T) {
		logger.debugs = nil
		typedbDB := NewDBWithLoggerAndFlags(db, "test", 5*time.Second, logger, true, false)

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs(123).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(123, "John", "john@example.com"))

		_, err := typedbDB.QueryRowMap(ctx, query, args...)
		if err != nil {
			t.Fatalf("QueryRowMap failed: %v", err)
		}

		// Should log query but without args
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
			t.Error("Expected 'query' key to be present when LogQueries=true")
		}
		if foundArgs {
			t.Error("Expected 'args' key to be absent when LogArgs=false")
		}
	})
}

// TestGetIntoGlobalLoggingConfig verifies that global logging config options work for GetInto
func TestGetIntoGlobalLoggingConfig(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	logger := &testLogger{}
	ctx := context.Background()
	query := "SELECT id, name, email FROM users WHERE id = $1"
	args := []any{123}
	var id int
	var name, email string

	t.Run("LogQueries=false disables query logging", func(t *testing.T) {
		logger.debugs = nil
		typedbDB := NewDBWithLoggerAndFlags(db, "test", 5*time.Second, logger, false, true)

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs(123).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(123, "John", "john@example.com"))

		err := typedbDB.GetInto(ctx, query, args, &id, &name, &email)
		if err != nil {
			t.Fatalf("GetInto failed: %v", err)
		}

		// Should log message but without query
		if len(logger.debugs) == 0 {
			t.Fatal("Expected Debug log even when LogQueries=false")
		}
		foundQuery := false
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "query" {
					foundQuery = true
				}
			}
		}
		if foundQuery {
			t.Error("Expected 'query' key to be absent when LogQueries=false")
		}
	})

	t.Run("LogArgs=false disables argument logging", func(t *testing.T) {
		logger.debugs = nil
		typedbDB := NewDBWithLoggerAndFlags(db, "test", 5*time.Second, logger, true, false)

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs(123).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(123, "John", "john@example.com"))

		err := typedbDB.GetInto(ctx, query, args, &id, &name, &email)
		if err != nil {
			t.Fatalf("GetInto failed: %v", err)
		}

		// Should log query but without args
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
			t.Error("Expected 'query' key to be present when LogQueries=true")
		}
		if foundArgs {
			t.Error("Expected 'args' key to be absent when LogArgs=false")
		}
	})
}

// TestQueryDoGlobalLoggingConfig verifies that global logging config options work for QueryDo
func TestQueryDoGlobalLoggingConfig(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	logger := &testLogger{}
	ctx := context.Background()
	query := "SELECT id, name, email FROM users WHERE email = $1"
	args := []any{"john@example.com"}

	t.Run("LogQueries=false disables query logging", func(t *testing.T) {
		logger.debugs = nil
		typedbDB := NewDBWithLoggerAndFlags(db, "test", 5*time.Second, logger, false, true)

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs("john@example.com").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(1, "John", "john@example.com"))

		err := typedbDB.QueryDo(ctx, query, args, func(rows *sql.Rows) error {
			return nil
		})
		if err != nil {
			t.Fatalf("QueryDo failed: %v", err)
		}

		// Should log message but without query
		if len(logger.debugs) == 0 {
			t.Fatal("Expected Debug log even when LogQueries=false")
		}
		foundQuery := false
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "query" {
					foundQuery = true
				}
			}
		}
		if foundQuery {
			t.Error("Expected 'query' key to be absent when LogQueries=false")
		}
	})

	t.Run("LogArgs=false disables argument logging", func(t *testing.T) {
		logger.debugs = nil
		typedbDB := NewDBWithLoggerAndFlags(db, "test", 5*time.Second, logger, true, false)

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs("john@example.com").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(1, "John", "john@example.com"))

		err := typedbDB.QueryDo(ctx, query, args, func(rows *sql.Rows) error {
			return nil
		})
		if err != nil {
			t.Fatalf("QueryDo failed: %v", err)
		}

		// Should log query but without args
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
			t.Error("Expected 'query' key to be present when LogQueries=true")
		}
		if foundArgs {
			t.Error("Expected 'args' key to be absent when LogArgs=false")
		}
	})
}
