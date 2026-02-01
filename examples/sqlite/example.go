package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/TheBlackHowling/typedb"
	"github.com/TheBlackHowling/typedb/examples/seed"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

func main() {
	ctx := context.Background()

	// Get DSN from environment variable or use default
	dsn := os.Getenv("SQLITE_DSN")
	if dsn == "" {
		dsn = "typedb_examples.db"
	}

	// Run migrations before opening typedb connection
	if err := runMigrations(dsn); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Open database connection with foreign keys enabled
	// SQLite requires foreign keys to be enabled per connection
	db, err := typedb.Open("sqlite3", dsn+"?_foreign_keys=1")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fmt.Println("✓ Connected to SQLite")

	// Clear and seed database with random data
	fmt.Println("\n--- Clearing and Seeding Database ---")
	if err := seed.ClearDatabase(ctx, db); err != nil {
		log.Fatalf("Failed to clear database: %v", err)
	}
	if err := seed.SeedDatabase(ctx, db, 10); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	// Run examples by category
	firstUser := runQueryExamples(ctx, db)
	runLoadExamples(ctx, db, firstUser)
	postID := runInsertExamples(ctx, db, firstUser)
	runUpdateExamples(ctx, db, firstUser)
	runTransactionExamples(ctx, db, firstUser)
	runRawQueryExamples(ctx, db, firstUser)
	runLoadCompositeExample(ctx, db, firstUser, postID)

	fmt.Println("\n✓ All examples completed successfully!")
}

func runMigrations(dsn string) error {
	// Open raw SQL connection for migrations
	db, err := sql.Open("sqlite3", dsn+"?_foreign_keys=1")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Discover all migration files in the migrations directory
	migrationsDir := "migrations"
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Collect all .up.sql files and sort them by filename
	var migrationFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}

	// Sort migration files by filename to ensure correct order
	sort.Strings(migrationFiles)

	// Run all migration files in order
	for _, migrationFile := range migrationFiles {
		migrationPath := filepath.Join(migrationsDir, migrationFile)
		sqlBytes, err := os.ReadFile(migrationPath) // #nosec G304 // file path is safe - relative path constructed with filepath.Join, not user input
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", migrationFile, err)
		}

		// Remove comments and clean up the SQL
		sqlContent := string(sqlBytes)
		lines := strings.Split(sqlContent, "\n")
		var cleanedLines []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			// Skip comment-only lines
			if strings.HasPrefix(line, "--") {
				continue
			}
			// Remove inline comments
			if idx := strings.Index(line, "--"); idx != -1 {
				line = line[:idx]
				line = strings.TrimSpace(line)
			}
			if line != "" {
				cleanedLines = append(cleanedLines, line)
			}
		}
		sqlContent = strings.Join(cleanedLines, " ")

		// Execute all statements - SQLite supports multiple statements in one Exec call
		if sqlContent != "" {
			if _, err := db.Exec(sqlContent); err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", migrationFile, err)
			}
		}
	}

	fmt.Println("✓ Migrations completed successfully")
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
