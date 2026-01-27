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
	defer closeDB(t, db)

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

func TestOracle_InsertAndLoad(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Get first user for foreign key
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE ROWNUM <= 1 ORDER BY id")
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

	// Register cleanup that runs even on failure
	t.Cleanup(func() {
		if returnedPost.ID != 0 {
			db.Exec(ctx, "DELETE FROM posts WHERE id = :1", returnedPost.ID)
		}
	})

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

func TestOracle_InsertAndGetID(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

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
	postID, err := typedb.InsertAndGetID(ctx, db,
		"INSERT INTO posts (user_id, title, content, tags, metadata, created_at) VALUES (:1, :2, :3, TO_CLOB(:4), TO_CLOB(:5), TO_TIMESTAMP(:6, 'YYYY-MM-DD\"T\"HH24:MI:SS\"Z\"')) RETURNING id",
		firstUser.ID, uniqueTitle, "Test content", "go,database", `{"test":true}`, "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("InsertAndGetID failed: %v", err)
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

func TestOracle_InsertAndGetID_WithInto(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Get first user for foreign key
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE ROWNUM <= 1 ORDER BY id")
	if err != nil || firstUser == nil {
		t.Fatal("Need at least one user in database for foreign key")
	}

	// Insert post and get ID using RETURNING ... INTO clause (Path A: already has INTO)
	uniqueTitle := fmt.Sprintf("Test Post WITH INTO %d", time.Now().UnixNano())
	postID, err := typedb.InsertAndGetID(ctx, db,
		"INSERT INTO posts (user_id, title, content, tags, metadata, created_at) VALUES (:1, :2, :3, TO_CLOB(:4), TO_CLOB(:5), TO_TIMESTAMP(:6, 'YYYY-MM-DD\"T\"HH24:MI:SS\"Z\"')) RETURNING id INTO :7",
		firstUser.ID, uniqueTitle, "Test content", "go,database", `{"test":true}`, "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("InsertAndGetID failed: %v", err)
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
