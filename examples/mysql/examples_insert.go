package main

import (
	"context"
	"fmt"
	"log"

	"github.com/TheBlackHowling/typedb"
)

// Example6_Insert demonstrates Insert - Insert a new user
func Example6_Insert(ctx context.Context, db *typedb.DB) {
	fmt.Println("\n--- Example 6: Insert - Insert New User ---")
	newUser := &User{
		Name:      "Example User",
		Email:     "example@example.com",
		CreatedAt: "2024-01-01 00:00:00",
	}
	if err := typedb.Insert(ctx, db, newUser); err != nil {
		log.Fatalf("Failed to insert user: %v", err)
	}
	fmt.Printf("  ✓ Inserted new user: %s\n", newUser.Email)
}

// Example7_InsertAndReturn demonstrates InsertAndReturn - Insert and get the full record
// Note: MySQL doesn't support RETURNING clause, so we use InsertAndGetId + Load instead
func Example7_InsertAndReturn(ctx context.Context, db *typedb.DB, firstUser *User) {
	fmt.Println("\n--- Example 7: InsertAndReturn (MySQL workaround) - Insert and Get Full Record ---")
	newPost := &Post{
		UserID:    firstUser.ID,
		Title:     "Example Post",
		Content:   "This is an example post created with InsertAndGetId",
		Published: true,
		CreatedAt: "2024-01-01 00:00:00",
	}
	postID, err := typedb.InsertAndGetId(ctx, db,
		"INSERT INTO posts (user_id, title, content, published, created_at) VALUES (?, ?, ?, ?, ?)",
		newPost.UserID, newPost.Title, newPost.Content, newPost.Published, newPost.CreatedAt)
	if err != nil {
		log.Fatalf("Failed to insert and get ID: %v", err)
	}
	insertedPost := &Post{ID: int(postID)}
	if err := typedb.Load(ctx, db, insertedPost); err != nil {
		log.Fatalf("Failed to load inserted post: %v", err)
	}
	fmt.Printf("  ✓ Inserted post: %s (ID: %d)\n", insertedPost.Title, insertedPost.ID)
}

// Example8_InsertAndGetId demonstrates InsertAndGetId - Insert and get just the ID
func Example8_InsertAndGetId(ctx context.Context, db *typedb.DB, firstUser *User) int64 {
	fmt.Println("\n--- Example 8: InsertAndGetId - Insert and Get ID ---")
	anotherPost := &Post{
		UserID:    firstUser.ID,
		Title:     "Another Post",
		Content:   "This post uses InsertAndGetId",
		Published: false,
		CreatedAt: "2024-01-01 00:00:00",
	}
	postID, err := typedb.InsertAndGetId(ctx, db,
		"INSERT INTO posts (user_id, title, content, published, created_at) VALUES (?, ?, ?, ?, ?)",
		anotherPost.UserID, anotherPost.Title, anotherPost.Content, anotherPost.Published, anotherPost.CreatedAt)
	if err != nil {
		log.Fatalf("Failed to insert and get ID: %v", err)
	}
	fmt.Printf("  ✓ Inserted post with ID: %d\n", postID)
	return postID
}

// runInsertExamples demonstrates Insert, InsertAndReturn, and InsertAndGetId operations.
// Returns the post ID for use in subsequent examples.
func runInsertExamples(ctx context.Context, db *typedb.DB, firstUser *User) int64 {
	Example6_Insert(ctx, db)
	Example7_InsertAndReturn(ctx, db, firstUser)
	return Example8_InsertAndGetId(ctx, db, firstUser)
}
