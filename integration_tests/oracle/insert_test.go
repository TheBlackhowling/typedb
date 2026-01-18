package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/TheBlackHowling/typedb"
	_ "github.com/sijms/go-ora/v2" // Oracle driver
)

func TestOracle_Insert(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Insert new user with unique email
	uniqueEmail := fmt.Sprintf("testinsert-%d@example.com", time.Now().UnixNano())
	newUser := &User{
		Name:  "Test Insert User",
		Email: uniqueEmail,
	}
	if err := typedb.Insert(ctx, db, newUser); err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Verify ID was set
	if newUser.ID == 0 {
		t.Error("User ID should be set after insert")
	}

	// Register cleanup that runs even on failure
	t.Cleanup(func() {
		if newUser.ID != 0 {
			db.Exec(ctx, "DELETE FROM users WHERE id = :1", newUser.ID)
		}
	})

	// Verify user was inserted
	loaded := &User{ID: newUser.ID}
	if err := typedb.Load(ctx, db, loaded); err != nil {
		t.Fatalf("Failed to load inserted user: %v", err)
	}

	if loaded.Name != "Test Insert User" {
		t.Errorf("Expected name 'Test Insert User', got '%s'", loaded.Name)
	}
	if loaded.Email != uniqueEmail {
		t.Errorf("Expected email '%s', got '%s'", uniqueEmail, loaded.Email)
	}
}

func TestOracle_InsertAndReturn(t *testing.T) {
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

	// Insert post with RETURNING clause using unique title
	uniqueTitle := fmt.Sprintf("Test Post %d", time.Now().UnixNano())
	insertedPost, err := typedb.InsertAndReturn[*Post](ctx, db,
		"INSERT INTO posts (user_id, title, content, tags, metadata, created_at) VALUES (:1, :2, :3, :4, :5, TO_TIMESTAMP(:6, 'YYYY-MM-DD\"T\"HH24:MI:SS\"Z\"')) RETURNING id, user_id, title, content, tags, metadata, created_at",
		firstUser.ID, uniqueTitle, "Test content", `["go","database"]`, `{"test":true}`, "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("InsertAndReturn failed: %v", err)
	}

	// Register cleanup that runs even on failure
	t.Cleanup(func() {
		if insertedPost.ID != 0 {
			db.Exec(ctx, "DELETE FROM posts WHERE id = :1", insertedPost.ID)
		}
	})

	// Verify returned post
	if insertedPost.ID == 0 {
		t.Error("Post ID should be set")
	}
	if insertedPost.Title != uniqueTitle {
		t.Errorf("Expected title '%s', got '%s'", uniqueTitle, insertedPost.Title)
	}
	if insertedPost.UserID != firstUser.ID {
		t.Errorf("Expected UserID %d, got %d", firstUser.ID, insertedPost.UserID)
	}
}

func TestOracle_InsertAndGetId(t *testing.T) {
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

	// Insert post and get ID (Oracle uses RETURNING) with unique title
	uniqueTitle := fmt.Sprintf("Test Post ID %d", time.Now().UnixNano())
	postID, err := typedb.InsertAndGetId(ctx, db,
		"INSERT INTO posts (user_id, title, content, tags, metadata, created_at) VALUES (:1, :2, :3, :4, :5, TO_TIMESTAMP(:6, 'YYYY-MM-DD\"T\"HH24:MI:SS\"Z\"')) RETURNING id",
		firstUser.ID, uniqueTitle, "Test content", `["go"]`, `{"test":true}`, "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("InsertAndGetId failed: %v", err)
	}

	// Register cleanup that runs even on failure
	t.Cleanup(func() {
		if postID != 0 {
			db.Exec(ctx, "DELETE FROM posts WHERE id = :1", postID)
		}
	})

	// Verify ID was returned
	if postID == 0 {
		t.Error("Post ID should not be zero")
	}

	// Verify post exists
	loaded := &Post{ID: int(postID)}
	if err := typedb.Load(ctx, db, loaded); err != nil {
		t.Fatalf("Failed to load inserted post: %v", err)
	}

	if loaded.Title != uniqueTitle {
		t.Errorf("Expected title '%s', got '%s'", uniqueTitle, loaded.Title)
	}
}
