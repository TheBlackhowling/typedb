package typedb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// sqlQueryExecutor is an internal interface for SQL query operations.
// Both *sql.DB and *sql.Tx implement these methods.
type sqlQueryExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// NewDB creates a DB instance from an existing *sql.DB connection.
// The timeout parameter sets the default timeout for operations.
// The driverName parameter specifies the database driver name (e.g., "postgres", "mysql").
// If driverName is empty, it will be detected from the connection if possible.
// The logger parameter is optional - if nil, uses the global logger (defaults to no-op).
func NewDB(db *sql.DB, driverName string, timeout time.Duration) *DB {
	return NewDBWithLogger(db, driverName, timeout, nil)
}

// NewDBWithLogger creates a DB instance with a specific logger.
// If logger is nil, uses the global logger (defaults to no-op).
func NewDBWithLogger(db *sql.DB, driverName string, timeout time.Duration, logger Logger) *DB {
	if logger == nil {
		logger = defaultLogger
	}
	return &DB{
		db:         db,
		driverName: driverName,
		timeout:    timeout,
		logger:     logger,
	}
}

// getLoggerHelper returns the logger, defaulting to no-op if nil.
// This provides defensive programming against manually constructed instances.
func getLoggerHelper(logger Logger) Logger {
	if logger == nil {
		return defaultLogger
	}
	return logger
}

// getLogger returns the logger for this DB instance, defaulting to no-op if nil.
// This provides defensive programming against manually constructed DB instances.
func (d *DB) getLogger() Logger {
	return getLoggerHelper(d.logger)
}

// withTimeoutHelper ensures we always have a bounded context per operation.
// If the context already has a deadline, it returns the context as-is.
// Otherwise, it creates a new context with the provided timeout.
func withTimeoutHelper(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if _, has := ctx.Deadline(); has {
		return ctx, func() {}
	}
	to := timeout
	if to <= 0 {
		to = 5 * time.Second
	}
	return context.WithTimeout(ctx, to)
}

// withTimeout ensures we always have a bounded context per DB operation.
// If the context already has a deadline, it returns the context as-is.
// Otherwise, it creates a new context with the DB's default timeout.
func (d *DB) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return withTimeoutHelper(ctx, d.timeout)
}

// execHelper executes a query that doesn't return rows, with logging and timeout handling.
func execHelper(ctx context.Context, exec sqlQueryExecutor, logger Logger, timeout time.Duration, query string, args ...any) (sql.Result, error) {
	logger = getLoggerHelper(logger)
	logger.Debug("Executing query", "query", query, "args", args)
	ctx, cancel := withTimeoutHelper(ctx, timeout)
	defer cancel()
	result, err := exec.ExecContext(ctx, query, args...)
	if err != nil {
		logger.Error("Query execution failed", "query", query, "args", args, "error", err)
		return nil, err
	}
	return result, nil
}

// Exec implements Executor.Exec
// Executes a query that doesn't return rows (INSERT/UPDATE/DELETE/DDL).
func (d *DB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return execHelper(ctx, d.db, d.logger, d.timeout, query, args...)
}

// queryAllHelper executes a query and returns all rows as []map[string]any, with logging and timeout handling.
func queryAllHelper(ctx context.Context, exec sqlQueryExecutor, logger Logger, timeout time.Duration, query string, args ...any) ([]map[string]any, error) {
	logger = getLoggerHelper(logger)
	logger.Debug("Querying all rows", "query", query, "args", args)
	ctx, cancel := withTimeoutHelper(ctx, timeout)
	defer cancel()

	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Error("Query failed", "query", query, "args", args, "error", err)
		return nil, err
	}
	defer rows.Close()

	result, err := scanRowsToMaps(rows)
	if err != nil {
		logger.Error("Failed to scan rows", "query", query, "error", err)
		return nil, err
	}
	return result, nil
}

// QueryAll implements Executor.QueryAll
// Returns all rows as []map[string]any.
func (d *DB) QueryAll(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
	return queryAllHelper(ctx, d.db, d.logger, d.timeout, query, args...)
}

// queryRowMapHelper executes a query and returns the first row as map[string]any, with logging and timeout handling.
func queryRowMapHelper(ctx context.Context, exec sqlQueryExecutor, logger Logger, timeout time.Duration, query string, args ...any) (map[string]any, error) {
	logger = getLoggerHelper(logger)
	ctx, cancel := withTimeoutHelper(ctx, timeout)
	defer cancel()

	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Error("Query failed", "query", query, "args", args, "error", err)
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			logger.Error("Row iteration error", "query", query, "error", err)
			return nil, err
		}
		return nil, ErrNotFound
	}

	row, err := scanRowToMap(rows)
	if err != nil {
		logger.Error("Failed to scan row", "query", query, "error", err)
		return nil, err
	}

	// Ensure no more rows (shouldn't happen for single row queries, but check anyway)
	if rows.Next() {
		err := fmt.Errorf("typedb: QueryRowMap returned multiple rows")
		logger.Error("Multiple rows returned", "query", query, "error", err)
		return nil, err
	}

	return row, nil
}

// QueryRowMap implements Executor.QueryRowMap
// Returns the first row as map[string]any.
// Returns ErrNotFound if no rows are returned.
func (d *DB) QueryRowMap(ctx context.Context, query string, args ...any) (map[string]any, error) {
	return queryRowMapHelper(ctx, d.db, d.logger, d.timeout, query, args...)
}

// getIntoHelper scans a single row into dest pointers, with logging and timeout handling.
func getIntoHelper(ctx context.Context, exec sqlQueryExecutor, logger Logger, timeout time.Duration, query string, args []any, dest ...any) error {
	logger = getLoggerHelper(logger)
	logger.Debug("Scanning row into destination", "query", query, "args", args)
	ctx, cancel := withTimeoutHelper(ctx, timeout)
	defer cancel()

	err := exec.QueryRowContext(ctx, query, args...).Scan(dest...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Debug("No rows found", "query", query)
			return ErrNotFound
		}
		logger.Error("Failed to scan row", "query", query, "args", args, "error", err)
		return err
	}
	return nil
}

// GetInto implements Executor.GetInto
// Scans a single row into dest pointers.
// Returns ErrNotFound if no rows are returned.
func (d *DB) GetInto(ctx context.Context, query string, args []any, dest ...any) error {
	return getIntoHelper(ctx, d.db, d.logger, d.timeout, query, args, dest...)
}

// queryDoHelper executes a query and calls scan for each row (streaming), with logging and timeout handling.
func queryDoHelper(ctx context.Context, exec sqlQueryExecutor, logger Logger, timeout time.Duration, query string, args []any, scan func(rows *sql.Rows) error) error {
	logger = getLoggerHelper(logger)
	logger.Debug("Executing streaming query", "query", query, "args", args)
	ctx, cancel := withTimeoutHelper(ctx, timeout)
	defer cancel()

	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Error("Query failed", "query", query, "args", args, "error", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err := scan(rows); err != nil {
			logger.Error("Scan callback failed", "query", query, "error", err)
			return err
		}
	}

	if err := rows.Err(); err != nil {
		logger.Error("Row iteration error", "query", query, "error", err)
		return err
	}
	return nil
}

// QueryDo implements Executor.QueryDo
// Executes a query and calls scan for each row (streaming).
func (d *DB) QueryDo(ctx context.Context, query string, args []any, scan func(rows *sql.Rows) error) error {
	return queryDoHelper(ctx, d.db, d.logger, d.timeout, query, args, scan)
}

// Close closes the database connection.
func (d *DB) Close() error {
	d.getLogger().Info("Closing database connection")
	err := d.db.Close()
	if err != nil {
		d.getLogger().Error("Failed to close database connection", "error", err)
		return err
	}
	return nil
}

// Ping verifies the connection to the database is still alive.
func (d *DB) Ping(ctx context.Context) error {
	d.getLogger().Debug("Pinging database")
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()
	err := d.db.PingContext(ctx)
	if err != nil {
		d.getLogger().Error("Database ping failed", "error", err)
		return err
	}
	return nil
}

// Begin starts a new transaction.
// Note: The context passed to BeginTx is used only for starting the transaction.
// The transaction itself is not bound to this context's lifecycle - operations
// within the transaction use their own contexts via withTimeout.
func (d *DB) Begin(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	d.getLogger().Debug("Beginning transaction")
	// Use the original context for BeginTx - don't add timeout here
	// because BeginTx itself should complete quickly, and we don't want
	// to bind the transaction to a context that might be canceled
	tx, err := d.db.BeginTx(ctx, opts)
	if err != nil {
		d.getLogger().Error("Failed to begin transaction", "error", err)
		return nil, err
	}

	return &Tx{
		tx:         tx,
		driverName: d.driverName,
		timeout:    d.timeout,
		logger:     d.logger,
	}, nil
}

// WithTx executes a function within a transaction.
// The transaction is automatically committed if the function returns nil,
// or rolled back if the function returns an error.
func (d *DB) WithTx(ctx context.Context, fn func(*Tx) error, opts *sql.TxOptions) error {
	d.getLogger().Debug("Executing function within transaction")
	tx, err := d.Begin(ctx, opts)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		d.getLogger().Debug("Function returned error, rolling back transaction", "error", err)
		_ = tx.Rollback()
		return err
	}

	d.getLogger().Debug("Function completed successfully, committing transaction")
	return tx.Commit()
}

// Exec implements Executor.Exec for transactions
func (t *Tx) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return execHelper(ctx, t.tx, t.logger, t.timeout, query, args...)
}

// QueryAll implements Executor.QueryAll for transactions
func (t *Tx) QueryAll(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
	return queryAllHelper(ctx, t.tx, t.logger, t.timeout, query, args...)
}

// QueryRowMap implements Executor.QueryRowMap for transactions
func (t *Tx) QueryRowMap(ctx context.Context, query string, args ...any) (map[string]any, error) {
	return queryRowMapHelper(ctx, t.tx, t.logger, t.timeout, query, args...)
}

// GetInto implements Executor.GetInto for transactions
func (t *Tx) GetInto(ctx context.Context, query string, args []any, dest ...any) error {
	return getIntoHelper(ctx, t.tx, t.logger, t.timeout, query, args, dest...)
}

// QueryDo implements Executor.QueryDo for transactions
func (t *Tx) QueryDo(ctx context.Context, query string, args []any, scan func(rows *sql.Rows) error) error {
	return queryDoHelper(ctx, t.tx, t.logger, t.timeout, query, args, scan)
}

// Commit commits the transaction.
func (t *Tx) Commit() error {
	t.logger.Info("Committing transaction")
	err := t.tx.Commit()
	if err != nil {
		t.logger.Error("Transaction commit failed", "error", err)
		return err
	}
	return nil
}

// Rollback rolls back the transaction.
func (t *Tx) Rollback() error {
	t.logger.Info("Rolling back transaction")
	err := t.tx.Rollback()
	if err != nil {
		t.logger.Error("Transaction rollback failed", "error", err)
		return err
	}
	return nil
}

// getLogger returns the logger for this Tx instance, defaulting to no-op if nil.
// This provides defensive programming against manually constructed Tx instances.
func (t *Tx) getLogger() Logger {
	return getLoggerHelper(t.logger)
}

// withTimeout ensures we always have a bounded context per Tx operation.
func (t *Tx) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return withTimeoutHelper(ctx, t.timeout)
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
		// Normalize column names to lowercase for case-insensitive matching
		// This handles databases like Oracle that return uppercase column names
		colKey := strings.ToLower(col)
		if b, ok := val.([]byte); ok {
			// Convert []byte to string for easier handling
			result[colKey] = string(b)
		} else {
			result[colKey] = val
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
	logger := defaultLogger
	cfg := &Config{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		OpTimeout:       5 * time.Second,
	}

	for _, opt := range opts {
		opt(cfg)
		if cfg.Logger != nil {
			logger = cfg.Logger
		}
	}

	logger.Info("Opening database connection", "driver", driverName)
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		logger.Error("Failed to open database connection", "driver", driverName, "error", err)
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	logger.Info("Validating registered models")
	// Validate all registered models before returning
	MustValidateAllRegistered()

	logger.Info("Database connection opened successfully", "driver", driverName, "maxOpenConns", cfg.MaxOpenConns, "maxIdleConns", cfg.MaxIdleConns)
	return NewDBWithLogger(db, driverName, cfg.OpTimeout, logger), nil
}

// OpenWithoutValidation opens a database connection without validation.
// Useful for testing or when you want to defer validation.
func OpenWithoutValidation(driverName, dsn string, opts ...Option) (*DB, error) {
	logger := defaultLogger
	cfg := &Config{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		OpTimeout:       5 * time.Second,
	}

	for _, opt := range opts {
		opt(cfg)
		if cfg.Logger != nil {
			logger = cfg.Logger
		}
	}

	logger.Info("Opening database connection without validation", "driver", driverName)
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		logger.Error("Failed to open database connection", "driver", driverName, "error", err)
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	logger.Info("Database connection opened successfully (without validation)", "driver", driverName)
	return NewDBWithLogger(db, driverName, cfg.OpTimeout, logger), nil
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

// WithLogger sets the logger for the database connection.
func WithLogger(logger Logger) Option {
	return func(cfg *Config) {
		cfg.Logger = logger
	}
}
