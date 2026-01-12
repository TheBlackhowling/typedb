package main

import (
	"context"
	"fmt"
	"log"

	"github.com/TheBlackHowling/typedb"
)

// Example13_QueryAll demonstrates DB.QueryAll - Raw query returning maps
func Example13_QueryAll(ctx context.Context, db *typedb.DB) {
	fmt.Println("\n--- Example 13: DB.QueryAll - Raw Query as Maps ---")
	rows, err := db.QueryAll(ctx, "SELECT id, name, email FROM users ORDER BY id LIMIT 3")
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}
	for _, row := range rows {
		fmt.Printf("  User: %s (%s) - ID: %v\n", row["name"], row["email"], row["id"])
	}
}

// Example14_QueryRowMap demonstrates DB.QueryRowMap - Single row as map
func Example14_QueryRowMap(ctx context.Context, db *typedb.DB, firstUser *User) {
	fmt.Println("\n--- Example 14: DB.QueryRowMap - Single Row as Map ---")
	row, err := db.QueryRowMap(ctx, "SELECT id, name, email FROM users WHERE id = $1", firstUser.ID)
	if err != nil {
		log.Fatalf("Failed to query row: %v", err)
	}
	if row != nil {
		fmt.Printf("  User: %s (%s)\n", row["name"], row["email"])
	}
}

// runRawQueryExamples demonstrates raw query operations returning maps.
func runRawQueryExamples(ctx context.Context, db *typedb.DB, firstUser *User) {
	Example13_QueryAll(ctx, db)
	Example14_QueryRowMap(ctx, db, firstUser)
}
