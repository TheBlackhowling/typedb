package seed

import (
	"context"
	"fmt"
	"strings"

	"github.com/TheBlackHowling/typedb"
	"github.com/jaswdr/faker"
)

// SeedDatabase seeds the database with random data using faker
func SeedDatabase(ctx context.Context, db typedb.Executor, numUsers int) error {
	fake := faker.New()

	// Seed Users
	fmt.Printf("Seeding %d users...\n", numUsers)
	userIDs := make([]int64, 0, numUsers)

	for i := 0; i < numUsers; i++ {
		user := &User{
			Name:  fake.Person().Name(),
			Email: fake.Internet().Email(),
			// CreatedAt will use database default (CURRENT_TIMESTAMP)
		}

		// Use Insert to insert the user (ID will be set automatically)
		if err := typedb.Insert(ctx, db, user); err != nil {
			return fmt.Errorf("failed to insert user: %w", err)
		}
		userIDs = append(userIDs, int64(user.ID))
	}

	// Seed Profiles (one per user)
	fmt.Printf("Seeding %d profiles...\n", len(userIDs))
	for _, userID := range userIDs {
		profile := &Profile{
			UserID:    int(userID),
			Bio:       fake.Lorem().Paragraph(3),
			AvatarURL: fake.Internet().URL(),
			Location:  fake.Address().City() + ", " + fake.Address().State(),
			Website:   fake.Internet().URL(),
			// CreatedAt will use database default (CURRENT_TIMESTAMP)
		}

		if err := typedb.Insert(ctx, db, profile); err != nil {
			return fmt.Errorf("failed to insert profile: %w", err)
		}
	}

	// Seed Posts (2-5 posts per user)
	fmt.Printf("Seeding posts...\n")
	totalPosts := 0
	for _, userID := range userIDs {
		numPosts := fake.IntBetween(2, 5)
		for j := 0; j < numPosts; j++ {
			post := &Post{
				UserID:    int(userID),
				Title:     fake.Lorem().Sentence(5),
				Content:   fake.Lorem().Paragraph(5),
				Published: fake.Bool(),
				// CreatedAt will use database default (CURRENT_TIMESTAMP)
			}

			if err := typedb.Insert(ctx, db, post); err != nil {
				return fmt.Errorf("failed to insert post: %w", err)
			}
			totalPosts++
		}
	}

	fmt.Printf("✓ Seeded %d users, %d profiles, and %d posts\n", numUsers, len(userIDs), totalPosts)
	return nil
}

// ClearDatabase clears all seeded data
func ClearDatabase(ctx context.Context, db typedb.Executor) error {
	// Delete in reverse order of foreign key dependencies
	// Use TRUNCATE TABLE for better performance, but fall back to DELETE if tables don't exist
	// Note: MSSQL requires TRUNCATE to be on its own line, and it doesn't work with foreign keys
	// So we'll use DELETE for all databases for consistency

	// Try to delete posts (ignore error if table doesn't exist)
	if _, err := db.Exec(ctx, "DELETE FROM posts"); err != nil {
		// Check if error is about table not existing - if so, that's okay
		errStr := strings.ToLower(err.Error())
		if !strings.Contains(errStr, "invalid object name") && !strings.Contains(errStr, "does not exist") && !strings.Contains(errStr, "no such table") {
			return fmt.Errorf("failed to delete posts: %w", err)
		}
	}
	if _, err := db.Exec(ctx, "DELETE FROM profiles"); err != nil {
		errStr := strings.ToLower(err.Error())
		if !strings.Contains(errStr, "invalid object name") && !strings.Contains(errStr, "does not exist") && !strings.Contains(errStr, "no such table") {
			return fmt.Errorf("failed to delete profiles: %w", err)
		}
	}
	if _, err := db.Exec(ctx, "DELETE FROM users"); err != nil {
		errStr := strings.ToLower(err.Error())
		if !strings.Contains(errStr, "invalid object name") && !strings.Contains(errStr, "does not exist") && !strings.Contains(errStr, "no such table") {
			return fmt.Errorf("failed to delete users: %w", err)
		}
	}
	fmt.Println("✓ Database cleared")
	return nil
}
