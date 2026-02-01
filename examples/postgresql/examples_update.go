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

// Example11_Update_PartialUpdate demonstrates partial update functionality
func Example11_Update_PartialUpdate(ctx context.Context, db *typedb.DB, firstUser *User) {
	fmt.Println("\n--- Example 11: Update - Partial Update ---")
	// Note: This example requires the User model to be registered with RegisterModelWithOptions
	// and ModelOptions{PartialUpdate: true}

	// Load user first to save original copy (required for partial update)
	userToUpdate := &User{ID: firstUser.ID}
	if err := typedb.Load(ctx, db, userToUpdate); err != nil {
		log.Fatalf("Failed to load user: %v", err)
	}

	originalName := userToUpdate.Name
	originalEmail := userToUpdate.Email
	fmt.Printf("  Original Name: %s\n", originalName)
	fmt.Printf("  Original Email: %s\n", originalEmail)

	// Modify only name, keep email unchanged
	userToUpdate.Name = "Partially Updated Name"
	// Email remains unchanged - will NOT be included in UPDATE

	if err := typedb.Update(ctx, db, userToUpdate); err != nil {
		log.Fatalf("Failed to update user: %v", err)
	}

	// Reload to verify only name was updated
	updatedUser := &User{ID: firstUser.ID}
	if err := typedb.Load(ctx, db, updatedUser); err != nil {
		log.Fatalf("Failed to load updated user: %v", err)
	}

	fmt.Printf("  ✓ Updated name to: %s\n", updatedUser.Name)
	fmt.Printf("  ✓ Email remained unchanged: %s\n", updatedUser.Email)

	// Verify email was not changed
	if updatedUser.Email != originalEmail {
		log.Fatalf("Email should not have changed, but it did. Original: %s, New: %s", originalEmail, updatedUser.Email)
	}
	fmt.Printf("  ✓ Partial update successful - only changed fields were updated\n")
}

// Example12_Update_SetToNil_RawSQL demonstrates setting a field to NULL using raw SQL
// This is required when partial update is disabled or for non-pointer fields
func Example12_Update_SetToNil_RawSQL(ctx context.Context, db *typedb.DB, firstUser *User) {
	fmt.Println("\n--- Example 12: Update - Set Field to NULL (Raw SQL) ---")
	// Note: This example shows how to set a field to NULL when partial update is disabled
	// or when you need explicit control over NULL values

	// Load user first to see current phone value
	userBeforeUpdate := &User{ID: firstUser.ID}
	if err := typedb.Load(ctx, db, userBeforeUpdate); err != nil {
		log.Fatalf("Failed to load user: %v", err)
	}
	if userBeforeUpdate.Phone != nil {
		fmt.Printf("  Current Phone: %s\n", *userBeforeUpdate.Phone)
	} else {
		fmt.Printf("  Current Phone: NULL\n")
	}

	// Set phone to NULL using raw SQL
	// This is necessary when partial update is disabled or for explicit NULL updates
	_, err := db.Exec(ctx, "UPDATE users SET phone = NULL WHERE id = $1", firstUser.ID)
	if err != nil {
		log.Fatalf("Failed to set phone to NULL: %v", err)
	}

	// Verify phone was set to NULL
	userAfterUpdate := &User{ID: firstUser.ID}
	if err := typedb.Load(ctx, db, userAfterUpdate); err != nil {
		log.Fatalf("Failed to load updated user: %v", err)
	}
	if userAfterUpdate.Phone == nil {
		fmt.Printf("  ✓ Phone successfully set to NULL\n")
	} else {
		log.Fatalf("Phone should be NULL but got: %s", *userAfterUpdate.Phone)
	}
}

// Example13_Update_SetToNil_PartialUpdate demonstrates setting a field to nil with partial update enabled
// When partial update is enabled, you can set a pointer field to nil and it will be included in the UPDATE
func Example13_Update_SetToNil_PartialUpdate(ctx context.Context, db *typedb.DB, firstUser *User) {
	fmt.Println("\n--- Example 13: Update - Set Field to NULL (Partial Update) ---")
	// Note: This example requires partial update to be enabled for the User model
	// (already enabled in models.go with RegisterModelWithOptions)

	// First, ensure phone has a value (set it if it's NULL)
	userToPrepare := &User{ID: firstUser.ID}
	if err := typedb.Load(ctx, db, userToPrepare); err != nil {
		log.Fatalf("Failed to load user: %v", err)
	}
	if userToPrepare.Phone == nil {
		// Set phone to a value first so we can demonstrate setting it to nil
		phoneValue := "123-456-7890"
		userToPrepare.Phone = &phoneValue
		if err := typedb.Update(ctx, db, userToPrepare); err != nil {
			log.Fatalf("Failed to set phone: %v", err)
		}
		fmt.Printf("  Set phone to: %s\n", *userToPrepare.Phone)
	} else {
		fmt.Printf("  Current Phone: %s\n", *userToPrepare.Phone)
	}

	// Load user again to save original state (required for partial update)
	userToUpdate := &User{ID: firstUser.ID}
	if err := typedb.Load(ctx, db, userToUpdate); err != nil {
		log.Fatalf("Failed to load user: %v", err)
	}

	// Set phone to nil - this will be detected as a change
	userToUpdate.Phone = nil

	// Update - phone will be included in UPDATE and set to NULL
	// This works because:
	// 1. Partial update is enabled
	// 2. Phone was non-nil before (detected as changed)
	// 3. Phone is now nil (will be included in UPDATE)
	if err := typedb.Update(ctx, db, userToUpdate); err != nil {
		log.Fatalf("Failed to update user: %v", err)
	}

	// Verify phone was set to NULL
	updatedUser := &User{ID: firstUser.ID}
	if err := typedb.Load(ctx, db, updatedUser); err != nil {
		log.Fatalf("Failed to load updated user: %v", err)
	}
	if updatedUser.Phone == nil {
		fmt.Printf("  ✓ Phone successfully set to NULL using partial update\n")
	} else {
		log.Fatalf("Phone should be NULL but got: %s", *updatedUser.Phone)
	}
}

// runUpdateExamples demonstrates Update operations.
func runUpdateExamples(ctx context.Context, db *typedb.DB, firstUser *User) {
	Example9_Update(ctx, db, firstUser)
	Example10_Update_AutoTimestamp(ctx, db, firstUser)
	Example11_Update_PartialUpdate(ctx, db, firstUser)
	Example12_Update_SetToNil_RawSQL(ctx, db, firstUser)
	Example13_Update_SetToNil_PartialUpdate(ctx, db, firstUser)
}
