package main

import (
	"database/sql"
	"os"
	"testing"

	"github.com/TheBlackHowling/typedb"
)

func getTestDSN() string {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "user:password@tcp(localhost:3306)/typedb_examples?parseTime=true"
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
