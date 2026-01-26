package typedb

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestExecHelperMasking verifies that execHelper correctly masks arguments
func TestExecHelperMasking(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	logger := &testLogger{}
	ctx := context.Background()
	query := "INSERT INTO users (name, email, password) VALUES ($1, $2, $3)"
	args := []any{"John", "john@example.com", "secret123"}

	t.Run("no mask indices - args logged as-is", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		mock.ExpectExec("INSERT INTO users").
			WithArgs("John", "john@example.com", "secret123").
			WillReturnResult(sqlmock.NewResult(1, 1))

		_, err := execHelper(ctx, db, logger, 5*time.Second, true, true, query, args...)
		if err != nil {
			t.Fatalf("execHelper failed: %v", err)
		}

		// Verify args are logged as-is
		foundArgs := false
		var loggedArgs []any
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "args" {
					foundArgs = true
					if logArgs, ok := entry.keyvals[i+1].([]any); ok {
						loggedArgs = logArgs
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
		if len(loggedArgs) != 3 {
			t.Fatalf("Expected 3 args, got %d", len(loggedArgs))
		}
		if loggedArgs[0] != "John" || loggedArgs[1] != "john@example.com" || loggedArgs[2] != "secret123" {
			t.Errorf("Expected args to be logged as-is, got %v", loggedArgs)
		}
	})

	t.Run("with mask indices - args masked", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		maskedCtx := WithMaskIndices(ctx, []int{2}) // Mask password at index 2

		mock.ExpectExec("INSERT INTO users").
			WithArgs("John", "john@example.com", "secret123").
			WillReturnResult(sqlmock.NewResult(1, 1))

		_, err := execHelper(maskedCtx, db, logger, 5*time.Second, true, true, query, args...)
		if err != nil {
			t.Fatalf("execHelper failed: %v", err)
		}

		// Verify args are masked
		foundArgs := false
		var loggedArgs []any
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "args" {
					foundArgs = true
					if logArgs, ok := entry.keyvals[i+1].([]any); ok {
						loggedArgs = logArgs
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
		if len(loggedArgs) != 3 {
			t.Fatalf("Expected 3 args, got %d", len(loggedArgs))
		}
		if loggedArgs[0] != "John" {
			t.Errorf("Expected first arg to be 'John', got %v", loggedArgs[0])
		}
		if loggedArgs[1] != "john@example.com" {
			t.Errorf("Expected second arg to be 'john@example.com', got %v", loggedArgs[1])
		}
		if loggedArgs[2] != "[REDACTED]" {
			t.Errorf("Expected third arg to be '[REDACTED]', got %v", loggedArgs[2])
		}
	})
}

// TestQueryAllHelperMasking verifies that queryAllHelper correctly masks arguments
func TestQueryAllHelperMasking(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	logger := &testLogger{}
	ctx := context.Background()
	query := "SELECT id, name, email FROM users WHERE email = $1 AND password = $2"
	args := []any{"john@example.com", "secret123"}

	t.Run("no mask indices - args logged as-is", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs("john@example.com", "secret123").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(1, "John", "john@example.com"))

		_, err := queryAllHelper(ctx, db, logger, 5*time.Second, true, true, query, args...)
		if err != nil {
			t.Fatalf("queryAllHelper failed: %v", err)
		}

		// Verify args are logged as-is
		foundArgs := false
		var loggedArgs []any
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "args" {
					foundArgs = true
					if logArgs, ok := entry.keyvals[i+1].([]any); ok {
						loggedArgs = logArgs
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
		if len(loggedArgs) != 2 {
			t.Fatalf("Expected 2 args, got %d", len(loggedArgs))
		}
		if loggedArgs[0] != "john@example.com" || loggedArgs[1] != "secret123" {
			t.Errorf("Expected args to be logged as-is, got %v", loggedArgs)
		}
	})

	t.Run("with mask indices - args masked", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		maskedCtx := WithMaskIndices(ctx, []int{1}) // Mask password at index 1

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs("john@example.com", "secret123").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(1, "John", "john@example.com"))

		_, err := queryAllHelper(maskedCtx, db, logger, 5*time.Second, true, true, query, args...)
		if err != nil {
			t.Fatalf("queryAllHelper failed: %v", err)
		}

		// Verify args are masked
		foundArgs := false
		var loggedArgs []any
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "args" {
					foundArgs = true
					if logArgs, ok := entry.keyvals[i+1].([]any); ok {
						loggedArgs = logArgs
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
		if len(loggedArgs) != 2 {
			t.Fatalf("Expected 2 args, got %d", len(loggedArgs))
		}
		if loggedArgs[0] != "john@example.com" {
			t.Errorf("Expected first arg to be 'john@example.com', got %v", loggedArgs[0])
		}
		if loggedArgs[1] != "[REDACTED]" {
			t.Errorf("Expected second arg to be '[REDACTED]', got %v", loggedArgs[1])
		}
	})
}

// TestQueryRowMapHelperMasking verifies that queryRowMapHelper correctly masks arguments
func TestQueryRowMapHelperMasking(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	logger := &testLogger{}
	ctx := context.Background()
	query := "SELECT id, name, email FROM users WHERE id = $1 AND password = $2"
	args := []any{123, "secret123"}

	t.Run("no mask indices - args logged as-is", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs(123, "secret123").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(123, "John", "john@example.com"))

		_, err := queryRowMapHelper(ctx, db, logger, 5*time.Second, true, true, query, args...)
		if err != nil {
			t.Fatalf("queryRowMapHelper failed: %v", err)
		}

		// Verify args are logged as-is
		foundArgs := false
		var loggedArgs []any
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "args" {
					foundArgs = true
					if logArgs, ok := entry.keyvals[i+1].([]any); ok {
						loggedArgs = logArgs
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
		if len(loggedArgs) != 2 {
			t.Fatalf("Expected 2 args, got %d", len(loggedArgs))
		}
		if loggedArgs[0] != 123 || loggedArgs[1] != "secret123" {
			t.Errorf("Expected args to be logged as-is, got %v", loggedArgs)
		}
	})

	t.Run("with mask indices - args masked", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		maskedCtx := WithMaskIndices(ctx, []int{1}) // Mask password at index 1

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs(123, "secret123").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(123, "John", "john@example.com"))

		_, err := queryRowMapHelper(maskedCtx, db, logger, 5*time.Second, true, true, query, args...)
		if err != nil {
			t.Fatalf("queryRowMapHelper failed: %v", err)
		}

		// Verify args are masked
		foundArgs := false
		var loggedArgs []any
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "args" {
					foundArgs = true
					if logArgs, ok := entry.keyvals[i+1].([]any); ok {
						loggedArgs = logArgs
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
		if len(loggedArgs) != 2 {
			t.Fatalf("Expected 2 args, got %d", len(loggedArgs))
		}
		if loggedArgs[0] != 123 {
			t.Errorf("Expected first arg to be 123, got %v", loggedArgs[0])
		}
		if loggedArgs[1] != "[REDACTED]" {
			t.Errorf("Expected second arg to be '[REDACTED]', got %v", loggedArgs[1])
		}
	})
}

// TestGetIntoHelperMasking verifies that getIntoHelper correctly masks arguments
func TestGetIntoHelperMasking(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	logger := &testLogger{}
	ctx := context.Background()
	query := "SELECT id, name, email FROM users WHERE email = $1 AND password = $2"
	args := []any{"john@example.com", "secret123"}
	var id int
	var name, email string

	t.Run("no mask indices - args logged as-is", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs("john@example.com", "secret123").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(1, "John", "john@example.com"))

		err := getIntoHelper(ctx, db, logger, 5*time.Second, true, true, query, args, &id, &name, &email)
		if err != nil {
			t.Fatalf("getIntoHelper failed: %v", err)
		}

		// Verify args are logged as-is
		foundArgs := false
		var loggedArgs []any
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "args" {
					foundArgs = true
					if logArgs, ok := entry.keyvals[i+1].([]any); ok {
						loggedArgs = logArgs
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
		if len(loggedArgs) != 2 {
			t.Fatalf("Expected 2 args, got %d", len(loggedArgs))
		}
		if loggedArgs[0] != "john@example.com" || loggedArgs[1] != "secret123" {
			t.Errorf("Expected args to be logged as-is, got %v", loggedArgs)
		}
	})

	t.Run("with mask indices - args masked", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		maskedCtx := WithMaskIndices(ctx, []int{1}) // Mask password at index 1

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs("john@example.com", "secret123").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(1, "John", "john@example.com"))

		err := getIntoHelper(maskedCtx, db, logger, 5*time.Second, true, true, query, args, &id, &name, &email)
		if err != nil {
			t.Fatalf("getIntoHelper failed: %v", err)
		}

		// Verify args are masked
		foundArgs := false
		var loggedArgs []any
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "args" {
					foundArgs = true
					if logArgs, ok := entry.keyvals[i+1].([]any); ok {
						loggedArgs = logArgs
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
		if len(loggedArgs) != 2 {
			t.Fatalf("Expected 2 args, got %d", len(loggedArgs))
		}
		if loggedArgs[0] != "john@example.com" {
			t.Errorf("Expected first arg to be 'john@example.com', got %v", loggedArgs[0])
		}
		if loggedArgs[1] != "[REDACTED]" {
			t.Errorf("Expected second arg to be '[REDACTED]', got %v", loggedArgs[1])
		}
	})
}

// TestQueryDoHelperMasking verifies that queryDoHelper correctly masks arguments
func TestQueryDoHelperMasking(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	logger := &testLogger{}
	ctx := context.Background()
	query := "SELECT id, name, email FROM users WHERE email = $1 AND password = $2"
	args := []any{"john@example.com", "secret123"}

	t.Run("no mask indices - args logged as-is", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs("john@example.com", "secret123").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(1, "John", "john@example.com"))

		scanCalled := false
		err := queryDoHelper(ctx, db, logger, 5*time.Second, true, true, query, args, func(rows *sql.Rows) error {
			scanCalled = true
			return nil
		})
		if err != nil {
			t.Fatalf("queryDoHelper failed: %v", err)
		}
		if !scanCalled {
			t.Error("Expected scan function to be called")
		}

		// Verify args are logged as-is
		foundArgs := false
		var loggedArgs []any
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "args" {
					foundArgs = true
					if logArgs, ok := entry.keyvals[i+1].([]any); ok {
						loggedArgs = logArgs
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
		if len(loggedArgs) != 2 {
			t.Fatalf("Expected 2 args, got %d", len(loggedArgs))
		}
		if loggedArgs[0] != "john@example.com" || loggedArgs[1] != "secret123" {
			t.Errorf("Expected args to be logged as-is, got %v", loggedArgs)
		}
	})

	t.Run("with mask indices - args masked", func(t *testing.T) {
		logger.debugs = nil
		logger.errors = nil

		maskedCtx := WithMaskIndices(ctx, []int{1}) // Mask password at index 1

		mock.ExpectQuery("SELECT id, name, email FROM users").
			WithArgs("john@example.com", "secret123").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(1, "John", "john@example.com"))

		scanCalled := false
		err := queryDoHelper(maskedCtx, db, logger, 5*time.Second, true, true, query, args, func(rows *sql.Rows) error {
			scanCalled = true
			return nil
		})
		if err != nil {
			t.Fatalf("queryDoHelper failed: %v", err)
		}
		if !scanCalled {
			t.Error("Expected scan function to be called")
		}

		// Verify args are masked
		foundArgs := false
		var loggedArgs []any
		for _, entry := range logger.debugs {
			for i := 0; i < len(entry.keyvals)-1; i += 2 {
				if entry.keyvals[i] == "args" {
					foundArgs = true
					if logArgs, ok := entry.keyvals[i+1].([]any); ok {
						loggedArgs = logArgs
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
		if len(loggedArgs) != 2 {
			t.Fatalf("Expected 2 args, got %d", len(loggedArgs))
		}
		if loggedArgs[0] != "john@example.com" {
			t.Errorf("Expected first arg to be 'john@example.com', got %v", loggedArgs[0])
		}
		if loggedArgs[1] != "[REDACTED]" {
			t.Errorf("Expected second arg to be '[REDACTED]', got %v", loggedArgs[1])
		}
	})
}
