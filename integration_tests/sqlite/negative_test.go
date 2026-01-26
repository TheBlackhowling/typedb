package main

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/TheBlackHowling/typedb"
)

func TestSQLite_Negative_InvalidQuery(t *testing.T) {
	db := setupTestDB(t)
	defer closeDB(t, db)
	defer os.Remove(getTestDSN())

	ctx := context.Background()

	// Invalid SQL query
	_, err := typedb.QueryAll[*User](ctx, db, "SELECT invalid_column FROM users")
	if err == nil {
		t.Fatal("QueryAll should return error for invalid SQL")
	}
}

func TestSQLite_Negative_ConstraintViolation(t *testing.T) {
	db := setupTestDB(t)
	defer closeDB(t, db)
	defer os.Remove(getTestDSN())

	ctx := context.Background()

	// Get existing user email
	existingUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users ORDER BY id LIMIT 1")
	if err != nil || existingUser == nil {
		t.Fatal("Need at least one user in database")
	}

	// Try to insert user with duplicate email (unique constraint violation)
	duplicateUser := &User{
		Name:  "Duplicate Email User",
		Email: existingUser.Email, // Duplicate email
	}
	err = typedb.Insert(ctx, db, duplicateUser)
	if err == nil {
		// Clean up if insert somehow succeeded
		if duplicateUser.ID != 0 {
			db.Exec(ctx, "DELETE FROM users WHERE id = ?", duplicateUser.ID)
		}
		t.Fatal("Insert should fail with unique constraint violation")
	}

	// Verify error indicates constraint violation
	if !strings.Contains(err.Error(), "unique") && !strings.Contains(err.Error(), "duplicate") && !strings.Contains(err.Error(), "UNIQUE") {
		t.Errorf("Expected constraint violation error, got: %v", err)
	}
}
