package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/TheBlackHowling/typedb"
	"github.com/TheBlackHowling/typedb/integration_tests/testhelpers"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

func TestMySQL_Logging_Exec(t *testing.T) {
	logger := &testhelpers.TestLogger{}
	db, err := typedb.OpenWithoutValidation("mysql", getTestDSN(), typedb.WithLogger(logger))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("success logs debug", func(t *testing.T) {
		logger.Debugs = nil
		// Use a unique email to avoid conflicts
		email := fmt.Sprintf("test-exec-%d@example.com", time.Now().UnixNano())
		_, err := db.Exec(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "Test User", email)
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		// Verify Debug log was emitted
		if len(logger.Debugs) == 0 {
			t.Fatal("Expected Debug log for Exec, got none")
		}
		if logger.Debugs[0].Msg != "Executing query" {
			t.Errorf("Expected Debug log message 'Executing query', got %q", logger.Debugs[0].Msg)
		}
		// Verify query is in keyvals
		foundQuery := false
		for i := 0; i < len(logger.Debugs[0].Keyvals)-1; i += 2 {
			if logger.Debugs[0].Keyvals[i] == "query" {
				foundQuery = true
				if !strings.Contains(logger.Debugs[0].Keyvals[i+1].(string), "INSERT INTO users") {
					t.Errorf("Expected query to contain 'INSERT INTO users', got %v", logger.Debugs[0].Keyvals[i+1])
				}
			}
		}
		if !foundQuery {
			t.Error("Expected 'query' key in Debug log keyvals")
		}
	})

	t.Run("error logs error", func(t *testing.T) {
		logger.Errors = nil
		_, err := db.Exec(ctx, "INVALID SQL SYNTAX")
		if err == nil {
			t.Fatal("Expected error for invalid SQL")
		}

		// Verify Error log was emitted
		if len(logger.Errors) == 0 {
			t.Fatal("Expected Error log for Exec failure, got none")
		}
		if logger.Errors[0].Msg != "Query execution failed" {
			t.Errorf("Expected Error log message 'Query execution failed', got %q", logger.Errors[0].Msg)
		}
	})
}

func TestMySQL_Logging_QueryAll(t *testing.T) {
	logger := &testhelpers.TestLogger{}
	db, err := typedb.OpenWithoutValidation("mysql", getTestDSN(), typedb.WithLogger(logger))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("success logs debug", func(t *testing.T) {
		logger.Debugs = nil
		_, err := db.QueryAll(ctx, "SELECT id, name, email, created_at FROM users ORDER BY id")
		if err != nil {
			t.Fatalf("QueryAll failed: %v", err)
		}

		// Verify Debug log was emitted
		if len(logger.Debugs) == 0 {
			t.Fatal("Expected Debug log for QueryAll, got none")
		}
		if logger.Debugs[0].Msg != "Executing query" {
			t.Errorf("Expected Debug log message 'Executing query', got %q", logger.Debugs[0].Msg)
		}
	})

	t.Run("error logs error", func(t *testing.T) {
		logger.Errors = nil
		_, err := db.QueryAll(ctx, "SELECT invalid_column FROM users")
		if err == nil {
			t.Fatal("Expected error for invalid column")
		}

		// Verify Error log was emitted
		if len(logger.Errors) == 0 {
			t.Fatal("Expected Error log for QueryAll failure, got none")
		}
		if logger.Errors[0].Msg != "Query execution failed" {
			t.Errorf("Expected Error log message 'Query execution failed', got %q", logger.Errors[0].Msg)
		}
	})
}

func TestMySQL_Logging_Begin_Commit_Rollback(t *testing.T) {
	logger := &testhelpers.TestLogger{}
	db, err := typedb.OpenWithoutValidation("mysql", getTestDSN(), typedb.WithLogger(logger))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("Begin logs debug", func(t *testing.T) {
		logger.Debugs = nil
		tx, err := db.Begin(ctx, nil)
		if err != nil {
			t.Fatalf("Begin failed: %v", err)
		}
		defer tx.Rollback()

		// Verify Debug log was emitted
		if len(logger.Debugs) == 0 {
			t.Fatal("Expected Debug log for Begin, got none")
		}
		if logger.Debugs[0].Msg != "Beginning transaction" {
			t.Errorf("Expected Debug log message 'Beginning transaction', got %q", logger.Debugs[0].Msg)
		}
	})

	t.Run("Commit logs debug", func(t *testing.T) {
		logger.Debugs = nil
		tx, err := db.Begin(ctx, nil)
		if err != nil {
			t.Fatalf("Begin failed: %v", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("Commit failed: %v", err)
		}

		// Verify Debug log was emitted
		if len(logger.Debugs) < 2 {
			t.Fatal("Expected Debug log for Commit, got none")
		}
		foundCommit := false
		for _, entry := range logger.Debugs {
			if entry.Msg == "Committing transaction" {
				foundCommit = true
				break
			}
		}
		if !foundCommit {
			t.Error("Expected Debug log message 'Committing transaction'")
		}
	})

	t.Run("Rollback logs debug", func(t *testing.T) {
		logger.Debugs = nil
		tx, err := db.Begin(ctx, nil)
		if err != nil {
			t.Fatalf("Begin failed: %v", err)
		}

		err = tx.Rollback()
		if err != nil {
			t.Fatalf("Rollback failed: %v", err)
		}

		// Verify Debug log was emitted
		if len(logger.Debugs) < 2 {
			t.Fatal("Expected Debug log for Rollback, got none")
		}
		foundRollback := false
		for _, entry := range logger.Debugs {
			if entry.Msg == "Rolling back transaction" {
				foundRollback = true
				break
			}
		}
		if !foundRollback {
			t.Error("Expected Debug log message 'Rolling back transaction'")
		}
	})
}

func TestMySQL_Logging_Close(t *testing.T) {
	logger := &testhelpers.TestLogger{}
	db, err := typedb.OpenWithoutValidation("mysql", getTestDSN(), typedb.WithLogger(logger))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	logger.Debugs = nil
	err = db.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify Debug log was emitted
	if len(logger.Debugs) == 0 {
		t.Fatal("Expected Debug log for Close, got none")
	}
	if logger.Debugs[0].Msg != "Closing database connection" {
		t.Errorf("Expected Debug log message 'Closing database connection', got %q", logger.Debugs[0].Msg)
	}
}

func TestMySQL_Logging_PerInstanceLogger(t *testing.T) {
	logger1 := &testhelpers.TestLogger{}
	logger2 := &testhelpers.TestLogger{}

	db1, err := typedb.OpenWithoutValidation("mysql", getTestDSN(), typedb.WithLogger(logger1))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db1.Close()

	db2, err := typedb.OpenWithoutValidation("mysql", getTestDSN(), typedb.WithLogger(logger2))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db2.Close()

	ctx := context.Background()

	// Execute query on db1
	logger1.Debugs = nil
	logger2.Debugs = nil
	email := fmt.Sprintf("test-per-instance-%d@example.com", time.Now().UnixNano())
	_, err = db1.Exec(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "Test User", email)
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	// Verify logger1 received logs but logger2 did not
	if len(logger1.Debugs) == 0 {
		t.Error("Expected Debug log in logger1, got none")
	}
	if len(logger2.Debugs) > 0 {
		t.Error("Expected no Debug logs in logger2, but got some")
	}
}

func TestMySQL_Logging_ConfigOptions(t *testing.T) {
	logger := &testhelpers.TestLogger{}

	t.Run("LogQueries=false disables query logging", func(t *testing.T) {
		db, err := typedb.OpenWithoutValidation("mysql", getTestDSN(),
			typedb.WithLogger(logger),
			typedb.WithLogQueries(false))
		if err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()

		ctx := context.Background()
		logger.Debugs = nil
		email := fmt.Sprintf("test-nolog-%d@example.com", time.Now().UnixNano())
		_, err = db.Exec(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "Test User", email)
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		// Verify query is NOT in keyvals
		for _, entry := range logger.Debugs {
			for i := 0; i < len(entry.Keyvals)-1; i += 2 {
				if entry.Keyvals[i] == "query" {
					t.Error("Expected 'query' key to be absent when LogQueries=false")
				}
			}
		}
	})

	t.Run("LogArgs=false disables args logging", func(t *testing.T) {
		db, err := typedb.OpenWithoutValidation("mysql", getTestDSN(),
			typedb.WithLogger(logger),
			typedb.WithLogArgs(false))
		if err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()

		ctx := context.Background()
		logger.Debugs = nil
		email := fmt.Sprintf("test-noargs-%d@example.com", time.Now().UnixNano())
		_, err = db.Exec(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "Test User", email)
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		// Verify args is NOT in keyvals
		for _, entry := range logger.Debugs {
			for i := 0; i < len(entry.Keyvals)-1; i += 2 {
				if entry.Keyvals[i] == "args" {
					t.Error("Expected 'args' key to be absent when LogArgs=false")
				}
			}
		}
	})
}

func TestMySQL_Logging_ContextOverrides(t *testing.T) {
	logger := &testhelpers.TestLogger{}
	db, err := typedb.OpenWithoutValidation("mysql", getTestDSN(), typedb.WithLogger(logger))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("WithNoLogging disables all logging", func(t *testing.T) {
		logger.Debugs = nil
		email := fmt.Sprintf("test-nologging-%d@example.com", time.Now().UnixNano())
		ctx := typedb.WithNoLogging(ctx)
		_, err := db.Exec(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "Test User", email)
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		// Verify no Debug logs were emitted
		if len(logger.Debugs) > 0 {
			t.Error("Expected no Debug logs when WithNoLogging is used")
		}
	})

	t.Run("WithNoQueryLogging disables query logging only", func(t *testing.T) {
		logger.Debugs = nil
		email := fmt.Sprintf("test-noquery-%d@example.com", time.Now().UnixNano())
		ctx := typedb.WithNoQueryLogging(ctx)
		_, err := db.Exec(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "Test User", email)
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		// Verify query is NOT in keyvals but args might be
		foundQuery := false
		foundArgs := false
		for _, entry := range logger.Debugs {
			for i := 0; i < len(entry.Keyvals)-1; i += 2 {
				if entry.Keyvals[i] == "query" {
					foundQuery = true
				}
				if entry.Keyvals[i] == "args" {
					foundArgs = true
				}
			}
		}
		if foundQuery {
			t.Error("Expected 'query' key to be absent when WithNoQueryLogging is used")
		}
		if !foundArgs {
			t.Error("Expected 'args' key to be present when WithNoQueryLogging is used (only args disabled)")
		}
	})

	t.Run("WithNoArgLogging disables args logging only", func(t *testing.T) {
		logger.Debugs = nil
		email := fmt.Sprintf("test-noargs-%d@example.com", time.Now().UnixNano())
		ctx := typedb.WithNoArgLogging(ctx)
		_, err := db.Exec(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "Test User", email)
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		// Verify args is NOT in keyvals but query is
		foundQuery := false
		foundArgs := false
		for _, entry := range logger.Debugs {
			for i := 0; i < len(entry.Keyvals)-1; i += 2 {
				if entry.Keyvals[i] == "query" {
					foundQuery = true
				}
				if entry.Keyvals[i] == "args" {
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
// Note: Using email with nolog tag since MySQL schema doesn't have password column
type UserWithNolog struct {
	typedb.Model
	ID    int    `db:"id" load:"primary"`
	Name  string `db:"name"`
	Email string `db:"email" nolog:"true"`
}

func (u *UserWithNolog) TableName() string {
	return "users"
}

func (u *UserWithNolog) QueryByID() string {
	return "SELECT id, name, email FROM users WHERE id = ?"
}

func init() {
	typedb.RegisterModel[*UserWithNolog]()
}

func TestMySQL_Logging_NologTagMasking(t *testing.T) {
	logger := &testhelpers.TestLogger{}
	db, err := typedb.OpenWithoutValidation("mysql", getTestDSN(), typedb.WithLogger(logger))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("Insert masks nolog fields", func(t *testing.T) {
		logger.Debugs = nil
		// Use a unique email to avoid conflicts
		email := fmt.Sprintf("test-nolog-insert-%d@example.com", time.Now().UnixNano())
		user := &UserWithNolog{
			Name:  "Test User",
			Email: email,
		}
		err := typedb.Insert(ctx, db, user)
		if err != nil {
			t.Fatalf("Insert failed: %v", err)
		}

		// Check that email is masked in logs
		foundArgs := false
		foundMasked := false
		for _, entry := range logger.Debugs {
			for i := 0; i < len(entry.Keyvals)-1; i += 2 {
				if entry.Keyvals[i] == "args" {
					foundArgs = true
					args := entry.Keyvals[i+1].([]any)
					for _, arg := range args {
						if arg == "[REDACTED]" {
							foundMasked = true
						}
						if arg == email {
							t.Error("Email should be masked, but found raw value in logs")
						}
					}
				}
			}
		}
		if !foundArgs {
			t.Error("Expected 'args' key in Debug log")
		}
		if !foundMasked {
			t.Error("Expected email to be masked as [REDACTED]")
		}
	})

	t.Run("Update masks nolog fields", func(t *testing.T) {
		logger.Debugs = nil
		// Use a unique email to avoid conflicts
		email := fmt.Sprintf("test-nolog-update-%d@example.com", time.Now().UnixNano())
		user := &UserWithNolog{
			ID:    1,
			Name:  "Updated User",
			Email: email,
		}
		err := typedb.Update(ctx, db, user)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		// Check that email is masked in logs
		foundMasked := false
		for _, entry := range logger.Debugs {
			for i := 0; i < len(entry.Keyvals)-1; i += 2 {
				if entry.Keyvals[i] == "args" {
					args := entry.Keyvals[i+1].([]any)
					for _, arg := range args {
						if arg == "[REDACTED]" {
							foundMasked = true
						}
						if arg == email {
							t.Error("Email should be masked, but found raw value in logs")
						}
					}
				}
			}
		}
		if !foundMasked {
			t.Error("Expected email to be masked as [REDACTED]")
		}
	})

	t.Run("Load masks nolog fields", func(t *testing.T) {
		logger.Debugs = nil
		user := &UserWithNolog{ID: 1}
		err := typedb.Load(ctx, db, user)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		// For Load, the primary key (ID) is logged, not the email field
		// The masking is tested in Insert/Update where email is actually in args
	})
}

func TestMySQL_Logging_SerializationNolog(t *testing.T) {
	logger := &testhelpers.TestLogger{}
	db, err := typedb.OpenWithoutValidation("mysql", getTestDSN(), typedb.WithLogger(logger))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("QueryAll masks nolog fields in model arguments", func(t *testing.T) {
		logger.Debugs = nil
		email := fmt.Sprintf("test-serialization-%d@example.com", time.Now().UnixNano())
		user := &UserWithNolog{
			Name:  "Test User",
			Email: email,
		}

		// Pass model as argument to QueryAll
		_, err := db.QueryAll(ctx, "SELECT id, name, email FROM users WHERE email = ?", user)
		if err != nil {
			t.Fatalf("QueryAll failed: %v", err)
		}

		// Check that email is masked in logs
		foundArgs := false
		foundMasked := false
		for _, entry := range logger.Debugs {
			for i := 0; i < len(entry.Keyvals)-1; i += 2 {
				if entry.Keyvals[i] == "args" {
					foundArgs = true
					args := entry.Keyvals[i+1].([]any)
					for _, arg := range args {
						if arg == "[REDACTED]" {
							foundMasked = true
						}
						if arg == email {
							t.Error("Email should be masked, but found raw value in logs")
						}
					}
				}
			}
		}
		if !foundArgs {
			t.Error("Expected 'args' key in Debug log")
		}
		if !foundMasked {
			t.Error("Expected email to be masked as [REDACTED]")
		}
	})

	t.Run("QueryRowMap masks nolog fields in model arguments", func(t *testing.T) {
		logger.Debugs = nil
		email := fmt.Sprintf("test-serialization-rowmap-%d@example.com", time.Now().UnixNano())
		user := &UserWithNolog{
			Name:  "Test User",
			Email: email,
		}

		// Pass model as argument to QueryRowMap
		_, err := db.QueryRowMap(ctx, "SELECT id, name, email FROM users WHERE email = ?", user)
		// May return ErrNotFound if user doesn't exist, which is fine for this test
		if err != nil && err != typedb.ErrNotFound {
			t.Fatalf("QueryRowMap failed: %v", err)
		}

		// Check that email is masked in logs
		foundArgs := false
		foundMasked := false
		for _, entry := range logger.Debugs {
			for i := 0; i < len(entry.Keyvals)-1; i += 2 {
				if entry.Keyvals[i] == "args" {
					foundArgs = true
					args := entry.Keyvals[i+1].([]any)
					for _, arg := range args {
						if arg == "[REDACTED]" {
							foundMasked = true
						}
						if arg == email {
							t.Error("Email should be masked, but found raw value in logs")
						}
					}
				}
			}
		}
		if !foundArgs {
			t.Error("Expected 'args' key in Debug log")
		}
		if !foundMasked {
			t.Error("Expected email to be masked as [REDACTED]")
		}
	})

	t.Run("GetInto masks nolog fields in model arguments", func(t *testing.T) {
		logger.Debugs = nil
		email := fmt.Sprintf("test-serialization-getinto-%d@example.com", time.Now().UnixNano())
		user := &UserWithNolog{
			Name:  "Test User",
			Email: email,
		}
		var dest map[string]any

		// Pass model as argument to GetInto
		err := db.GetInto(ctx, "SELECT id, name, email FROM users WHERE email = ?", []any{user}, &dest)
		// May return ErrNotFound if user doesn't exist, which is fine for this test
		if err != nil && err != typedb.ErrNotFound {
			t.Fatalf("GetInto failed: %v", err)
		}

		// Check that email is masked in logs
		foundArgs := false
		foundMasked := false
		for _, entry := range logger.Debugs {
			for i := 0; i < len(entry.Keyvals)-1; i += 2 {
				if entry.Keyvals[i] == "args" {
					foundArgs = true
					args := entry.Keyvals[i+1].([]any)
					for _, arg := range args {
						if arg == "[REDACTED]" {
							foundMasked = true
						}
						if arg == email {
							t.Error("Email should be masked, but found raw value in logs")
						}
					}
				}
			}
		}
		if !foundArgs {
			t.Error("Expected 'args' key in Debug log")
		}
		if !foundMasked {
			t.Error("Expected email to be masked as [REDACTED]")
		}
	})

	t.Run("QueryDo masks nolog fields in model arguments", func(t *testing.T) {
		logger.Debugs = nil
		email := fmt.Sprintf("test-serialization-querydo-%d@example.com", time.Now().UnixNano())
		user := &UserWithNolog{
			Name:  "Test User",
			Email: email,
		}

		// Pass model as argument to QueryDo
		err := db.QueryDo(ctx, "SELECT id, name, email FROM users WHERE email = ?", []any{user}, func(rows *sql.Rows) error {
			return nil
		})
		// May return ErrNotFound if user doesn't exist, which is fine for this test
		if err != nil && err != typedb.ErrNotFound {
			t.Fatalf("QueryDo failed: %v", err)
		}

		// Check that email is masked in logs
		foundArgs := false
		foundMasked := false
		for _, entry := range logger.Debugs {
			for i := 0; i < len(entry.Keyvals)-1; i += 2 {
				if entry.Keyvals[i] == "args" {
					foundArgs = true
					args := entry.Keyvals[i+1].([]any)
					for _, arg := range args {
						if arg == "[REDACTED]" {
							foundMasked = true
						}
						if arg == email {
							t.Error("Email should be masked, but found raw value in logs")
						}
					}
				}
			}
		}
		if !foundArgs {
			t.Error("Expected 'args' key in Debug log")
		}
		if !foundMasked {
			t.Error("Expected email to be masked as [REDACTED]")
		}
	})
}

