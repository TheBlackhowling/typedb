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

// Example10_Update_AutoTimestamp demonstrates Update with auto-populated timestamp
func Example10_Update_AutoTimestamp(ctx context.Context, db *typedb.DB, firstUser *User) {
	fmt.Println("\n--- Example 10: Update - Auto-Timestamp ---")
	// Note: This example requires the User model to have an UpdatedAt field with dbUpdate:"auto-timestamp" tag
	// and the database table to have an updated_at column
	
	// Load user first to get the initial UpdatedAt value
	userBeforeUpdate := &User{ID: firstUser.ID}
	if err := typedb.Load(ctx, db, userBeforeUpdate); err != nil {
		log.Fatalf("Failed to load user before update: %v", err)
	}
	originalUpdatedAt := userBeforeUpdate.UpdatedAt
	fmt.Printf("  Original UpdatedAt: %s\n", originalUpdatedAt)
	
	// Update user - UpdatedAt will be automatically populated with CURRENT_TIMESTAMP
	userToUpdate := &User{
		ID:   firstUser.ID,
		Name: "Updated Name with Auto Timestamp",
		// UpdatedAt is not set - will be auto-populated by database
	}
	if err := typedb.Update(ctx, db, userToUpdate); err != nil {
		log.Fatalf("Failed to update user: %v", err)
	}
	
	// Reload user to verify UpdatedAt was changed
	updatedUser := &User{ID: firstUser.ID}
	if err := typedb.Load(ctx, db, updatedUser); err != nil {
		log.Fatalf("Failed to load updated user: %v", err)
	}
	fmt.Printf("  ✓ Updated user name to: %s\n", updatedUser.Name)
	fmt.Printf("  New UpdatedAt: %s\n", updatedUser.UpdatedAt)
	
	// Verify UpdatedAt changed (should be different from original)
	if updatedUser.UpdatedAt == originalUpdatedAt {
		log.Fatalf("UpdatedAt should have changed, but it's still: %s", updatedUser.UpdatedAt)
	}
	fmt.Printf("  ✓ UpdatedAt was automatically updated (changed from original value)\n")
}

// runUpdateExamples demonstrates Update operations.
func runUpdateExamples(ctx context.Context, db *typedb.DB, firstUser *User) {
	Example9_Update(ctx, db, firstUser)
	Example10_Update_AutoTimestamp(ctx, db, firstUser)
}
