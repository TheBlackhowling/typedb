package main

import (
	"context"
	"fmt"
	"log"

	"github.com/TheBlackHowling/typedb"
)

// Example4_Load demonstrates Load - Load user by primary key
func Example4_Load(ctx context.Context, db *typedb.DB, firstUser *User) {
	fmt.Println("\n--- Example 4: Load - Load User by Primary Key ---")
	userToLoad := &User{ID: firstUser.ID}
	if err := typedb.Load(ctx, db, userToLoad); err != nil {
		log.Fatalf("Failed to load user: %v", err)
	}
	fmt.Printf("  Loaded user: %s (%s)\n", userToLoad.Name, userToLoad.Email)
}

// Example5_LoadByField demonstrates LoadByField - Load user by email
func Example5_LoadByField(ctx context.Context, db *typedb.DB, firstUser *User) {
	fmt.Println("\n--- Example 5: LoadByField - Load User by Email ---")
	userByEmail := &User{Email: firstUser.Email}
	if err := typedb.LoadByField(ctx, db, userByEmail, "Email"); err != nil {
		log.Fatalf("Failed to load user by email: %v", err)
	}
	fmt.Printf("  Loaded user by email: %s (ID: %d)\n", userByEmail.Name, userByEmail.ID)
}

// Example11_LoadByField demonstrates LoadByField - Load profile by user_id
func Example11_LoadByField(ctx context.Context, db *typedb.DB, firstUser *User) {
	fmt.Println("\n--- Example 11: LoadByField - Load Profile by User ID ---")
	profile := &Profile{UserID: firstUser.ID}
	if err := typedb.LoadByField(ctx, db, profile, "UserID"); err != nil {
		log.Fatalf("Failed to load profile: %v", err)
	}
	fmt.Printf("  Profile: %s (Location: %s)\n", profile.Bio[:min(50, len(profile.Bio))], profile.Location)
}

// Example11b_LoadByComposite demonstrates LoadByComposite - Load user_post by composite key
func Example11b_LoadByComposite(ctx context.Context, db *typedb.DB, firstUser *User, postID int64) {
	fmt.Println("\n--- Example 11b: LoadByComposite - Load UserPost by Composite Key ---")
	// First, create a user_post relationship using raw INSERT (composite keys don't work with Insert)
	_, err := db.Exec(ctx, "INSERT INTO user_posts (user_id, post_id) VALUES (?, ?) ON DUPLICATE KEY UPDATE user_id=user_id", firstUser.ID, postID)
	if err != nil {
		log.Fatalf("Failed to insert user_post: %v", err)
	}
	// Now load it back using composite key
	loadedUserPost := &UserPost{
		UserID: firstUser.ID,
		PostID: int(postID),
	}
	if err := typedb.LoadByComposite(ctx, db, loadedUserPost, "userpost"); err != nil {
		log.Fatalf("Failed to load user_post by composite key: %v", err)
	}
	fmt.Printf("  Loaded UserPost: User %d favorited Post %d at %s\n", loadedUserPost.UserID, loadedUserPost.PostID, loadedUserPost.FavoritedAt)
}

// runLoadExamples demonstrates Load and LoadByField operations.
func runLoadExamples(ctx context.Context, db *typedb.DB, firstUser *User) {
	Example4_Load(ctx, db, firstUser)
	Example5_LoadByField(ctx, db, firstUser)
	Example11_LoadByField(ctx, db, firstUser)
}

// runLoadCompositeExample demonstrates LoadByComposite operation.
func runLoadCompositeExample(ctx context.Context, db *typedb.DB, firstUser *User, postID int64) {
	Example11b_LoadByComposite(ctx, db, firstUser, postID)
}
