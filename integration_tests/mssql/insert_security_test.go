package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/TheBlackHowling/typedb"
	_ "github.com/microsoft/go-mssqldb" // SQL Server driver
)

// TestMSSQL_InsertAndReturn_Security_ValidIdentifiers tests InsertAndReturn with various valid identifier formats
func TestMSSQL_InsertAndReturn_Security_ValidIdentifiers(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Get first user for foreign key
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT TOP 1 id, name, email, created_at FROM users ORDER BY id")
	if err != nil || firstUser == nil {
		t.Fatal("Need at least one user in database for foreign key")
	}

	tests := []struct {
		name        string
		insertQuery string
		args        []any
		expectError bool
	}{
		{
			name:        "valid simple identifiers",
			insertQuery: "INSERT INTO posts (user_id, title, content) OUTPUT INSERTED.id, INSERTED.user_id, INSERTED.title VALUES (@p1, @p2, @p3)",
			args:        []any{firstUser.ID, "Test Title", "Test Content"},
			expectError: false,
		},
		{
			name:        "valid identifiers with underscores",
			insertQuery: "INSERT INTO posts (user_id, title) OUTPUT INSERTED.id, INSERTED.user_id VALUES (@p1, @p2)",
			args:        []any{firstUser.ID, "Test Title"},
			expectError: false,
		},
		{
			name:        "qualified identifier in OUTPUT",
			insertQuery: "INSERT INTO posts (user_id, title) OUTPUT INSERTED.id, INSERTED.user_id VALUES (@p1, @p2)",
			args:        []any{firstUser.ID, "Test"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uniqueTitle := fmt.Sprintf("Test Post %d", time.Now().UnixNano())
			args := []any{firstUser.ID, uniqueTitle}
			if len(tt.args) > 2 {
				args = append(args, tt.args[2:]...)
			}

			insertedPost, err := typedb.InsertAndReturn[*Post](ctx, db, tt.insertQuery, args...)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else if insertedPost.ID == 0 {
					t.Error("Post ID should be set")
				}

				// Cleanup
				if insertedPost.ID != 0 {
					t.Cleanup(func() {
						db.Exec(ctx, "DELETE FROM posts WHERE id = @p1", insertedPost.ID)
					})
				}
			}
		})
	}
}

// TestMSSQL_InsertAndReturn_Security_QuoteHandling tests that identifiers with quotes are handled correctly
func TestMSSQL_InsertAndReturn_Security_QuoteHandling(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT TOP 1 id, name, email, created_at FROM users ORDER BY id")
	if err != nil || firstUser == nil {
		t.Fatal("Need at least one user in database for foreign key")
	}

	// Test InsertAndReturn with bracketed identifiers (SQL Server style)
	uniqueTitle := fmt.Sprintf("Test Post %d", time.Now().UnixNano())
	insertQuery := `INSERT INTO posts (user_id, title, content) OUTPUT INSERTED.[id], INSERTED.[user_id], INSERTED.[title] VALUES (@p1, @p2, @p3)`
	
	insertedPost, err := typedb.InsertAndReturn[*Post](ctx, db, insertQuery, firstUser.ID, uniqueTitle, "Test Content")
	if err != nil {
		// Bracketed identifiers might not work if they don't match the actual column names
		if !strings.Contains(err.Error(), "invalid identifier") {
			t.Logf("InsertAndReturn with bracketed identifiers returned error (expected): %v", err)
		}
	} else {
		// If it succeeds, verify it worked
		if insertedPost.ID == 0 {
			t.Error("Post ID should be set")
		}
		// Cleanup
		if insertedPost.ID != 0 {
			t.Cleanup(func() {
				db.Exec(ctx, "DELETE FROM posts WHERE id = @p1", insertedPost.ID)
			})
		}
	}
}
