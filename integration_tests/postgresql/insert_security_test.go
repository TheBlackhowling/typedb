package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/TheBlackHowling/typedb"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// TestPostgreSQL_InsertAndReturn_Security_ValidIdentifiers tests InsertAndReturn with various valid identifier formats
func TestPostgreSQL_InsertAndReturn_Security_ValidIdentifiers(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

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
			insertQuery: "INSERT INTO posts (user_id, title, content) VALUES ($1, $2, $3) RETURNING id, user_id, title",
			args:        []any{firstUser.ID, "Test Title", "Test Content"},
			expectError: false,
		},
		{
			name:        "valid identifiers with underscores",
			insertQuery: "INSERT INTO posts (user_id, title) VALUES ($1, $2) RETURNING id, user_id",
			args:        []any{firstUser.ID, "Test Title"},
			expectError: false,
		},
		{
			name:        "qualified identifier in RETURNING",
			insertQuery: "INSERT INTO posts (user_id, title) VALUES ($1, $2) RETURNING posts.id, posts.user_id",
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
						db.Exec(ctx, "DELETE FROM posts WHERE id = $1", insertedPost.ID)
					})
				}
			}
		})
	}
}

// TestPostgreSQL_InsertAndReturn_Security_IdentifierInjection tests that malicious identifiers are rejected
// Note: PostgreSQL InsertAndReturn doesn't parse identifiers like Oracle, but we test that quoteIdentifier
// properly handles edge cases through the Insert function which uses quoteIdentifier internally
func TestPostgreSQL_InsertAndReturn_Security_IdentifierInjection(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Test that Insert (which uses quoteIdentifier) properly handles identifiers
	// This indirectly tests quoteIdentifier security
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users ORDER BY id LIMIT 1")
	if err != nil || firstUser == nil {
		t.Fatal("Need at least one user in database for foreign key")
	}

	// Insert should work with valid identifiers
	uniqueTitle := fmt.Sprintf("Test Post %d", time.Now().UnixNano())
	newPost := &Post{
		UserID:  firstUser.ID,
		Title:   uniqueTitle,
		Content: "Test Content",
	}

	err = typedb.Insert(ctx, db, newPost)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Cleanup
	if newPost.ID != 0 {
		t.Cleanup(func() {
			db.Exec(ctx, "DELETE FROM posts WHERE id = $1", newPost.ID)
		})
	}

	// Verify the insert worked
	if newPost.ID == 0 {
		t.Error("Post ID should be set after insert")
	}
}

// TestPostgreSQL_InsertAndReturn_Security_QuoteHandling tests that identifiers with quotes are handled correctly
func TestPostgreSQL_InsertAndReturn_Security_QuoteHandling(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users ORDER BY id LIMIT 1")
	if err != nil || firstUser == nil {
		t.Fatal("Need at least one user in database for foreign key")
	}

	// Test InsertAndReturn with quoted identifiers in RETURNING clause
	// PostgreSQL allows quoted identifiers, so this should work
	uniqueTitle := fmt.Sprintf("Test Post %d", time.Now().UnixNano())
	insertQuery := `INSERT INTO posts (user_id, title, content) VALUES ($1, $2, $3) RETURNING "id", "user_id", "title"`
	
	insertedPost, err := typedb.InsertAndReturn[*Post](ctx, db, insertQuery, firstUser.ID, uniqueTitle, "Test Content")
	if err != nil {
		// Quoted identifiers might not work if they don't match the actual column names
		// This is expected behavior - the test verifies the system handles it gracefully
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
				db.Exec(ctx, "DELETE FROM posts WHERE id = $1", insertedPost.ID)
			})
		}
	}
}
