package typedb

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

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
