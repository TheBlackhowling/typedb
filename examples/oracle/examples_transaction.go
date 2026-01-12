package main

import (
	"context"
	"fmt"
	"log"

	"github.com/TheBlackHowling/typedb"
)

// Example12_Transaction demonstrates Transaction - Update multiple records atomically
func Example12_Transaction(ctx context.Context, db *typedb.DB, firstUser *User) {
	fmt.Println("\n--- Example 12: Transaction - Atomic Updates ---")
	err := db.WithTx(ctx, func(tx *typedb.Tx) error {
		// Update user
		userInTx := &User{ID: firstUser.ID, Name: "Transaction User"}
		if err := typedb.Update(ctx, tx, userInTx); err != nil {
			return err
		}

		// Insert post in same transaction
		txPost := &Post{
			UserID:    firstUser.ID,
			Title:     "Transaction Post",
			Content:   "Created in a transaction",
			Published: true,
			// CreatedAt will use database default (CURRENT_TIMESTAMP)
		}
		if err := typedb.Insert(ctx, tx, txPost); err != nil {
			return err
		}
		return nil
	}, nil)
	if err != nil {
		log.Fatalf("Transaction failed: %v", err)
	}
	fmt.Println("  ✓ Transaction completed successfully")
}

// Example12b_TransactionRollback demonstrates Transaction rollback on error
func Example12b_TransactionRollback(ctx context.Context, db *typedb.DB, firstUser *User) {
	fmt.Println("\n--- Example 12b: Transaction Rollback - Automatic Rollback on Error ---")

	// Get original user name for verification
	originalUser := &User{ID: firstUser.ID}
	if err := typedb.Load(ctx, db, originalUser); err != nil {
		log.Fatalf("Failed to load user: %v", err)
	}
	originalName := originalUser.Name

	// Attempt transaction that will fail
	err := db.WithTx(ctx, func(tx *typedb.Tx) error {
		// Update user (this succeeds)
		userInTx := &User{ID: firstUser.ID, Name: "Rollback Test Name"}
		if err := typedb.Update(ctx, tx, userInTx); err != nil {
			return err
		}

		// Try to insert post with invalid user_id (this will fail due to foreign key constraint)
		invalidPost := &Post{
			UserID:    999999, // Non-existent user ID
			Title:     "This Will Fail",
			Content:   "Foreign key violation",
			Published: true,
			// CreatedAt will use database default (CURRENT_TIMESTAMP)
		}
		if err := typedb.Insert(ctx, tx, invalidPost); err != nil {
			return err // Return error to trigger rollback
		}
		return nil
	}, nil)

	// Transaction should have failed
	if err == nil {
		log.Fatalf("Expected transaction to fail, but it succeeded")
	}
	fmt.Printf("  ✓ Transaction failed as expected: %v\n", err)

	// Verify rollback: user name should be unchanged
	rolledBackUser := &User{ID: firstUser.ID}
	if err := typedb.Load(ctx, db, rolledBackUser); err != nil {
		log.Fatalf("Failed to load user after rollback: %v", err)
	}
	if rolledBackUser.Name != originalName {
		log.Fatalf("Rollback verification failed: expected name '%s', got '%s'", originalName, rolledBackUser.Name)
	}
	fmt.Printf("  ✓ Rollback verified: user name is still '%s' (unchanged)\n", rolledBackUser.Name)
}

// runTransactionExamples demonstrates transaction operations.
func runTransactionExamples(ctx context.Context, db *typedb.DB, firstUser *User) {
	Example12_Transaction(ctx, db, firstUser)
	Example12b_TransactionRollback(ctx, db, firstUser)
}
