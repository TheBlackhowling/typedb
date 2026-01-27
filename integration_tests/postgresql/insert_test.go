package main

import (
	"context"
	"testing"

	"github.com/TheBlackHowling/typedb"
	_ "github.com/lib/pq" // PostgreSQL driver
)

func TestPostgreSQL_Insert(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

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
		db.Exec(ctx, "DELETE FROM users WHERE id = $1", newUser.ID)
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

func TestPostgreSQL_InsertAndLoad(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
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

	// Insert post using InsertAndLoad
	newPost := &Post{
		UserID:   firstUser.ID,
		Title:    "Test Post",
		Content:  "Test content",
		Tags:     `{"go","database"}`,
		Metadata: `{"test":true}`,
	}
	returnedPost, err := typedb.InsertAndLoad[*Post](ctx, db, newPost)
	if err != nil {
		t.Fatalf("InsertAndLoad failed: %v", err)
	}

	// Clean up
	defer func() {
		db.Exec(ctx, "DELETE FROM posts WHERE id = $1", returnedPost.ID)
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

func TestPostgreSQL_InsertAndGetID(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
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

	// Insert post and get ID
	postID, err := typedb.InsertAndGetID(ctx, db,
		"INSERT INTO posts (user_id, title, content, tags, metadata, created_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		firstUser.ID, "Test Post ID", "Test content", "{\"go\"}", "{\"test\":true}", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("InsertAndGetID failed: %v", err)
	}

	// Clean up
	defer func() {
		db.Exec(ctx, "DELETE FROM posts WHERE id = $1", postID)
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

func TestPostgreSQL_InsertAndGetID_MissingIdColumn(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
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

	// Insert post with RETURNING clause that doesn't return 'id' column
	_, err = typedb.InsertAndGetID(ctx, db,
		"INSERT INTO posts (user_id, title, content, tags, metadata, created_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING title",
		firstUser.ID, "Test Post", "Test content", "{\"go\"}", "{\"test\":true}", "2024-01-01T00:00:00Z")

	if err == nil {
		t.Fatal("Expected error for missing ID column")
	}

	expectedError := "typedb: InsertAndGetID RETURNING/OUTPUT clause did not return 'id' column"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}
