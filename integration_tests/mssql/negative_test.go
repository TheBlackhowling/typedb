package main

import (
	"context"
	"strings"
	"testing"

	"github.com/TheBlackHowling/typedb"
	_ "github.com/microsoft/go-mssqldb" // SQL Server driver
)

func TestMSSQL_Negative_InvalidQuery(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Invalid SQL query
	_, err = typedb.QueryAll[*User](ctx, db, "SELECT invalid_column FROM users")
	if err == nil {
		t.Fatal("QueryAll should return error for invalid SQL")
	}
}

func TestMSSQL_Negative_ConstraintViolation(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Get existing user email
	existingUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT TOP 1 id, name, email, created_at FROM users ORDER BY id")
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
			db.Exec(ctx, "DELETE FROM users WHERE id = @p1", duplicateUser.ID)
		}
		t.Fatal("Insert should fail with unique constraint violation")
	}

	// Verify error indicates constraint violation
	if !strings.Contains(err.Error(), "unique") && !strings.Contains(err.Error(), "duplicate") && !strings.Contains(err.Error(), "UNIQUE") {
		t.Errorf("Expected constraint violation error, got: %v", err)
	}
}
