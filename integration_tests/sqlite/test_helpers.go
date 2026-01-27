package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/TheBlackHowling/typedb"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

func getTestDSN() string {
	dsn := os.Getenv("SQLITE_DSN")
	if dsn == "" {
		dsn = "typedb_examples_test.db"
	}
	return dsn
}

func setupTestDB(t *testing.T) *typedb.DB {
	return setupTestDBWithLogger(t, nil)
}

func setupTestDBWithLogger(t *testing.T, logger typedb.Logger) *typedb.DB {
	dsn := getTestDSN()

	// Remove existing test database (for clean state)
	_ = os.Remove(dsn) // Ignore error - file may not exist

	// Run migrations if database doesn't exist or is empty
	// In docker-compose workflow, migrations may already be run, but for local testing we need them
	runMigrationsIfNeeded(t, dsn)

	// Now open with typedb
	var opts []typedb.Option
	if logger != nil {
		opts = append(opts, typedb.WithLogger(logger))
	}
	db, err := typedb.OpenWithoutValidation("sqlite3", dsn, opts...)
	if err != nil {
		t.Fatalf("Failed to open typedb connection: %v", err)
	}

	return db
}

// closeDB safely closes a typedb.DB instance, logging any errors.
// Use this in defer statements to handle Close() errors properly.
func closeDB(t *testing.T, db *typedb.DB) {
	if db == nil {
		return
	}
	if err := db.Close(); err != nil {
		t.Logf("Warning: failed to close typedb.DB: %v", err)
	}
}

// closeSQLDB safely closes a *sql.DB instance, logging any errors.
// Use this in defer statements to handle Close() errors properly.
func closeSQLDB(t *testing.T, db *sql.DB) {
	if db == nil {
		return
	}
	if err := db.Close(); err != nil {
		t.Logf("Warning: failed to close *sql.DB: %v", err)
	}
}

func runMigrationsIfNeeded(t *testing.T, dsn string) {
	// Open raw database connection for migrations
	sqlDB, err := sql.Open("sqlite3", dsn)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer closeSQLDB(t, sqlDB)

	// Check if migrations are needed by checking if users table exists
	var tableExists int
	err = sqlDB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&tableExists)
	if err == nil && tableExists > 0 {
		// Table exists, migrations likely already run
		return
	}

	// Run migrations
	driver, err := sqlite3.WithInstance(sqlDB, &sqlite3.Config{})
	if err != nil {
		t.Fatalf("Failed to create migration driver: %v", err)
	}

	// Get migrations directory path - resolve relative to test file location
	_, testFile, _, _ := runtime.Caller(1) // Caller(1) because we're called from setupTestDBWithLogger
	testDir := filepath.Dir(testFile)
	migrationsPath := filepath.Join(testDir, "migrations")

	// Convert to absolute path for file:// URL
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		t.Fatalf("Failed to resolve migrations path: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+filepath.ToSlash(absPath),
		"sqlite3", driver)
	if err != nil {
		t.Fatalf("Failed to create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("Failed to run migrations: %v", err)
	}
}
