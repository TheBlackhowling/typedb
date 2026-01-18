package main

import (
	"context"
	"os"
	"testing"

	"github.com/TheBlackHowling/typedb"
)

func TestSQLite_Insert(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer os.Remove(getTestDSN())

	ctx := context.Background()

	// Insert new user
	newUser := &User{
		Name:  "Test Insert User",
		Email: "testinsert@example.com",
	}
	if err := typedb.Insert(ctx, db, newUser); err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Verify ID was set
	if newUser.ID == 0 {
		t.Error("User ID should be set after insert")
	}

	// Verify user was inserted
	loaded := &User{ID: newUser.ID}
	if err := typedb.Load(ctx, db, loaded); err != nil {
		t.Fatalf("Failed to load inserted user: %v", err)
	}

	if loaded.Name != "Test Insert User" {
		t.Errorf("Expected name 'Test Insert User', got '%s'", loaded.Name)
	}
	if loaded.Email != "testinsert@example.com" {
		t.Errorf("Expected email 'testinsert@example.com', got '%s'", loaded.Email)
	}
}

func TestSQLite_InsertAndReturn(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer os.Remove(getTestDSN())

	ctx := context.Background()

	// Get first user for foreign key
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users ORDER BY id LIMIT 1")
	if err != nil || firstUser == nil {
		t.Fatal("Need at least one user in database for foreign key")
	}

	// Insert post with RETURNING clause (SQLite supports RETURNING)
	insertedPost, err := typedb.InsertAndReturn[*Post](ctx, db,
		"INSERT INTO posts (user_id, title, content, tags, metadata, created_at) VALUES (?, ?, ?, ?, ?, ?) RETURNING id, user_id, title, content, tags, metadata, created_at",
		firstUser.ID, "Test Post", "Test content", `["go","database"]`, `{"test":true}`, "2024-01-01 00:00:00")
	if err != nil {
		t.Fatalf("InsertAndReturn failed: %v", err)
	}

	// Verify returned post
	if insertedPost.ID == 0 {
		t.Error("Post ID should be set")
	}
	if insertedPost.Title != "Test Post" {
		t.Errorf("Expected title 'Test Post', got '%s'", insertedPost.Title)
	}
	if insertedPost.UserID != firstUser.ID {
		t.Errorf("Expected UserID %d, got %d", firstUser.ID, insertedPost.UserID)
	}
}

func TestSQLite_InsertAndGetId(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer os.Remove(getTestDSN())

	ctx := context.Background()

	// Get first user for foreign key
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users ORDER BY id LIMIT 1")
	if err != nil || firstUser == nil {
		t.Fatal("Need at least one user in database for foreign key")
	}

	// Insert post and get ID (SQLite supports RETURNING)
	postID, err := typedb.InsertAndGetId(ctx, db,
		"INSERT INTO posts (user_id, title, content, tags, metadata, created_at) VALUES (?, ?, ?, ?, ?, ?) RETURNING id",
		firstUser.ID, "Test Post ID", "Test content", `["go"]`, `{"test":true}`, "2024-01-01 00:00:00")
	if err != nil {
		t.Fatalf("InsertAndGetId failed: %v", err)
	}

	// Verify ID was returned
	if postID == 0 {
		t.Error("Post ID should not be zero")
	}

	// Verify post exists
	loaded := &Post{ID: int(postID)}
	if err := typedb.Load(ctx, db, loaded); err != nil {
		t.Fatalf("Failed to load inserted post: %v", err)
	}

	if loaded.Title != "Test Post ID" {
		t.Errorf("Expected title 'Test Post ID', got '%s'", loaded.Title)
	}
}
