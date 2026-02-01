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

// Example12_Update_SetToNil_RawSQL demonstrates setting a field to NULL using raw SQL
func Example12_Update_SetToNil_RawSQL(ctx context.Context, db *typedb.DB, firstUser *User) {
	fmt.Println("\n--- Example 12: Update - Set Field to NULL (Raw SQL) ---")
	userBeforeUpdate := &User{ID: firstUser.ID}
	if err := typedb.Load(ctx, db, userBeforeUpdate); err != nil {
		log.Fatalf("Failed to load user: %v", err)
	}
	if userBeforeUpdate.Phone != nil {
		fmt.Printf("  Current Phone: %s\n", *userBeforeUpdate.Phone)
	} else {
		fmt.Printf("  Current Phone: NULL\n")
	}
	_, err := db.Exec(ctx, "UPDATE users SET phone = NULL WHERE id = ?", firstUser.ID)
	if err != nil {
		log.Fatalf("Failed to set phone to NULL: %v", err)
	}
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
func Example13_Update_SetToNil_PartialUpdate(ctx context.Context, db *typedb.DB, firstUser *User) {
	fmt.Println("\n--- Example 13: Update - Set Field to NULL (Partial Update) ---")
	userToPrepare := &User{ID: firstUser.ID}
	if err := typedb.Load(ctx, db, userToPrepare); err != nil {
		log.Fatalf("Failed to load user: %v", err)
	}
	if userToPrepare.Phone == nil {
		phoneValue := "123-456-7890"
		userToPrepare.Phone = &phoneValue
		if err := typedb.Update(ctx, db, userToPrepare); err != nil {
			log.Fatalf("Failed to set phone: %v", err)
		}
		fmt.Printf("  Set phone to: %s\n", *userToPrepare.Phone)
	} else {
		fmt.Printf("  Current Phone: %s\n", *userToPrepare.Phone)
	}
	userToUpdate := &User{ID: firstUser.ID}
	if err := typedb.Load(ctx, db, userToUpdate); err != nil {
		log.Fatalf("Failed to load user: %v", err)
	}
	userToUpdate.Phone = nil
	if err := typedb.Update(ctx, db, userToUpdate); err != nil {
		log.Fatalf("Failed to update user: %v", err)
	}
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
	Example12_Update_SetToNil_RawSQL(ctx, db, firstUser)
	Example13_Update_SetToNil_PartialUpdate(ctx, db, firstUser)
}
