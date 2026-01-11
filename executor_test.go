package typedb

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestNewDB(t *testing.T) {
	sqlDB := &sql.DB{}
	timeout := 10 * time.Second

	db := NewDB(sqlDB, timeout)

	if db.db != sqlDB {
		t.Error("Expected db to be set")
	}
	if db.timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, db.timeout)
	}
}

func TestDB_WithTimeout(t *testing.T) {
	db := NewDB(&sql.DB{}, 5*time.Second)

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
	ctxWithDeadline, _ := context.WithTimeout(context.Background(), 2*time.Second)
	newCtx2, cancel2 := db.withTimeout(ctxWithDeadline)
	if newCtx2 != ctxWithDeadline {
		t.Error("Expected same context when deadline already exists")
	}
	cancel2()

	// Test with zero timeout (should default to 5s)
	dbZero := NewDB(&sql.DB{}, 0)
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
	// Test Close with a real database connection
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skipf("sqlite3 driver not available: %v", err)
	}
	if sqlDB == nil {
		t.Skip("sqlite3 driver not available")
	}

	db := NewDB(sqlDB, 5*time.Second)
	err = db.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

func TestDB_Ping(t *testing.T) {
	// Test Ping with a real database connection
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skipf("sqlite3 driver not available: %v", err)
	}
	if sqlDB == nil {
		t.Skip("sqlite3 driver not available")
	}
	defer sqlDB.Close()

	db := NewDB(sqlDB, 5*time.Second)
	ctx := context.Background()
	err = db.Ping(ctx)
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
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
	// Note: This requires a real database connection, so we'll test with sqlite in-memory
	// Skip if sqlite3 driver is not available
	typedbDB, err := OpenWithoutValidation("sqlite3", ":memory:")
	if err != nil {
		t.Skipf("sqlite3 driver not available: %v", err)
	}
	if typedbDB == nil {
		t.Skip("sqlite3 driver not available")
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
	db, err := OpenWithoutValidation("sqlite3", ":memory:",
		WithMaxOpenConns(20),
		WithMaxIdleConns(10),
		WithTimeout(10*time.Second),
	)
	if err != nil {
		t.Skipf("sqlite3 driver not available: %v", err)
	}
	if db == nil {
		t.Skip("sqlite3 driver not available")
	}
	defer db.Close()

	if db.timeout != 10*time.Second {
		t.Errorf("Expected timeout 10s, got %v", db.timeout)
	}
}

func TestOpenWithoutValidation(t *testing.T) {
	db, err := OpenWithoutValidation("sqlite3", ":memory:")
	if err != nil {
		t.Skipf("sqlite3 driver not available: %v", err)
	}
	if db == nil {
		t.Skip("sqlite3 driver not available")
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
	ctxWithDeadline, _ := context.WithTimeout(context.Background(), 2*time.Second)
	newCtx2, cancel2 := typedbTx.withTimeout(ctxWithDeadline)
	if newCtx2 != ctxWithDeadline {
		t.Error("Expected same context when deadline already exists")
	}
	cancel2()
}

func TestTx_Commit(t *testing.T) {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skipf("sqlite3 driver not available: %v", err)
	}
	if sqlDB == nil {
		t.Skip("sqlite3 driver not available")
	}
	defer sqlDB.Close()

	tx, err := sqlDB.Begin()
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	typedbTx := &Tx{
		tx:      tx,
		timeout: 5 * time.Second,
	}

	err = typedbTx.Commit()
	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}
}

func TestTx_Rollback(t *testing.T) {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skipf("sqlite3 driver not available: %v", err)
	}
	if sqlDB == nil {
		t.Skip("sqlite3 driver not available")
	}
	defer sqlDB.Close()

	tx, err := sqlDB.Begin()
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	typedbTx := &Tx{
		tx:      tx,
		timeout: 5 * time.Second,
	}

	err = typedbTx.Rollback()
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}
}
