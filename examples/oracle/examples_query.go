package main

import (
	"context"
	"fmt"
	"log"

	"github.com/TheBlackHowling/typedb"
)

// Example1_QueryAll demonstrates QueryAll - Get all users
func Example1_QueryAll(ctx context.Context, db *typedb.DB) {
	fmt.Println("\n--- Example 1: QueryAll - Get All Users ---")
	users, err := typedb.QueryAll[*User](ctx, db, "SELECT * FROM (SELECT id, name, email, created_at FROM users ORDER BY id) WHERE ROWNUM <= 5")
	if err != nil {
		log.Fatalf("Failed to query users: %v", err)
	}
	for _, user := range users {
		fmt.Printf("  User: %s (%s) - ID: %d\n", user.Name, user.Email, user.ID)
	}
}

// Example2_QueryFirst demonstrates QueryFirst - Get first user
func Example2_QueryFirst(ctx context.Context, db *typedb.DB) *User {
	fmt.Println("\n--- Example 2: QueryFirst - Get First User ---")
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT * FROM (SELECT id, name, email, created_at FROM users ORDER BY id) WHERE ROWNUM <= 1")
	if err != nil {
		log.Fatalf("Failed to query first user: %v", err)
	}
	if firstUser != nil {
		fmt.Printf("  First user: %s (%s)\n", firstUser.Name, firstUser.Email)
	}
	return firstUser
}

// Example3_QueryOne demonstrates QueryOne - Get exactly one user
func Example3_QueryOne(ctx context.Context, db *typedb.DB, firstUser *User) {
	fmt.Println("\n--- Example 3: QueryOne - Get One User ---")
	oneUser, err := typedb.QueryOne[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = :1", firstUser.ID)
	if err != nil {
		log.Fatalf("Failed to query one user: %v", err)
	}
	fmt.Printf("  Found user: %s (%s)\n", oneUser.Name, oneUser.Email)
}

// Example10_QueryAllWithJoins demonstrates QueryAll with joins - Get posts with user info
func Example10_QueryAllWithJoins(ctx context.Context, db *typedb.DB) {
	fmt.Println("\n--- Example 10: QueryAll with Joins - Posts with User Info ---")
	posts, err := typedb.QueryAll[*Post](ctx, db, `
		SELECT * FROM (
			SELECT p.id, p.user_id, p.title, p.content, p.published, p.created_at 
			FROM posts p 
			WHERE p.published = :1 
			ORDER BY p.created_at DESC
		) WHERE ROWNUM <= 5
	`, true)
	if err != nil {
		log.Fatalf("Failed to query posts: %v", err)
	}
	for _, post := range posts {
		fmt.Printf("  Post: %s (User ID: %d, Published: %v)\n", post.Title, post.UserID, post.Published)
	}
}

// runQueryExamples demonstrates QueryAll, QueryFirst, and QueryOne operations.
// Returns the first user for use in subsequent examples.
func runQueryExamples(ctx context.Context, db *typedb.DB) *User {
	Example1_QueryAll(ctx, db)
	firstUser := Example2_QueryFirst(ctx, db)
	Example3_QueryOne(ctx, db, firstUser)
	Example10_QueryAllWithJoins(ctx, db)
	return firstUser
}
