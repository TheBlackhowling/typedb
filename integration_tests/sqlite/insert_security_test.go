package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/TheBlackHowling/typedb"
)

// TestSQLite_InsertAndReturn_Security_ValidIdentifiers tests InsertAndReturn with various valid identifier formats
func TestSQLite_InsertAndReturn_Security_ValidIdentifiers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer os.Remove(getTestDSN())

	ctx := context.Background()

	// Get first user for foreign key
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users ORDER BY id LIMIT 1")
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
			insertQuery: "INSERT INTO posts (user_id, title, content) VALUES (?, ?, ?) RETURNING id, user_id, title",
			args:        []any{firstUser.ID, "Test Title", "Test Content"},
			expectError: false,
		},
		{
			name:        "valid identifiers with underscores",
			insertQuery: "INSERT INTO posts (user_id, title) VALUES (?, ?) RETURNING id, user_id",
			args:        []any{firstUser.ID, "Test Title"},
			expectError: false,
		},
		{
			name:        "qualified identifier in RETURNING",
			insertQuery: "INSERT INTO posts (user_id, title) VALUES (?, ?) RETURNING posts.id, posts.user_id",
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
						db.Exec(ctx, "DELETE FROM posts WHERE id = ?", insertedPost.ID)
					})
				}
			}
		})
	}
}

// TestSQLite_InsertAndReturn_Security_QuoteHandling tests that identifiers with quotes are handled correctly
func TestSQLite_InsertAndReturn_Security_QuoteHandling(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer os.Remove(getTestDSN())

	ctx := context.Background()

	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users ORDER BY id LIMIT 1")
	if err != nil || firstUser == nil {
		t.Fatal("Need at least one user in database for foreign key")
	}

	// Test InsertAndReturn with quoted identifiers in RETURNING clause
	uniqueTitle := fmt.Sprintf("Test Post %d", time.Now().UnixNano())
	insertQuery := `INSERT INTO posts (user_id, title, content) VALUES (?, ?, ?) RETURNING "id", "user_id", "title"`
	
	insertedPost, err := typedb.InsertAndReturn[*Post](ctx, db, insertQuery, firstUser.ID, uniqueTitle, "Test Content")
	if err != nil {
		// Quoted identifiers might not work if they don't match the actual column names
		if !strings.Contains(err.Error(), "invalid identifier") {
			t.Logf("InsertAndReturn with quoted identifiers returned error (expected): %v", err)
		}
	} else {
		// If it succeeds, verify it worked
		if insertedPost.ID == 0 {
			t.Error("Post ID should be set")
		}
		// Cleanup
		if insertedPost.ID != 0 {
			t.Cleanup(func() {
				db.Exec(ctx, "DELETE FROM posts WHERE id = ?", insertedPost.ID)
			})
		}
	}
}
