package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/TheBlackHowling/typedb"
	"github.com/TheBlackHowling/typedb/examples/seed"
	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	ctx := context.Background()

	// Get DSN from environment variable or use default
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://user:password@localhost/typedb_examples?sslmode=disable"
	}

	// Open database connection
	db, err := typedb.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fmt.Println("✓ Connected to PostgreSQL")

	// Clear and seed database with random data
	fmt.Println("\n--- Clearing and Seeding Database ---")
	if err := seed.ClearDatabase(ctx, db); err != nil {
		log.Fatalf("Failed to clear database: %v", err)
	}
	if err := seed.SeedDatabase(ctx, db, 10); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	// Run examples by category
	firstUser := runQueryExamples(ctx, db)
	runLoadExamples(ctx, db, firstUser)
	postID := runInsertExamples(ctx, db, firstUser)
	runUpdateExamples(ctx, db, firstUser)
	runTransactionExamples(ctx, db, firstUser)
	runRawQueryExamples(ctx, db, firstUser)
	runLoadCompositeExample(ctx, db, firstUser, postID)

	fmt.Println("\n✓ All examples completed successfully!")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
