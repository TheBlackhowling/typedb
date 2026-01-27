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
		Name:  "Example User",
		Email: "example@example.com",
		// CreatedAt will use database default (CURRENT_TIMESTAMP)
	}
	if err := typedb.Insert(ctx, db, newUser); err != nil {
		log.Fatalf("Failed to insert user: %v", err)
	}
	fmt.Printf("  ✓ Inserted new user: %s\n", newUser.Email)
}

// Example7_InsertAndLoad demonstrates InsertAndLoad - Insert and get the full record
func Example7_InsertAndLoad(ctx context.Context, db *typedb.DB, firstUser *User) {
	fmt.Println("\n--- Example 7: InsertAndLoad - Insert and Get Full Record ---")
	newPost := &Post{
		UserID:    firstUser.ID,
		Title:     "Example Post",
		Content:   "This is an example post created with InsertAndLoad",
		Published: true,
		// CreatedAt will use database default (CURRENT_TIMESTAMP)
	}
	returnedPost, err := typedb.InsertAndLoad[*Post](ctx, db, newPost)
	if err != nil {
		log.Fatalf("Failed to insert and load post: %v", err)
	}
	fmt.Printf("  ✓ Inserted post: %s (ID: %d, CreatedAt: %s)\n", returnedPost.Title, returnedPost.ID, returnedPost.CreatedAt)
}

// Example8_InsertAndGetID demonstrates InsertAndGetID - Insert and get just the ID
func Example8_InsertAndGetID(ctx context.Context, db *typedb.DB, firstUser *User) int64 {
	fmt.Println("\n--- Example 8: InsertAndGetID - Insert and Get ID ---")
	anotherPost := &Post{
		UserID:    firstUser.ID,
		Title:     "Another Post",
		Content:   "This post uses InsertAndGetID",
		Published: false,
		// CreatedAt will use database default (CURRENT_TIMESTAMP)
	}
	postID, err := typedb.InsertAndGetID(ctx, db,
		"INSERT INTO posts (user_id, title, content, published) VALUES (:1, :2, :3, :4) RETURNING id",
		anotherPost.UserID, anotherPost.Title, anotherPost.Content, anotherPost.Published)
	if err != nil {
		log.Fatalf("Failed to insert and get ID: %v", err)
	}
	fmt.Printf("  ✓ Inserted post with ID: %d\n", postID)
	return postID
}

// runInsertExamples demonstrates Insert, InsertAndLoad, and InsertAndGetID operations.
// Returns the post ID for use in subsequent examples.
func runInsertExamples(ctx context.Context, db *typedb.DB, firstUser *User) int64 {
	Example6_Insert(ctx, db)
	Example7_InsertAndLoad(ctx, db, firstUser)
	return Example8_InsertAndGetID(ctx, db, firstUser)
}
