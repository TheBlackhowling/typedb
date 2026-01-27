package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/TheBlackHowling/typedb"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

// TestMySQL_Insert_Security_ValidIdentifiers tests Insert with various valid identifier formats
// MySQL doesn't support RETURNING, so we test Insert which uses quoteIdentifier internally
func TestMySQL_Insert_Security_ValidIdentifiers(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Get first user for foreign key
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users ORDER BY id LIMIT 1")
	if err != nil || firstUser == nil {
		t.Fatal("Need at least one user in database for foreign key")
	}

	// Test Insert with valid identifiers (table and column names come from struct tags)
	// This indirectly tests that quoteIdentifier works correctly
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
			db.Exec(ctx, "DELETE FROM posts WHERE id = ?", newPost.ID)
		})
	}

	// Verify the insert worked
	if newPost.ID == 0 {
		t.Error("Post ID should be set after insert")
	}
}

// TestMySQL_InsertAndGetID_Security tests InsertAndGetID which uses quoteIdentifier internally
func TestMySQL_InsertAndGetID_Security(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users ORDER BY id LIMIT 1")
	if err != nil || firstUser == nil {
		t.Fatal("Need at least one user in database for foreign key")
	}

	// Test InsertAndGetID with valid query
	uniqueTitle := fmt.Sprintf("Test Post %d", time.Now().UnixNano())
	postID, err := typedb.InsertAndGetID(ctx, db,
		"INSERT INTO posts (user_id, title, content, tags, metadata, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		firstUser.ID, uniqueTitle, "Test content", `["go","database"]`, `{"test":true}`, "2024-01-01 00:00:00")
	if err != nil {
		t.Fatalf("InsertAndGetID failed: %v", err)
	}

	// Cleanup
	if postID != 0 {
		t.Cleanup(func() {
			db.Exec(ctx, "DELETE FROM posts WHERE id = ?", postID)
		})
	}

	// Verify ID was returned
	if postID == 0 {
		t.Error("Post ID should not be zero")
	}
}
