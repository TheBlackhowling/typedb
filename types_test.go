package typedb

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestDBStruct(t *testing.T) {
	// Test that DB struct can be created (fields only, no methods yet)
	db := &DB{
		db:      nil, // Will be set in executor.go
		timeout: 5 * time.Second,
		logger:  GetLogger(),
	}

	if db.timeout != 5*time.Second {
		t.Errorf("DB.timeout = %v, want %v", db.timeout, 5*time.Second)
	}
}

func TestTxStruct(t *testing.T) {
	// Test that Tx struct can be created (fields only, no methods yet)
	tx := &Tx{
		tx:      nil, // Will be set in executor.go
		timeout: 5 * time.Second,
		logger:  GetLogger(),
	}

	if tx.timeout != 5*time.Second {
		t.Errorf("Tx.timeout = %v, want %v", tx.timeout, 5*time.Second)
	}
}

func TestConfigStruct(t *testing.T) {
	cfg := &Config{
		DSN:             "postgres://localhost/test",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		OpTimeout:       5 * time.Second,
	}

	if cfg.DSN != "postgres://localhost/test" {
		t.Errorf("Config.DSN = %q, want %q", cfg.DSN, "postgres://localhost/test")
	}

	if cfg.MaxOpenConns != 10 {
		t.Errorf("Config.MaxOpenConns = %d, want %d", cfg.MaxOpenConns, 10)
	}
}

func TestModelStruct(t *testing.T) {
	// Test that Model struct can be created
	model := &Model{}

	if model == nil {
		t.Fatal("Model should not be nil")
	}
}

// Note: Interface implementation tests will be added in later layers:
// - ModelInterface implementation tested in model_test.go (Layer 7)
// - Executor implementation tested in executor_test.go (Layer 4)

// MockExecutor is a test helper that implements Executor interface
// Used for testing in later layers
type MockExecutor struct {
	ExecFunc        func(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryAllFunc    func(ctx context.Context, query string, args ...any) ([]map[string]any, error)
	QueryRowMapFunc func(ctx context.Context, query string, args ...any) (map[string]any, error)
	GetIntoFunc     func(ctx context.Context, query string, args []any, dest ...any) error
	QueryDoFunc     func(ctx context.Context, query string, args []any, scan func(rows *sql.Rows) error) error
}

func (m *MockExecutor) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if m.ExecFunc != nil {
		return m.ExecFunc(ctx, query, args...)
	}
	return nil, nil
}

func (m *MockExecutor) QueryAll(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
	if m.QueryAllFunc != nil {
		return m.QueryAllFunc(ctx, query, args...)
	}
	return nil, nil
}

func (m *MockExecutor) QueryRowMap(ctx context.Context, query string, args ...any) (map[string]any, error) {
	if m.QueryRowMapFunc != nil {
		return m.QueryRowMapFunc(ctx, query, args...)
	}
	return nil, nil
}

func (m *MockExecutor) GetInto(ctx context.Context, query string, args []any, dest ...any) error {
	if m.GetIntoFunc != nil {
		return m.GetIntoFunc(ctx, query, args, dest...)
	}
	return nil
}

func (m *MockExecutor) QueryDo(ctx context.Context, query string, args []any, scan func(rows *sql.Rows) error) error {
	if m.QueryDoFunc != nil {
		return m.QueryDoFunc(ctx, query, args, scan)
	}
	return nil
}
