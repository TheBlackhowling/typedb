package typedb

import (
	"context"
	"database/sql"
	"time"
)

// Executor defines the interface for executing database queries.
// Both DB and Tx implement this interface, providing a unified API.
type Executor interface {
	// Exec executes a query that doesn't return rows (INSERT/UPDATE/DELETE/DDL)
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)

	// QueryAll returns all rows as []map[string]any
	QueryAll(ctx context.Context, query string, args ...any) ([]map[string]any, error)

	// QueryRowMap returns the first row as map[string]any (or sql.ErrNoRows)
	QueryRowMap(ctx context.Context, query string, args ...any) (map[string]any, error)

	// GetInto scans a single row into dest pointers
	GetInto(ctx context.Context, query string, args []any, dest ...any) error

	// QueryDo executes a query and calls scan for each row (streaming)
	QueryDo(ctx context.Context, query string, args []any, scan func(rows *sql.Rows) error) error
}

// DB wraps *sql.DB and provides query execution with timeout handling.
// DB implements the Executor interface.
type DB struct {
	db      *sql.DB
	timeout time.Duration // default timeout for operations
}

// Tx wraps *sql.Tx and provides transaction-scoped query execution.
// Tx implements the Executor interface.
type Tx struct {
	tx      *sql.Tx
	timeout time.Duration
}

// Config holds database connection and pool configuration.
type Config struct {
	DSN             string        // Connection string
	MaxOpenConns    int           // Default: 10
	MaxIdleConns    int           // Default: 5
	ConnMaxLifetime time.Duration // Default: 30m
	ConnMaxIdleTime time.Duration // Default: 5m
	OpTimeout       time.Duration // Default: 5s
}

// ModelInterface defines the contract for model types that can be deserialized.
// Models must implement Deserialize to convert database rows into struct fields.
type ModelInterface interface {
	Deserialize(row map[string]any) error
}

// Model is the base struct that models should embed.
// It provides common functionality for model types.
// Methods will be implemented in model.go (Layer 7).
type Model struct{}

// Option configures DB connection settings.
// Used with Open() and OpenWithoutValidation() functions.
type Option func(*Config)
