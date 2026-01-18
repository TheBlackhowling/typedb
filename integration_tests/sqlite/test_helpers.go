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
	dsn := getTestDSN()

	// Remove existing test database
	os.Remove(dsn)

	// Open raw database connection for migrations
	sqlDB, err := sql.Open("sqlite3", dsn)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Run migrations
	driver, err := sqlite3.WithInstance(sqlDB, &sqlite3.Config{})
	if err != nil {
		sqlDB.Close()
		t.Fatalf("Failed to create migration driver: %v", err)
	}

	// Get migrations directory path - resolve relative to test file location
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)
	migrationsPath := filepath.Join(testDir, "migrations")

	// Convert to absolute path for file:// URL
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		sqlDB.Close()
		t.Fatalf("Failed to resolve migrations path: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+filepath.ToSlash(absPath),
		"sqlite3", driver)
	if err != nil {
		sqlDB.Close()
		t.Fatalf("Failed to create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		sqlDB.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Close the raw connection - migrations are done
	sqlDB.Close()

	// Now open with typedb (will use the same database file)
	db, err := typedb.OpenWithoutValidation("sqlite3", dsn)
	if err != nil {
		t.Fatalf("Failed to open typedb connection: %v", err)
	}

	return db
}
