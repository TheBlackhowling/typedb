package testhelpers

import (
	"database/sql"
	"testing"

	"github.com/TheBlackHowling/typedb"
)

// CloseDB safely closes a typedb.DB instance, logging any errors.
// Use this in defer statements to handle Close() errors properly.
func CloseDB(t *testing.T, db *typedb.DB) {
	if db == nil {
		return
	}
	if err := db.Close(); err != nil {
		t.Logf("Warning: failed to close typedb.DB: %v", err)
	}
}

// CloseSQLDB safely closes a *sql.DB instance, logging any errors.
// Use this in defer statements to handle Close() errors properly.
func CloseSQLDB(t *testing.T, db *sql.DB) {
	if db == nil {
		return
	}
	if err := db.Close(); err != nil {
		t.Logf("Warning: failed to close *sql.DB: %v", err)
	}
}

// CloseRows safely closes a *sql.Rows instance, logging any errors.
// Use this in defer statements to handle Close() errors properly.
func CloseRows(t *testing.T, rows *sql.Rows) {
	if rows == nil {
		return
	}
	if err := rows.Close(); err != nil {
		t.Logf("Warning: failed to close *sql.Rows: %v", err)
	}
}
