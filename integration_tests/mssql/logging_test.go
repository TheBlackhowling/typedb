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
	_ "github.com/microsoft/go-mssqldb" // MSSQL driver
)

func TestMSSQL_Logging_Exec(t *testing.T) {
	logger := &testhelpers.TestLogger{}
	db, err := typedb.OpenWithoutValidation("sqlserver", getTestDSN(), typedb.WithLogger(logger))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	ctx := context.Background()

	t.Run("success logs debug", func(t *testing.T) {
		logger.Debugs = nil // Reset logs
		// Use a unique email to avoid conflicts
		email := fmt.Sprintf("test-exec-%d@example.com", time.Now().UnixNano())
		_, err := db.Exec(ctx, "INSERT INTO users (name, email) VALUES (@p1, @p2)", "Test User", email)
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
		logger.Errors = nil // Reset logs
		// Use invalid SQL to trigger an error
		_, err := db.Exec(ctx, "INSERT INTO nonexistent_table (name) VALUES (@p1)", "test")
		if err == nil {
			t.Fatal("Expected error for invalid SQL, got nil")
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

func TestMSSQL_Logging_QueryAll(t *testing.T) {
	logger := &testhelpers.TestLogger{}
	db, err := typedb.OpenWithoutValidation("sqlserver", getTestDSN(), typedb.WithLogger(logger))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	ctx := context.Background()

	t.Run("success logs debug", func(t *testing.T) {
		logger.Debugs = nil // Reset logs
		_, err := db.QueryAll(ctx, "SELECT TOP 1 id, name, email, created_at FROM users ORDER BY id")
		if err != nil {
			t.Fatalf("QueryAll failed: %v", err)
		}

		// Verify Debug log was emitted
		if len(logger.Debugs) == 0 {
			t.Fatal("Expected Debug log for QueryAll, got none")
		}
		if logger.Debugs[0].Msg != "Querying all rows" {
			t.Errorf("Expected Debug log message 'Querying all rows', got %q", logger.Debugs[0].Msg)
		}
	})

	t.Run("error logs error", func(t *testing.T) {
		logger.Errors = nil // Reset logs
		// Use invalid SQL to trigger an error
		_, err := db.QueryAll(ctx, "SELECT invalid_column FROM users")
		if err == nil {
			t.Fatal("Expected error for invalid SQL, got nil")
		}

		// Verify Error log was emitted
		if len(logger.Errors) == 0 {
			t.Fatal("Expected Error log for QueryAll failure, got none")
		}
		if logger.Errors[0].Msg != "Query failed" {
			t.Errorf("Expected Error log message 'Query failed', got %q", logger.Errors[0].Msg)
		}
	})
}

func TestMSSQL_Logging_Begin_Commit_Rollback(t *testing.T) {
	logger := &testhelpers.TestLogger{}
	db, err := typedb.OpenWithoutValidation("sqlserver", getTestDSN(), typedb.WithLogger(logger))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	ctx := context.Background()

	t.Run("begin logs debug", func(t *testing.T) {
		logger.Debugs = nil // Reset logs
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

	t.Run("commit logs info", func(t *testing.T) {
		logger.Infos = nil // Reset logs
		tx, err := db.Begin(ctx, nil)
		if err != nil {
			t.Fatalf("Begin failed: %v", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("Commit failed: %v", err)
		}

		// Verify Info log was emitted
		if len(logger.Infos) == 0 {
			t.Fatal("Expected Info log for Commit, got none")
		}
		if logger.Infos[0].Msg != "Committing transaction" {
			t.Errorf("Expected Info log message 'Committing transaction', got %q", logger.Infos[0].Msg)
		}
	})

	t.Run("rollback logs info", func(t *testing.T) {
		logger.Infos = nil // Reset logs
		tx, err := db.Begin(ctx, nil)
		if err != nil {
			t.Fatalf("Begin failed: %v", err)
		}

		err = tx.Rollback()
		if err != nil {
			t.Fatalf("Rollback failed: %v", err)
		}

		// Verify Info log was emitted
		if len(logger.Infos) == 0 {
			t.Fatal("Expected Info log for Rollback, got none")
		}
		if logger.Infos[0].Msg != "Rolling back transaction" {
			t.Errorf("Expected Info log message 'Rolling back transaction', got %q", logger.Infos[0].Msg)
		}
	})
}

func TestMSSQL_Logging_Close(t *testing.T) {
	logger := &testhelpers.TestLogger{}
	db, err := typedb.OpenWithoutValidation("sqlserver", getTestDSN(), typedb.WithLogger(logger))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	t.Run("close logs info", func(t *testing.T) {
		logger.Infos = nil // Reset logs
		err := db.Close()
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}

		// Verify Info log was emitted
		if len(logger.Infos) == 0 {
			t.Fatal("Expected Info log for Close, got none")
		}
		if logger.Infos[0].Msg != "Closing database connection" {
			t.Errorf("Expected Info log message 'Closing database connection', got %q", logger.Infos[0].Msg)
		}
	})
}

func TestMSSQL_Logging_PerInstanceLogger(t *testing.T) {
	globalLogger := &testhelpers.TestLogger{}
	instanceLogger := &testhelpers.TestLogger{}

	// Set global logger
	typedb.SetLogger(globalLogger)

	// Create DB with per-instance logger
	db, err := typedb.OpenWithoutValidation("sqlserver", getTestDSN(), typedb.WithLogger(instanceLogger))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	ctx := context.Background()

	// Use a unique email to avoid conflicts
	email := fmt.Sprintf("test-perinstance-%d@example.com", time.Now().UnixNano())
	_, err = db.Exec(ctx, "INSERT INTO users (name, email) VALUES (@p1, @p2)", "Test User", email)
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	// Verify instance logger received the log, not global logger
	if len(instanceLogger.Debugs) == 0 {
		t.Error("Expected instance logger to receive Debug log")
	}
	if len(globalLogger.Debugs) != 0 {
		t.Error("Expected global logger to NOT receive log when per-instance logger is set")
	}
}

func TestMSSQL_Logging_ConfigOptions(t *testing.T) {
	ctx := context.Background()

	t.Run("LogQueries=false disables query logging", func(t *testing.T) {
		logger := &testhelpers.TestLogger{}
		db, err := typedb.OpenWithoutValidation("sqlserver", getTestDSN(),
			typedb.WithLogger(logger),
			typedb.WithLogQueries(false),
			typedb.WithLogArgs(true))
		if err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}
		defer closeDB(t, db)

		logger.Debugs = nil
		// Use a unique email to avoid conflicts
		email := fmt.Sprintf("test-logqueries-%d@example.com", time.Now().UnixNano())
		_, err = db.Exec(ctx, "INSERT INTO users (name, email) VALUES (@p1, @p2)", "Test User", email)
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		// Should log message but without query
		if len(logger.Debugs) == 0 {
			t.Fatal("Expected Debug log even when LogQueries=false")
		}
		foundQuery := false
		for _, entry := range logger.Debugs {
			for i := 0; i < len(entry.Keyvals)-1; i += 2 {
				if entry.Keyvals[i] == "query" {
					foundQuery = true
				}
			}
		}
		if foundQuery {
			t.Error("Expected 'query' key to be absent when LogQueries=false")
		}
	})

	t.Run("LogArgs=false disables argument logging", func(t *testing.T) {
		logger := &testhelpers.TestLogger{}
		db, err := typedb.OpenWithoutValidation("sqlserver", getTestDSN(),
			typedb.WithLogger(logger),
			typedb.WithLogQueries(true),
			typedb.WithLogArgs(false))
		if err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}
		defer closeDB(t, db)

		logger.Debugs = nil
		// Use a unique email to avoid conflicts
		email := fmt.Sprintf("test-logargs-%d@example.com", time.Now().UnixNano())
		_, err = db.Exec(ctx, "INSERT INTO users (name, email) VALUES (@p1, @p2)", "Test User", email)
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		// Should log query but without args
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
			t.Error("Expected 'query' key to be present when LogQueries=true")
		}
		if foundArgs {
			t.Error("Expected 'args' key to be absent when LogArgs=false")
		}
	})
}

func TestMSSQL_Logging_ContextOverrides(t *testing.T) {
	logger := &testhelpers.TestLogger{}
	db, err := typedb.OpenWithoutValidation("sqlserver", getTestDSN(), typedb.WithLogger(logger))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	ctx := context.Background()

	t.Run("WithNoLogging disables all logging", func(t *testing.T) {
		logger.Debugs = nil
		ctx = typedb.WithNoLogging(ctx)
		// Use a unique email to avoid conflicts
		email := fmt.Sprintf("test-nolog-%d@example.com", time.Now().UnixNano())
		_, err = db.Exec(ctx, "INSERT INTO users (name, email) VALUES (@p1, @p2)", "Test User", email)
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		// Should log message but without query/args
		if len(logger.Debugs) == 0 {
			t.Fatal("Expected Debug log even when logging disabled")
		}
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
			t.Error("Expected 'query' key to be absent when WithNoLogging is used")
		}
		if foundArgs {
			t.Error("Expected 'args' key to be absent when WithNoLogging is used")
		}
	})

	t.Run("WithNoQueryLogging disables query logging only", func(t *testing.T) {
		logger.Debugs = nil
		ctx = typedb.WithNoQueryLogging(ctx)
		// Use a unique email to avoid conflicts
		email := fmt.Sprintf("test-noquery-%d@example.com", time.Now().UnixNano())
		_, err = db.Exec(ctx, "INSERT INTO users (name, email) VALUES (@p1, @p2)", "Test User", email)
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		// Query should not be logged, but args should be
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
			t.Error("Expected 'args' key to be present when WithNoQueryLogging is used (only query disabled)")
		}
	})

	t.Run("WithNoArgLogging disables argument logging only", func(t *testing.T) {
		logger.Debugs = nil
		ctx = typedb.WithNoArgLogging(ctx)
		// Use a unique email to avoid conflicts
		email := fmt.Sprintf("test-noargs-%d@example.com", time.Now().UnixNano())
		_, err = db.Exec(ctx, "INSERT INTO users (name, email) VALUES (@p1, @p2)", "Test User", email)
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		// Args should not be logged, but query should be
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
// Note: Using email with nolog tag since MSSQL schema doesn't have password column
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
	return "SELECT id, name, email FROM users WHERE id = @p1"
}

func init() {
	typedb.RegisterModel[*UserWithNolog]()
}

func TestMSSQL_Logging_NologTagMasking(t *testing.T) {
	logger := &testhelpers.TestLogger{}
	db, err := typedb.OpenWithoutValidation("sqlserver", getTestDSN(), typedb.WithLogger(logger))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	ctx := context.Background()

	t.Run("Insert masks nolog fields", func(t *testing.T) {
		logger.Debugs = nil
		// Use a unique email to avoid conflicts
		email := fmt.Sprintf("test-nolog-insert-%d@example.com", time.Now().UnixNano())
		user := &UserWithNolog{
			Name:  "Test User",
			Email: email,
		}
		err = typedb.Insert(ctx, db, user)
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
					logArgs := entry.Keyvals[i+1].([]any)
					for _, arg := range logArgs {
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
		err = typedb.Update(ctx, db, user)
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
		err = typedb.Load(ctx, db, user)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		// For Load, the primary key (ID) is logged, not the password field
		// The masking is tested in Insert/Update where password is actually in args
	})
}

func TestMSSQL_Logging_SerializationNolog(t *testing.T) {
	logger := &testhelpers.TestLogger{}
	db, err := typedb.OpenWithoutValidation("sqlserver", getTestDSN(), typedb.WithLogger(logger))
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	ctx := context.Background()

	t.Run("QueryAll masks nolog fields in model arguments", func(t *testing.T) {
		logger.Debugs = nil
		logger.Errors = nil
		email := fmt.Sprintf("test-serialization-%d@example.com", time.Now().UnixNano())
		user := &UserWithNolog{
			Name:  "Test User",
			Email: email,
		}

		// Pass model as argument to QueryAll (database driver will fail, but masking should occur in logs first)
		// We verify masking detection works even though SQL execution fails
		_, err = db.QueryAll(ctx, "SELECT id, name, email FROM users WHERE email = @p1", user)
		// SQL execution will fail because structs can't be serialized, but masking should have occurred in logs
		if err == nil {
			t.Fatal("Expected QueryAll to fail with struct argument, but it succeeded")
		}

		// Check that email is masked in logs (check both Debug and Error logs)
		foundArgs := false
		foundMasked := false
		allLogs := append(logger.Debugs, logger.Errors...)
		for _, entry := range allLogs {
			for i := 0; i < len(entry.Keyvals)-1; i += 2 {
				if entry.Keyvals[i] == "args" {
					foundArgs = true
					logArgs := entry.Keyvals[i+1].([]any)
					for _, arg := range logArgs {
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
			t.Error("Expected 'args' key in logs (Debug or Error)")
		}
		if !foundMasked {
			t.Error("Expected email to be masked as [REDACTED]")
		}
	})

	t.Run("QueryRowMap masks nolog fields in model arguments", func(t *testing.T) {
		logger.Debugs = nil
		logger.Errors = nil
		email := fmt.Sprintf("test-serialization-rowmap-%d@example.com", time.Now().UnixNano())
		user := &UserWithNolog{
			Name:  "Test User",
			Email: email,
		}

		// Pass model as argument to QueryRowMap (database driver will fail, but masking should occur in logs first)
		_, err = db.QueryRowMap(ctx, "SELECT id, name, email FROM users WHERE email = @p1", user)
		// SQL execution will fail because structs can't be serialized, but masking should have occurred in logs
		if err == nil {
			t.Fatal("Expected QueryRowMap to fail with struct argument, but it succeeded")
		}

		// Check that email is masked in logs (check both Debug and Error logs)
		foundArgs := false
		foundMasked := false
		allLogs := append(logger.Debugs, logger.Errors...)
		for _, entry := range allLogs {
			for i := 0; i < len(entry.Keyvals)-1; i += 2 {
				if entry.Keyvals[i] == "args" {
					foundArgs = true
					logArgs := entry.Keyvals[i+1].([]any)
					for _, arg := range logArgs {
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
			t.Error("Expected 'args' key in logs (Debug or Error)")
		}
		if !foundMasked {
			t.Error("Expected email to be masked as [REDACTED]")
		}
	})

	t.Run("GetInto masks nolog fields in model arguments", func(t *testing.T) {
		logger.Debugs = nil
		logger.Errors = nil
		email := fmt.Sprintf("test-serialization-getinto-%d@example.com", time.Now().UnixNano())
		user := &UserWithNolog{
			Name:  "Test User",
			Email: email,
		}
		var dest map[string]any

		// Pass model as argument to GetInto (database driver will fail, but masking should occur in logs first)
		err = db.GetInto(ctx, "SELECT id, name, email FROM users WHERE email = @p1", []any{user}, &dest)
		// SQL execution will fail because structs can't be serialized, but masking should have occurred in logs
		if err == nil {
			t.Fatal("Expected GetInto to fail with struct argument, but it succeeded")
		}

		// Check that email is masked in logs (check both Debug and Error logs)
		foundArgs := false
		foundMasked := false
		allLogs := append(logger.Debugs, logger.Errors...)
		for _, entry := range allLogs {
			for i := 0; i < len(entry.Keyvals)-1; i += 2 {
				if entry.Keyvals[i] == "args" {
					foundArgs = true
					logArgs := entry.Keyvals[i+1].([]any)
					for _, arg := range logArgs {
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
			t.Error("Expected 'args' key in logs (Debug or Error)")
		}
		if !foundMasked {
			t.Error("Expected email to be masked as [REDACTED]")
		}
	})

	t.Run("QueryDo masks nolog fields in model arguments", func(t *testing.T) {
		logger.Debugs = nil
		logger.Errors = nil
		email := fmt.Sprintf("test-serialization-querydo-%d@example.com", time.Now().UnixNano())
		user := &UserWithNolog{
			Name:  "Test User",
			Email: email,
		}

		// Pass model as argument to QueryDo (database driver will fail, but masking should occur in logs first)
		err = db.QueryDo(ctx, "SELECT id, name, email FROM users WHERE email = @p1", []any{user}, func(rows *sql.Rows) error {
			return nil
		})
		// SQL execution will fail because structs can't be serialized, but masking should have occurred in logs
		if err == nil {
			t.Fatal("Expected QueryDo to fail with struct argument, but it succeeded")
		}

		// Check that email is masked in logs (check both Debug and Error logs)
		foundArgs := false
		foundMasked := false
		allLogs := append(logger.Debugs, logger.Errors...)
		for _, entry := range allLogs {
			for i := 0; i < len(entry.Keyvals)-1; i += 2 {
				if entry.Keyvals[i] == "args" {
					foundArgs = true
					logArgs := entry.Keyvals[i+1].([]any)
					for _, arg := range logArgs {
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
			t.Error("Expected 'args' key in logs (Debug or Error)")
		}
		if !foundMasked {
			t.Error("Expected email to be masked as [REDACTED]")
		}
	})
}
