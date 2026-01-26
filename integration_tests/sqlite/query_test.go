package main

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/TheBlackHowling/typedb"
)

func TestSQLite_QueryAll(t *testing.T) {
	db := setupTestDB(t)
	defer closeDB(t, db)
	defer os.Remove(getTestDSN())

	ctx := context.Background()
	users, err := typedb.QueryAll[*User](ctx, db, "SELECT id, name, email, created_at FROM users ORDER BY id")
	if err != nil {
		t.Fatalf("QueryAll failed: %v", err)
	}

	if len(users) == 0 {
		t.Fatal("Expected at least one user")
	}

	if users[0].ID == 0 {
		t.Error("User ID should not be zero")
	}
	if users[0].Name == "" {
		t.Error("User name should not be empty")
	}
}

func TestSQLite_QueryFirst(t *testing.T) {
	db := setupTestDB(t)
	defer closeDB(t, db)
	defer os.Remove(getTestDSN())

	ctx := context.Background()
	user, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = ?", 1)
	if err != nil {
		t.Fatalf("QueryFirst failed: %v", err)
	}

	if user == nil {
		t.Fatal("Expected a user, got nil")
	}

	if user.ID != 1 {
		t.Errorf("Expected user ID 1, got %d", user.ID)
	}
}

func TestSQLite_QueryOne(t *testing.T) {
	db := setupTestDB(t)
	defer closeDB(t, db)
	defer os.Remove(getTestDSN())

	ctx := context.Background()
	user, err := typedb.QueryOne[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = ?", 1)
	if err != nil {
		t.Fatalf("QueryOne failed: %v", err)
	}

	if user.ID != 1 {
		t.Errorf("Expected user ID 1, got %d", user.ID)
	}
}

func TestSQLite_QueryFirst_NoRows(t *testing.T) {
	db := setupTestDB(t)
	defer closeDB(t, db)
	defer os.Remove(getTestDSN())

	ctx := context.Background()

	// Query for non-existent user
	user, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = ?", 99999)
	if err != nil {
		t.Fatalf("QueryFirst should not return error for no rows, got: %v", err)
	}

	if user != nil {
		t.Error("QueryFirst should return nil for no rows")
	}
}

func TestSQLite_QueryOne_NoRows(t *testing.T) {
	db := setupTestDB(t)
	defer closeDB(t, db)
	defer os.Remove(getTestDSN())

	ctx := context.Background()

	// Query for non-existent user
	user, err := typedb.QueryOne[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = ?", 99999)
	if err == nil {
		t.Fatal("QueryOne should return error for no rows")
	}

	if err != typedb.ErrNotFound {
		t.Errorf("QueryOne should return ErrNotFound for no rows, got: %v", err)
	}

	if user != nil {
		t.Error("QueryOne should return nil when error occurs")
	}
}

func TestSQLite_QueryOne_MultipleRows(t *testing.T) {
	db := setupTestDB(t)
	defer closeDB(t, db)
	defer os.Remove(getTestDSN())

	ctx := context.Background()

	// Query that returns multiple rows (no WHERE clause)
	user, err := typedb.QueryOne[*User](ctx, db, "SELECT id, name, email, created_at FROM users")
	if err == nil {
		t.Fatal("QueryOne should return error for multiple rows")
	}

	if !strings.Contains(err.Error(), "multiple rows") {
		t.Errorf("QueryOne should return error about multiple rows, got: %v", err)
	}

	if user != nil {
		t.Error("QueryOne should return nil when error occurs")
	}
}
