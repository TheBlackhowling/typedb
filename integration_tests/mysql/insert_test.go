package main

import (
	"context"
	"testing"

	"github.com/TheBlackHowling/typedb"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

func TestMySQL_Insert(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
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
		db.Exec(ctx, "DELETE FROM users WHERE id = ?", newUser.ID)
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

func TestMySQL_InsertAndLoad(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
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

	// Insert post using InsertAndLoad
	newPost := &Post{
		UserID:   firstUser.ID,
		Title:    "Test Post",
		Content:  "Test content",
		Tags:     `["go","database"]`,
		Metadata: `{"test":true}`,
	}
	returnedPost, err := typedb.InsertAndLoad[*Post](ctx, db, newPost)
	if err != nil {
		t.Fatalf("InsertAndLoad failed: %v", err)
	}

	// Clean up
	defer func() {
		db.Exec(ctx, "DELETE FROM posts WHERE id = ?", returnedPost.ID)
	}()

	// Verify returned post is fully populated
	if returnedPost.ID == 0 {
		t.Error("Post ID should be set")
	}
	if returnedPost.Title != "Test Post" {
		t.Errorf("Expected title 'Test Post', got '%s'", returnedPost.Title)
	}
	if returnedPost.UserID != firstUser.ID {
		t.Errorf("Expected UserID %d, got %d", firstUser.ID, returnedPost.UserID)
	}
	if returnedPost.CreatedAt == "" {
		t.Error("CreatedAt should be populated from database")
	}
}

func TestMySQL_InsertAndGetId(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
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

	// Insert post and get ID (MySQL uses LastInsertId)
	postID, err := typedb.InsertAndGetId(ctx, db,
		"INSERT INTO posts (user_id, title, content, tags, metadata, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		firstUser.ID, "Test Post ID", "Test content", `["go"]`, `{"test":true}`, "2024-01-01 00:00:00")
	if err != nil {
		t.Fatalf("InsertAndGetId failed: %v", err)
	}

	// Clean up
	defer func() {
		db.Exec(ctx, "DELETE FROM posts WHERE id = ?", postID)
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

func TestMySQL_InsertAndGetId_MissingIdColumn(t *testing.T) {
	// MySQL doesn't support RETURNING clause in INSERT statements (or uses LastInsertId by default)
	// The "missing ID column" error only applies to databases that use RETURNING/OUTPUT clauses.
	// MySQL uses LastInsertId() path which doesn't have this error case.
	// MySQL 8.0.19+ supports RETURNING, but the test environment may not have it,
	// and even if it does, MySQL would throw a SQL syntax error before reaching the missing ID check.
	t.Skip("MySQL uses LastInsertId() path without RETURNING clause, so missing ID column error doesn't apply")
}