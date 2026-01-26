package main

import (
	"database/sql"
	"os"
	"testing"

	"github.com/TheBlackHowling/typedb"
)

func getTestDSN() string {
	dsn := os.Getenv("ORACLE_DSN")
	if dsn == "" {
		dsn = "oracle://user:password@localhost:1521/XE"
	}
	return dsn
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
