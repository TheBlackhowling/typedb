package main

import (
	"context"
	"testing"

	"github.com/TheBlackHowling/typedb"
	_ "github.com/microsoft/go-mssqldb" // SQL Server driver
)

func TestMSSQL_Insert(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

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

	// Clean up
	defer func() {
		db.Exec(ctx, "DELETE FROM users WHERE id = @p1", newUser.ID)
	}()

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

func TestMSSQL_InsertAndReturn(t *testing.T) {
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

	// Insert post with OUTPUT clause (MSSQL equivalent of RETURNING)
	insertedPost, err := typedb.InsertAndReturn[*Post](ctx, db,
		"INSERT INTO posts (user_id, title, content, tags, metadata, created_at) OUTPUT INSERTED.id, INSERTED.user_id, INSERTED.title, INSERTED.content, INSERTED.tags, INSERTED.metadata, INSERTED.created_at VALUES (@p1, @p2, @p3, @p4, @p5, @p6)",
		firstUser.ID, "Test Post", "Test content", `["go","database"]`, `{"test":true}`, "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("InsertAndReturn failed: %v", err)
	}

	// Clean up
	defer func() {
		db.Exec(ctx, "DELETE FROM posts WHERE id = @p1", insertedPost.ID)
	}()

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

func TestMSSQL_InsertAndGetId(t *testing.T) {
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

	// Insert post and get ID (MSSQL uses OUTPUT clause)
	postID, err := typedb.InsertAndGetId(ctx, db,
		"INSERT INTO posts (user_id, title, content, tags, metadata, created_at) OUTPUT INSERTED.id VALUES (@p1, @p2, @p3, @p4, @p5, @p6)",
		firstUser.ID, "Test Post ID", "Test content", `["go"]`, `{"test":true}`, "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("InsertAndGetId failed: %v", err)
	}

	// Clean up
	defer func() {
		db.Exec(ctx, "DELETE FROM posts WHERE id = @p1", postID)
	}()

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
