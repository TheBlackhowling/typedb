package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dsn := os.Getenv("SQLITE_DSN")
	if dsn == "" {
		dsn = "typedb_examples_test.db"
	}

	db, err := sql.Open("sqlite3", dsn+"?_foreign_keys=1")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		log.Fatalf("Failed to enable foreign keys: %v", err)
	}

	migrationFile := filepath.Join("migrations", "000001_create_tables.up.sql")
	sqlBytes, err := ioutil.ReadFile(migrationFile)
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
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
	// when using the sqlite3 driver
	if _, err := db.Exec(sqlContent); err != nil {
		log.Fatalf("Failed to execute migration: %v", err)
	}

	fmt.Println("✓ Migrations completed successfully")
}
