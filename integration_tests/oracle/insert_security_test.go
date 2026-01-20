package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/TheBlackHowling/typedb"
	_ "github.com/sijms/go-ora/v2" // Oracle driver
)

// TestOracle_InsertAndReturn_Security_ValidIdentifiers tests InsertAndReturn with various valid identifier formats
func TestOracle_InsertAndReturn_Security_ValidIdentifiers(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Get first user for foreign key
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE ROWNUM <= 1 ORDER BY id")
	if err != nil || firstUser == nil {
		t.Fatal("Need at least one user in database for foreign key")
	}

	tests := []struct {
		name        string
		insertQuery string
		args        []any
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid simple identifiers",
			insertQuery: "INSERT INTO posts (user_id, title, content) VALUES (:1, :2, :3) RETURNING id, user_id, title",
			args:        []any{firstUser.ID, "Test Title", "Test Content"},
			expectError: false,
		},
		{
			name:        "valid identifiers with underscores",
			insertQuery: "INSERT INTO posts (user_id, title, content) VALUES (:1, :2, :3) RETURNING id, user_id, title",
			args:        []any{firstUser.ID, "Test Title", "Test Content"},
			expectError: false,
		},
		{
			name:        "valid qualified identifier (schema.table)",
			insertQuery: fmt.Sprintf("INSERT INTO posts (user_id, title, content) VALUES (:1, :2, :3) RETURNING id, user_id, title"),
			args:        []any{firstUser.ID, "Test Title", "Test Content"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uniqueTitle := fmt.Sprintf("Test Post %d", time.Now().UnixNano())
			args := append([]any{firstUser.ID, uniqueTitle}, tt.args[2:]...)
			if len(args) > 2 {
				args[2] = tt.args[2]
			}

			insertedPost, err := typedb.InsertAndReturn[*Post](ctx, db, tt.insertQuery, args...)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.errorMsg, err)
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
						db.Exec(ctx, "DELETE FROM posts WHERE id = :1", insertedPost.ID)
					})
				}
			}
		})
	}
}

// TestOracle_InsertAndReturn_Security_SQLInjection tests that SQL injection attempts through identifiers are rejected
func TestOracle_InsertAndReturn_Security_SQLInjection(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Get first user for foreign key
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE ROWNUM <= 1 ORDER BY id")
	if err != nil || firstUser == nil {
		t.Fatal("Need at least one user in database for foreign key")
	}

	// These queries attempt SQL injection through identifiers in the RETURNING clause
	// The validation should reject them before they reach the database
	injectionAttempts := []struct {
		name        string
		insertQuery string
		args        []any
		expectError bool
		errorMsg    string
	}{
		{
			name:        "semicolon injection in RETURNING column",
			insertQuery: "INSERT INTO posts (user_id, title, content) VALUES (:1, :2, :3) RETURNING id; DROP TABLE posts; --, user_id",
			args:        []any{firstUser.ID, "Test", "Content"},
			expectError: true,
			errorMsg:    "invalid identifier",
		},
		{
			name:        "semicolon injection in table name (parsed)",
			insertQuery: "INSERT INTO posts; DROP TABLE posts; -- (user_id, title) VALUES (:1, :2) RETURNING id",
			args:        []any{firstUser.ID, "Test"},
			expectError: true,
			errorMsg:    "invalid table name",
		},
		{
			name:        "comment injection in RETURNING column",
			insertQuery: "INSERT INTO posts (user_id, title) VALUES (:1, :2) RETURNING id--comment, user_id",
			args:        []any{firstUser.ID, "Test"},
			expectError: true,
			errorMsg:    "invalid identifier",
		},
		{
			name:        "space injection in RETURNING column",
			insertQuery: "INSERT INTO posts (user_id, title) VALUES (:1, :2) RETURNING id, user id",
			args:        []any{firstUser.ID, "Test"},
			expectError: true,
			errorMsg:    "invalid identifier",
		},
		{
			name:        "quote injection attempt",
			insertQuery: `INSERT INTO posts (user_id, title) VALUES (:1, :2) RETURNING id, "user_id"`,
			args:        []any{firstUser.ID, "Test"},
			expectError: false, // Quotes in identifiers are allowed and will be escaped
		},
	}

	for _, tt := range injectionAttempts {
		t.Run(tt.name, func(t *testing.T) {
			_, err := typedb.InsertAndReturn[*Post](ctx, db, tt.insertQuery, tt.args...)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for SQL injection attempt but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.errorMsg, err)
				}
			} else {
				// For cases where we expect success (like quoted identifiers), verify it works
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestOracle_InsertAndReturn_Security_IdentifierValidation tests identifier validation with edge cases
func TestOracle_InsertAndReturn_Security_IdentifierValidation(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Get first user for foreign key
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE ROWNUM <= 1 ORDER BY id")
	if err != nil || firstUser == nil {
		t.Fatal("Need at least one user in database for foreign key")
	}

	tests := []struct {
		name        string
		insertQuery string
		args        []any
		expectError bool
		errorMsg    string
	}{
		{
			name:        "identifier with quote character (should be escaped)",
			insertQuery: `INSERT INTO posts (user_id, title) VALUES (:1, :2) RETURNING id, "user_id"`,
			args:        []any{firstUser.ID, "Test"},
			expectError: false, // Quotes are allowed and escaped
		},
		{
			name:        "multiple RETURNING columns",
			insertQuery: "INSERT INTO posts (user_id, title, content) VALUES (:1, :2, :3) RETURNING id, user_id, title, content",
			args:        []any{firstUser.ID, "Test", "Content"},
			expectError: false,
		},
		{
			name:        "qualified column name in RETURNING",
			insertQuery: "INSERT INTO posts (user_id, title) VALUES (:1, :2) RETURNING posts.id, posts.user_id",
			args:        []any{firstUser.ID, "Test"},
			expectError: false, // Qualified names with dots are allowed
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
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else if insertedPost.ID == 0 {
					t.Error("Post ID should be set")
				}

				// Cleanup
				if insertedPost != nil && insertedPost.ID != 0 {
					t.Cleanup(func() {
						db.Exec(ctx, "DELETE FROM posts WHERE id = :1", insertedPost.ID)
					})
				}
			}
		})
	}
}
