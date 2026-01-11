package typedb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// NewDB creates a DB instance from an existing *sql.DB connection.
// The timeout parameter sets the default timeout for operations.
func NewDB(db *sql.DB, timeout time.Duration) *DB {
	return &DB{
		db:      db,
		timeout: timeout,
	}
}

// withTimeout ensures we always have a bounded context per DB operation.
// If the context already has a deadline, it returns the context as-is.
// Otherwise, it creates a new context with the DB's default timeout.
func (d *DB) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, has := ctx.Deadline(); has {
		return ctx, func() {}
	}
	to := d.timeout
	if to <= 0 {
		to = 5 * time.Second
	}
	return context.WithTimeout(ctx, to)
}

// Exec implements Executor.Exec
// Executes a query that doesn't return rows (INSERT/UPDATE/DELETE/DDL).
func (d *DB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()
	return d.db.ExecContext(ctx, query, args...)
}

// QueryAll implements Executor.QueryAll
// Returns all rows as []map[string]any.
func (d *DB) QueryAll(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()

	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRowsToMaps(rows)
}

// QueryRowMap implements Executor.QueryRowMap
// Returns the first row as map[string]any.
// Returns ErrNotFound if no rows are returned.
func (d *DB) QueryRowMap(ctx context.Context, query string, args ...any) (map[string]any, error) {
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()

	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, ErrNotFound
	}

	row, err := scanRowToMap(rows)
	if err != nil {
		return nil, err
	}

	// Ensure no more rows (shouldn't happen for single row queries, but check anyway)
	if rows.Next() {
		return nil, fmt.Errorf("typedb: QueryRowMap returned multiple rows")
	}

	return row, nil
}

// GetInto implements Executor.GetInto
// Scans a single row into dest pointers.
// Returns ErrNotFound if no rows are returned.
func (d *DB) GetInto(ctx context.Context, query string, args []any, dest ...any) error {
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()

	err := d.db.QueryRowContext(ctx, query, args...).Scan(dest...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// QueryDo implements Executor.QueryDo
// Executes a query and calls scan for each row (streaming).
func (d *DB) QueryDo(ctx context.Context, query string, args []any, scan func(rows *sql.Rows) error) error {
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()

	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err := scan(rows); err != nil {
			return err
		}
	}

	return rows.Err()
}

// Close closes the database connection.
func (d *DB) Close() error {
	return d.db.Close()
}

// Ping verifies the connection to the database is still alive.
func (d *DB) Ping(ctx context.Context) error {
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()
	return d.db.PingContext(ctx)
}

// Begin starts a new transaction.
func (d *DB) Begin(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()

	tx, err := d.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &Tx{
		tx:      tx,
		timeout: d.timeout,
	}, nil
}

// WithTx executes a function within a transaction.
// The transaction is automatically committed if the function returns nil,
// or rolled back if the function returns an error.
func (d *DB) WithTx(ctx context.Context, fn func(*Tx) error, opts *sql.TxOptions) error {
	tx, err := d.Begin(ctx, opts)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Exec implements Executor.Exec for transactions
func (t *Tx) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	ctx, cancel := t.withTimeout(ctx)
	defer cancel()
	return t.tx.ExecContext(ctx, query, args...)
}

// QueryAll implements Executor.QueryAll for transactions
func (t *Tx) QueryAll(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
	ctx, cancel := t.withTimeout(ctx)
	defer cancel()

	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRowsToMaps(rows)
}

// QueryRowMap implements Executor.QueryRowMap for transactions
func (t *Tx) QueryRowMap(ctx context.Context, query string, args ...any) (map[string]any, error) {
	ctx, cancel := t.withTimeout(ctx)
	defer cancel()

	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, ErrNotFound
	}

	row, err := scanRowToMap(rows)
	if err != nil {
		return nil, err
	}

	// Ensure no more rows
	if rows.Next() {
		return nil, fmt.Errorf("typedb: QueryRowMap returned multiple rows")
	}

	return row, nil
}

// GetInto implements Executor.GetInto for transactions
func (t *Tx) GetInto(ctx context.Context, query string, args []any, dest ...any) error {
	ctx, cancel := t.withTimeout(ctx)
	defer cancel()

	err := t.tx.QueryRowContext(ctx, query, args...).Scan(dest...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// QueryDo implements Executor.QueryDo for transactions
func (t *Tx) QueryDo(ctx context.Context, query string, args []any, scan func(rows *sql.Rows) error) error {
	ctx, cancel := t.withTimeout(ctx)
	defer cancel()

	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err := scan(rows); err != nil {
			return err
		}
	}

	return rows.Err()
}

// Commit commits the transaction.
func (t *Tx) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction.
func (t *Tx) Rollback() error {
	return t.tx.Rollback()
}

// withTimeout ensures we always have a bounded context per Tx operation.
func (t *Tx) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, has := ctx.Deadline(); has {
		return ctx, func() {}
	}
	to := t.timeout
	if to <= 0 {
		to = 5 * time.Second
	}
	return context.WithTimeout(ctx, to)
}

// scanRowsToMaps scans all rows into a slice of maps.
func scanRowsToMaps(rows *sql.Rows) ([]map[string]any, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]any
	for rows.Next() {
		row, err := scanRowToMapWithCols(rows, cols)
		if err != nil {
			return nil, err
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// scanRowToMapWithCols scans a single row into a map[string]any using pre-fetched columns.
func scanRowToMapWithCols(rows *sql.Rows, cols []string) (map[string]any, error) {
	values := make([]any, len(cols))
	valuePtrs := make([]any, len(cols))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, err
	}

	result := make(map[string]any)
	for i, col := range cols {
		val := values[i]
		if b, ok := val.([]byte); ok {
			// Convert []byte to string for easier handling
			result[col] = string(b)
		} else {
			result[col] = val
		}
	}

	return result, nil
}

// scanRowToMap scans a single row into a map[string]any.
// Assumes rows.Next() has already been called.
func scanRowToMap(rows *sql.Rows) (map[string]any, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	return scanRowToMapWithCols(rows, cols)
}

// Open opens a database connection with validation.
// Calls MustValidateAllRegistered() to ensure all registered models are valid.
func Open(driverName, dsn string, opts ...Option) (*DB, error) {
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		OpTimeout:       5 * time.Second,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Validate all registered models before returning
	MustValidateAllRegistered()

	return NewDB(db, cfg.OpTimeout), nil
}

// OpenWithoutValidation opens a database connection without validation.
// Useful for testing or when you want to defer validation.
func OpenWithoutValidation(driverName, dsn string, opts ...Option) (*DB, error) {
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		OpTimeout:       5 * time.Second,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	return NewDB(db, cfg.OpTimeout), nil
}

// Option functions for configuring database connections

// WithMaxOpenConns sets the maximum number of open connections.
func WithMaxOpenConns(n int) Option {
	return func(cfg *Config) {
		cfg.MaxOpenConns = n
	}
}

// WithMaxIdleConns sets the maximum number of idle connections.
func WithMaxIdleConns(n int) Option {
	return func(cfg *Config) {
		cfg.MaxIdleConns = n
	}
}

// WithConnMaxLifetime sets the maximum lifetime of a connection.
func WithConnMaxLifetime(d time.Duration) Option {
	return func(cfg *Config) {
		cfg.ConnMaxLifetime = d
	}
}

// WithConnMaxIdleTime sets the maximum idle time of a connection.
func WithConnMaxIdleTime(d time.Duration) Option {
	return func(cfg *Config) {
		cfg.ConnMaxIdleTime = d
	}
}

// WithTimeout sets the default operation timeout.
func WithTimeout(d time.Duration) Option {
	return func(cfg *Config) {
		cfg.OpTimeout = d
	}
}
