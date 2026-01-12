package main

import (
	"context"
	"fmt"
	"log"

	"github.com/TheBlackHowling/typedb"
)

// Example9_Update demonstrates Update - Update a user
func Example9_Update(ctx context.Context, db *typedb.DB, firstUser *User) {
	fmt.Println("\n--- Example 9: Update - Update User ---")
	userToUpdate := &User{
		ID:   firstUser.ID,
		Name: "Updated Name",
	}
	if err := typedb.Update(ctx, db, userToUpdate); err != nil {
		log.Fatalf("Failed to update user: %v", err)
	}
	// Verify update
	updatedUser := &User{ID: firstUser.ID}
	if err := typedb.Load(ctx, db, updatedUser); err != nil {
		log.Fatalf("Failed to load updated user: %v", err)
	}
	fmt.Printf("  ✓ Updated user name to: %s\n", updatedUser.Name)
}

// runUpdateExamples demonstrates Update operations.
func runUpdateExamples(ctx context.Context, db *typedb.DB, firstUser *User) {
	Example9_Update(ctx, db, firstUser)
}
