package typedb

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/mattn/go-sqlite3" // SQLite driver for Open* tests
)

func TestNewDB(t *testing.T) {
	sqlDB := &sql.DB{}
	timeout := 10 * time.Second

	db := NewDB(sqlDB, "test", timeout)

	if db.db != sqlDB {
		t.Error("Expected db to be set")
	}
	if db.timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, db.timeout)
	}
}

func TestDB_WithTimeout(t *testing.T) {
	db := NewDB(&sql.DB{}, "test", 5*time.Second)

	// Test with context that has no deadline
	ctx := context.Background()
	newCtx, cancel := db.withTimeout(ctx)
	if newCtx == ctx {
		t.Error("Expected new context with timeout")
	}
	deadline, ok := newCtx.Deadline()
	if !ok {
		t.Error("Expected context to have deadline")
	}
	if deadline.Sub(time.Now()) > 6*time.Second || deadline.Sub(time.Now()) < 4*time.Second {
		t.Errorf("Expected deadline around 5s, got %v", deadline.Sub(time.Now()))
	}
	cancel()

	// Test with context that already has deadline
	ctxWithDeadline, cancelDeadline := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelDeadline()
	newCtx2, cancel2 := db.withTimeout(ctxWithDeadline)
	if newCtx2 != ctxWithDeadline {
		t.Error("Expected same context when deadline already exists")
	}
	cancel2()

	// Test with zero timeout (should default to 5s)
	dbZero := NewDB(&sql.DB{}, "test", 0)
	ctx3 := context.Background()
	newCtx3, cancel3 := dbZero.withTimeout(ctx3)
	deadline3, ok := newCtx3.Deadline()
	if !ok {
		t.Error("Expected context to have deadline")
	}
	if deadline3.Sub(time.Now()) > 6*time.Second || deadline3.Sub(time.Now()) < 4*time.Second {
		t.Errorf("Expected default deadline around 5s, got %v", deadline3.Sub(time.Now()))
	}
	cancel3()
}

func TestDB_Close(t *testing.T) {
	// Test Close with a mock database connection
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}

	mock.ExpectClose()

	db := NewDB(sqlDB, "test", 5*time.Second)
	err = db.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestDB_Ping(t *testing.T) {
	// Test Ping with a mock database connection
	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer sqlDB.Close()

	mock.ExpectPing()

	db := NewDB(sqlDB, "test", 5*time.Second)
	ctx := context.Background()
	err = db.Ping(ctx)
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestScanRowToMap(t *testing.T) {
	// This is tested indirectly through QueryAll/QueryRowMap tests
	// For now, we'll test the helper functions when we have a real DB setup
}

func TestScanRowsToMaps(t *testing.T) {
	// This is tested indirectly through QueryAll tests
	// For now, we'll test the helper functions when we have a real DB setup
}

func TestOpen(t *testing.T) {
	// Test that Open creates a DB with default config
	// Uses sqlite in-memory database for testing
	typedbDB, err := OpenWithoutValidation("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("OpenWithoutValidation failed: %v", err)
	}
	defer typedbDB.Close()

	if typedbDB.db == nil {
		t.Error("Expected db to be set")
	}
	if typedbDB.timeout != 5*time.Second {
		t.Errorf("Expected default timeout 5s, got %v", typedbDB.timeout)
	}
}

func TestOpen_WithOptions(t *testing.T) {
	// Test OpenWithOptions with sqlite in-memory database
	db, err := OpenWithoutValidation("sqlite3", ":memory:",
		WithMaxOpenConns(20),
		WithMaxIdleConns(10),
		WithTimeout(10*time.Second),
	)
	if err != nil {
		t.Fatalf("OpenWithoutValidation failed: %v", err)
	}
	defer db.Close()

	if db.timeout != 10*time.Second {
		t.Errorf("Expected timeout 10s, got %v", db.timeout)
	}
}

func TestOpenWithoutValidation(t *testing.T) {
	// Test OpenWithoutValidation with sqlite in-memory database
	db, err := OpenWithoutValidation("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("OpenWithoutValidation failed: %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Error("Expected db to be non-nil")
	}
}

func TestOptionFunctions(t *testing.T) {
	cfg := &Config{}

	WithMaxOpenConns(20)(cfg)
	if cfg.MaxOpenConns != 20 {
		t.Errorf("Expected MaxOpenConns 20, got %d", cfg.MaxOpenConns)
	}

	WithMaxIdleConns(10)(cfg)
	if cfg.MaxIdleConns != 10 {
		t.Errorf("Expected MaxIdleConns 10, got %d", cfg.MaxIdleConns)
	}

	timeout := 15 * time.Second
	WithTimeout(timeout)(cfg)
	if cfg.OpTimeout != timeout {
		t.Errorf("Expected OpTimeout %v, got %v", timeout, cfg.OpTimeout)
	}

	lifetime := 60 * time.Minute
	WithConnMaxLifetime(lifetime)(cfg)
	if cfg.ConnMaxLifetime != lifetime {
		t.Errorf("Expected ConnMaxLifetime %v, got %v", lifetime, cfg.ConnMaxLifetime)
	}

	idleTime := 10 * time.Minute
	WithConnMaxIdleTime(idleTime)(cfg)
	if cfg.ConnMaxIdleTime != idleTime {
		t.Errorf("Expected ConnMaxIdleTime %v, got %v", idleTime, cfg.ConnMaxIdleTime)
	}
}

func TestTx_WithTimeout(t *testing.T) {
	// Create a mock transaction (we don't need a real DB for this test)
	typedbTx := &Tx{
		tx:      nil, // Not needed for timeout test
		timeout: 5 * time.Second,
		logger:  GetLogger(),
	}

	// Test with context that has no deadline
	ctx := context.Background()
	newCtx, cancel := typedbTx.withTimeout(ctx)
	if newCtx == ctx {
		t.Error("Expected new context with timeout")
	}
	deadline, ok := newCtx.Deadline()
	if !ok {
		t.Error("Expected context to have deadline")
	}
	if deadline.Sub(time.Now()) > 6*time.Second || deadline.Sub(time.Now()) < 4*time.Second {
		t.Errorf("Expected deadline around 5s, got %v", deadline.Sub(time.Now()))
	}
	cancel()

	// Test with context that already has deadline
	ctxWithDeadline, cancelDeadline := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelDeadline()
	newCtx2, cancel2 := typedbTx.withTimeout(ctxWithDeadline)
	if newCtx2 != ctxWithDeadline {
		t.Error("Expected same context when deadline already exists")
	}
	cancel2()
}

func TestTx_Commit(t *testing.T) {
	// Test Commit with a mock database connection
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer sqlDB.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	tx, err := sqlDB.Begin()
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	typedbTx := &Tx{
		tx:      tx,
		timeout: 5 * time.Second,
		logger:  GetLogger(),
	}

	err = typedbTx.Commit()
	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestTx_Rollback(t *testing.T) {
	// Test Rollback with a mock database connection
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer sqlDB.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	tx, err := sqlDB.Begin()
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	typedbTx := &Tx{
		tx:      tx,
		timeout: 5 * time.Second,
		logger:  GetLogger(),
	}

	err = typedbTx.Rollback()
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

// ========== SQLMock Tests for DB Methods ==========

func TestDB_Exec_Sqlmock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "test", 5*time.Second)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO users").
			WithArgs("test").
			WillReturnResult(sqlmock.NewResult(1, 1))

		result, err := typedbDB.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "test")
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		lastID, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("LastInsertId failed: %v", err)
		}
		if lastID != 1 {
			t.Errorf("Expected LastInsertId 1, got %d", lastID)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mock.ExpectExec("INSERT INTO users").
			WithArgs("test").
			WillReturnError(expectedErr)

		_, err := typedbDB.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "test")
		if err != expectedErr {
			t.Fatalf("Expected error %v, got %v", expectedErr, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %v", err)
		}
	})
}

func TestDB_QueryAll_Sqlmock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "test", 5*time.Second)
	ctx := context.Background()

	t.Run("success with rows", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Alice").
			AddRow(2, "Bob")

		mock.ExpectQuery("SELECT id, name FROM users").
			WillReturnRows(rows)

		result, err := typedbDB.QueryAll(ctx, "SELECT id, name FROM users")
		if err != nil {
			t.Fatalf("QueryAll failed: %v", err)
		}

		if len(result) != 2 {
			t.Fatalf("Expected 2 rows, got %d", len(result))
		}

		if result[0]["id"] != int64(1) || result[0]["name"] != "Alice" {
			t.Errorf("First row incorrect: %+v", result[0])
		}

		if result[1]["id"] != int64(2) || result[1]["name"] != "Bob" {
			t.Errorf("Second row incorrect: %+v", result[1])
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %v", err)
		}
	})

	t.Run("success with empty result", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"})

		mock.ExpectQuery("SELECT id, name FROM users").
			WillReturnRows(rows)

		result, err := typedbDB.QueryAll(ctx, "SELECT id, name FROM users")
		if err != nil {
			t.Fatalf("QueryAll failed: %v", err)
		}

		if len(result) != 0 {
			t.Fatalf("Expected 0 rows, got %d", len(result))
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mock.ExpectQuery("SELECT id, name FROM users").
			WillReturnError(expectedErr)

		_, err := typedbDB.QueryAll(ctx, "SELECT id, name FROM users")
		if err != expectedErr {
			t.Fatalf("Expected error %v, got %v", expectedErr, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %v", err)
		}
	})
}

func TestDB_QueryRowMap_Sqlmock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "test", 5*time.Second)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Alice")

		mock.ExpectQuery("SELECT id, name FROM users WHERE id").
			WithArgs(1).
			WillReturnRows(rows)

		result, err := typedbDB.QueryRowMap(ctx, "SELECT id, name FROM users WHERE id = $1", 1)
		if err != nil {
			t.Fatalf("QueryRowMap failed: %v", err)
		}

		if result["id"] != int64(1) || result["name"] != "Alice" {
			t.Errorf("Row incorrect: %+v", result)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %v", err)
		}
	})

	t.Run("no rows - returns ErrNotFound", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"})

		mock.ExpectQuery("SELECT id, name FROM users WHERE id").
			WithArgs(999).
			WillReturnRows(rows)

		_, err := typedbDB.QueryRowMap(ctx, "SELECT id, name FROM users WHERE id = $1", 999)
		if err != ErrNotFound {
			t.Fatalf("Expected ErrNotFound, got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %v", err)
		}
	})

	t.Run("multiple rows - returns error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Alice").
			AddRow(2, "Bob")

		mock.ExpectQuery("SELECT id, name FROM users").
			WillReturnRows(rows)

		_, err := typedbDB.QueryRowMap(ctx, "SELECT id, name FROM users")
		if err == nil {
			t.Fatal("Expected error for multiple rows")
		}

		if err.Error() != "typedb: QueryRowMap returned multiple rows" {
			t.Errorf("Expected multiple rows error, got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %v", err)
		}
	})

	t.Run("query error", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mock.ExpectQuery("SELECT id, name FROM users").
			WillReturnError(expectedErr)

		_, err := typedbDB.QueryRowMap(ctx, "SELECT id, name FROM users")
		if err != expectedErr {
			t.Fatalf("Expected error %v, got %v", expectedErr, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %v", err)
		}
	})
}

func TestDB_GetInto_Sqlmock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "test", 5*time.Second)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Alice")

		mock.ExpectQuery("SELECT id, name FROM users WHERE id").
			WithArgs(1).
			WillReturnRows(rows)

		var id int
		var name string
		err := typedbDB.GetInto(ctx, "SELECT id, name FROM users WHERE id = $1", []any{1}, &id, &name)
		if err != nil {
			t.Fatalf("GetInto failed: %v", err)
		}

		if id != 1 || name != "Alice" {
			t.Errorf("Values incorrect: id=%d, name=%s", id, name)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %v", err)
		}
	})

	t.Run("no rows - returns ErrNotFound", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"})

		mock.ExpectQuery("SELECT id, name FROM users WHERE id").
			WithArgs(999).
			WillReturnRows(rows)

		var id int
		var name string
		err := typedbDB.GetInto(ctx, "SELECT id, name FROM users WHERE id = $1", []any{999}, &id, &name)
		if err != ErrNotFound {
			t.Fatalf("Expected ErrNotFound, got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mock.ExpectQuery("SELECT id, name FROM users").
			WillReturnError(expectedErr)

		var id int
		var name string
		err := typedbDB.GetInto(ctx, "SELECT id, name FROM users", []any{}, &id, &name)
		if err != expectedErr {
			t.Fatalf("Expected error %v, got %v", expectedErr, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %v", err)
		}
	})
}

func TestDB_QueryDo_Sqlmock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "test", 5*time.Second)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Alice").
			AddRow(2, "Bob")

		mock.ExpectQuery("SELECT id, name FROM users").
			WillReturnRows(rows)

		count := 0
		err := typedbDB.QueryDo(ctx, "SELECT id, name FROM users", []any{}, func(rows *sql.Rows) error {
			count++
			return nil
		})
		if err != nil {
			t.Fatalf("QueryDo failed: %v", err)
		}

		if count != 2 {
			t.Errorf("Expected scan to be called 2 times, got %d", count)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %v", err)
		}
	})

	t.Run("scan error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Alice")

		mock.ExpectQuery("SELECT id, name FROM users").
			WillReturnRows(rows)

		scanErr := errors.New("scan error")
		err := typedbDB.QueryDo(ctx, "SELECT id, name FROM users", []any{}, func(rows *sql.Rows) error {
			return scanErr
		})
		if err != scanErr {
			t.Fatalf("Expected scan error %v, got %v", scanErr, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %v", err)
		}
	})

	t.Run("query error", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mock.ExpectQuery("SELECT id, name FROM users").
			WillReturnError(expectedErr)

		err := typedbDB.QueryDo(ctx, "SELECT id, name FROM users", []any{}, func(rows *sql.Rows) error {
			return nil
		})
		if err != expectedErr {
			t.Fatalf("Expected error %v, got %v", expectedErr, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %v", err)
		}
	})
}

// ========== SQLMock Tests for Tx Methods ==========

func TestTx_Exec_Sqlmock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mockTx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	typedbTx := &Tx{
		tx:      mockTx,
		timeout: 5 * time.Second,
		logger:  GetLogger(),
	}
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO users").
			WithArgs("test").
			WillReturnResult(sqlmock.NewResult(1, 1))

		result, err := typedbTx.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "test")
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		lastID, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("LastInsertId failed: %v", err)
		}
		if lastID != 1 {
			t.Errorf("Expected LastInsertId 1, got %d", lastID)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mock.ExpectExec("INSERT INTO users").
			WithArgs("test").
			WillReturnError(expectedErr)

		_, err := typedbTx.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "test")
		if err != expectedErr {
			t.Fatalf("Expected error %v, got %v", expectedErr, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %v", err)
		}
	})
}

func TestTx_QueryAll_Sqlmock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mockTx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	typedbTx := &Tx{
		tx:      mockTx,
		timeout: 5 * time.Second,
		logger:  GetLogger(),
	}
	ctx := context.Background()

	t.Run("success with rows", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Alice").
			AddRow(2, "Bob")

		mock.ExpectQuery("SELECT id, name FROM users").
			WillReturnRows(rows)

		result, err := typedbTx.QueryAll(ctx, "SELECT id, name FROM users")
		if err != nil {
			t.Fatalf("QueryAll failed: %v", err)
		}

		if len(result) != 2 {
			t.Fatalf("Expected 2 rows, got %d", len(result))
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mock.ExpectQuery("SELECT id, name FROM users").
			WillReturnError(expectedErr)

		_, err := typedbTx.QueryAll(ctx, "SELECT id, name FROM users")
		if err != expectedErr {
			t.Fatalf("Expected error %v, got %v", expectedErr, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %v", err)
		}
	})
}

func TestTx_QueryRowMap_Sqlmock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mockTx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	typedbTx := &Tx{
		tx:      mockTx,
		timeout: 5 * time.Second,
		logger:  GetLogger(),
	}
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Alice")

		mock.ExpectQuery("SELECT id, name FROM users WHERE id").
			WithArgs(1).
			WillReturnRows(rows)

		result, err := typedbTx.QueryRowMap(ctx, "SELECT id, name FROM users WHERE id = $1", 1)
		if err != nil {
			t.Fatalf("QueryRowMap failed: %v", err)
		}

		if result["id"] != int64(1) || result["name"] != "Alice" {
			t.Errorf("Row incorrect: %+v", result)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %v", err)
		}
	})

	t.Run("no rows - returns ErrNotFound", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"})

		mock.ExpectQuery("SELECT id, name FROM users WHERE id").
			WithArgs(999).
			WillReturnRows(rows)

		_, err := typedbTx.QueryRowMap(ctx, "SELECT id, name FROM users WHERE id = $1", 999)
		if err != ErrNotFound {
			t.Fatalf("Expected ErrNotFound, got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %v", err)
		}
	})

	t.Run("multiple rows - returns error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Alice").
			AddRow(2, "Bob")

		mock.ExpectQuery("SELECT id, name FROM users").
			WillReturnRows(rows)

		_, err := typedbTx.QueryRowMap(ctx, "SELECT id, name FROM users")
		if err == nil {
			t.Fatal("Expected error for multiple rows")
		}

		if err.Error() != "typedb: QueryRowMap returned multiple rows" {
			t.Errorf("Expected multiple rows error, got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %v", err)
		}
	})
}

func TestTx_GetInto_Sqlmock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mockTx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	typedbTx := &Tx{
		tx:      mockTx,
		timeout: 5 * time.Second,
		logger:  GetLogger(),
	}
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Alice")

		mock.ExpectQuery("SELECT id, name FROM users WHERE id").
			WithArgs(1).
			WillReturnRows(rows)

		var id int
		var name string
		err := typedbTx.GetInto(ctx, "SELECT id, name FROM users WHERE id = $1", []any{1}, &id, &name)
		if err != nil {
			t.Fatalf("GetInto failed: %v", err)
		}

		if id != 1 || name != "Alice" {
			t.Errorf("Values incorrect: id=%d, name=%s", id, name)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %v", err)
		}
	})

	t.Run("no rows - returns ErrNotFound", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"})

		mock.ExpectQuery("SELECT id, name FROM users WHERE id").
			WithArgs(999).
			WillReturnRows(rows)

		var id int
		var name string
		err := typedbTx.GetInto(ctx, "SELECT id, name FROM users WHERE id = $1", []any{999}, &id, &name)
		if err != ErrNotFound {
			t.Fatalf("Expected ErrNotFound, got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %v", err)
		}
	})
}

func TestTx_QueryDo_Sqlmock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mockTx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	typedbTx := &Tx{
		tx:      mockTx,
		timeout: 5 * time.Second,
		logger:  GetLogger(),
	}
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Alice").
			AddRow(2, "Bob")

		mock.ExpectQuery("SELECT id, name FROM users").
			WillReturnRows(rows)

		count := 0
		err := typedbTx.QueryDo(ctx, "SELECT id, name FROM users", []any{}, func(rows *sql.Rows) error {
			count++
			return nil
		})
		if err != nil {
			t.Fatalf("QueryDo failed: %v", err)
		}

		if count != 2 {
			t.Errorf("Expected scan to be called 2 times, got %d", count)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mock.ExpectQuery("SELECT id, name FROM users").
			WillReturnError(expectedErr)

		err := typedbTx.QueryDo(ctx, "SELECT id, name FROM users", []any{}, func(rows *sql.Rows) error {
			return nil
		})
		if err != expectedErr {
			t.Fatalf("Expected error %v, got %v", expectedErr, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unmet expectations: %v", err)
		}
	})
}
