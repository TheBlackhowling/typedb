package main

import (
	"context"
	"strings"
	"testing"

	"github.com/TheBlackHowling/typedb"
	_ "github.com/lib/pq" // PostgreSQL driver
)

func TestPostgreSQL_QueryAll(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	users, err := typedb.QueryAll[*User](ctx, db, "SELECT id, name, email, created_at FROM users ORDER BY id")
	if err != nil {
		t.Fatalf("QueryAll failed: %v", err)
	}

	if len(users) == 0 {
		t.Fatal("Expected at least one user")
	}

	// Verify first user
	if users[0].ID == 0 {
		t.Error("User ID should not be zero")
	}
	if users[0].Name == "" {
		t.Error("User name should not be empty")
	}
	if users[0].Email == "" {
		t.Error("User email should not be empty")
	}
}

func TestPostgreSQL_QueryFirst(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	user, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = $1", 1)
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

func TestPostgreSQL_QueryOne(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	user, err := typedb.QueryOne[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = $1", 1)
	if err != nil {
		t.Fatalf("QueryOne failed: %v", err)
	}

	if user.ID != 1 {
		t.Errorf("Expected user ID 1, got %d", user.ID)
	}
}

func TestPostgreSQL_QueryFirst_NoRows(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Query for non-existent user
	user, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = $1", 99999)
	if err != nil {
		t.Fatalf("QueryFirst should not return error for no rows, got: %v", err)
	}

	if user != nil {
		t.Error("QueryFirst should return nil for no rows")
	}
}

func TestPostgreSQL_QueryOne_NoRows(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Query for non-existent user
	user, err := typedb.QueryOne[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = $1", 99999)
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

func TestPostgreSQL_QueryOne_MultipleRows(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

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
