package main

import (
	"context"
	"testing"
	"time"

	"github.com/TheBlackHowling/typedb"
	_ "github.com/sijms/go-ora/v2" // Oracle driver
)

func TestOracle_Update(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Get first user
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE ROWNUM <= 1 ORDER BY id")
	if err != nil || firstUser == nil {
		t.Fatal("Need at least one user in database")
	}

	originalName := firstUser.Name
	userID := firstUser.ID

	// Register cleanup to restore original name even on failure
	t.Cleanup(func() {
		if userID != 0 {
			restoreUser := &User{
				ID:   userID,
				Name: originalName,
			}
			typedb.Update(ctx, db, restoreUser)
		}
	})

	// Update user
	userToUpdate := &User{
		ID:   userID,
		Name: "Updated Test Name",
	}
	if err := typedb.Update(ctx, db, userToUpdate); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	updatedUser := &User{ID: userID}
	if err := typedb.Load(ctx, db, updatedUser); err != nil {
		t.Fatalf("Failed to load updated user: %v", err)
	}

	if updatedUser.Name != "Updated Test Name" {
		t.Errorf("Expected name 'Updated Test Name', got '%s'", updatedUser.Name)
	}
}

func TestOracle_Update_AutoTimestamp(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Get first user
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at, updated_at FROM users WHERE ROWNUM <= 1 ORDER BY id")
	if err != nil || firstUser == nil {
		t.Fatal("Need at least one user in database")
	}

	originalUpdatedAt := firstUser.UpdatedAt
	originalName := firstUser.Name

	// Register cleanup to restore original values
	t.Cleanup(func() {
		if firstUser.ID != 0 {
			restoreUser := &User{
				ID:   firstUser.ID,
				Name: originalName,
			}
			typedb.Update(ctx, db, restoreUser)
		}
	})

	// Update user - UpdatedAt should be auto-populated
	userToUpdate := &User{
		ID:   firstUser.ID,
		Name: "Updated Name for Timestamp Test",
		// UpdatedAt is not set - should be auto-populated with database timestamp
	}
	if err := typedb.Update(ctx, db, userToUpdate); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Wait a moment to ensure timestamp changes (database timestamp precision)
	time.Sleep(2 * time.Second)

	// Verify update and check updated_at was populated
	updatedUser := &User{ID: firstUser.ID}
	if err := typedb.Load(ctx, db, updatedUser); err != nil {
		t.Fatalf("Failed to load updated user: %v", err)
	}

	if updatedUser.Name != "Updated Name for Timestamp Test" {
		t.Errorf("Expected name 'Updated Name for Timestamp Test', got '%s'", updatedUser.Name)
	}

	// Verify updated_at was set (should be populated after update)
	if updatedUser.UpdatedAt == "" {
		t.Error("UpdatedAt should be populated after update")
	}
	// Verify UpdatedAt changed from the original value
	// If original was empty/NULL, it should now be set (different)
	// If original had a value, it should have changed
	if updatedUser.UpdatedAt == originalUpdatedAt {
		t.Errorf("UpdatedAt should have changed after update. Original: %q, New: %q", originalUpdatedAt, updatedUser.UpdatedAt)
	}
}

func TestOracle_Update_PartialUpdate(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Get first user
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at, updated_at FROM users WHERE ROWNUM <= 1 ORDER BY id")
	if err != nil || firstUser == nil {
		t.Fatal("Need at least one user in database")
	}

	originalName := firstUser.Name
	originalEmail := firstUser.Email

	// Register cleanup to restore original values
	t.Cleanup(func() {
		if firstUser.ID != 0 {
			restoreUser := &User{
				ID:    firstUser.ID,
				Name:  originalName,
				Email: originalEmail,
			}
			typedb.Update(ctx, db, restoreUser)
		}
	})

	// Load user to save original copy (required for partial update)
	userToUpdate := &User{ID: firstUser.ID}
	if err := typedb.Load(ctx, db, userToUpdate); err != nil {
		t.Fatalf("Failed to load user: %v", err)
	}

	originalLoadedEmail := userToUpdate.Email

	// Modify only name, keep email unchanged
	userToUpdate.Name = "Partial Update Test Name"
	// Email remains unchanged - should not be included in UPDATE

	if err := typedb.Update(ctx, db, userToUpdate); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Reload user to verify only name was updated
	updatedUser := &User{ID: firstUser.ID}
	if err := typedb.Load(ctx, db, updatedUser); err != nil {
		t.Fatalf("Failed to load updated user: %v", err)
	}

	// Verify name was updated
	if updatedUser.Name != "Partial Update Test Name" {
		t.Errorf("Expected name 'Partial Update Test Name', got '%s'", updatedUser.Name)
	}

	// Verify email was NOT changed (should remain the same)
	if updatedUser.Email != originalLoadedEmail {
		t.Errorf("Expected email to remain unchanged '%s', got '%s'", originalLoadedEmail, updatedUser.Email)
	}

	// Test 2: Update multiple fields
	userToUpdate2 := &User{ID: firstUser.ID}
	if err := typedb.Load(ctx, db, userToUpdate2); err != nil {
		t.Fatalf("Failed to load user for second test: %v", err)
	}

	userToUpdate2.Name = "Updated Name Again"
	userToUpdate2.Email = "updated.email@example.com"

	if err := typedb.Update(ctx, db, userToUpdate2); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Reload to verify both fields were updated
	updatedUser2 := &User{ID: firstUser.ID}
	if err := typedb.Load(ctx, db, updatedUser2); err != nil {
		t.Fatalf("Failed to load updated user: %v", err)
	}

	if updatedUser2.Name != "Updated Name Again" {
		t.Errorf("Expected name 'Updated Name Again', got '%s'", updatedUser2.Name)
	}
	if updatedUser2.Email != "updated.email@example.com" {
		t.Errorf("Expected email 'updated.email@example.com', got '%s'", updatedUser2.Email)
	}
}

func TestOracle_Update_NonPartialUpdate(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Use Post model which doesn't have partial update enabled
	// Get first user to create post if needed
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE ROWNUM <= 1 ORDER BY id")
	if err != nil || firstUser == nil {
		t.Fatal("Need at least one user in database for non-partial update test")
	}

	// Get or create a post for testing
	firstPost, err := typedb.QueryFirst[*Post](ctx, db, "SELECT id, user_id, title, content, tags, metadata, created_at, updated_at FROM posts WHERE ROWNUM <= 1 ORDER BY id")
	if err != nil || firstPost == nil {
		// Create a post if none exists
		newPost := &Post{
			UserID:  firstUser.ID,
			Title:   "Test Post for Non-Partial Update",
			Content: "Original content",
		}
		if err := typedb.Insert(ctx, db, newPost); err != nil {
			t.Fatalf("Failed to create test post: %v", err)
		}
		firstPost = newPost
	}

	originalPostTitle := firstPost.Title
	originalPostContent := firstPost.Content

	t.Cleanup(func() {
		if firstPost.ID != 0 {
			restorePost := &Post{
				ID:      firstPost.ID,
				UserID:  firstPost.UserID,
				Title:   originalPostTitle,
				Content: originalPostContent,
			}
			typedb.Update(ctx, db, restorePost)
		}
	})

	// Load post to get all current values
	postBeforeUpdate := &Post{ID: firstPost.ID}
	if err := typedb.Load(ctx, db, postBeforeUpdate); err != nil {
		t.Fatalf("Failed to load post: %v", err)
	}

	originalLoadedTitle := postBeforeUpdate.Title
	originalLoadedContent := postBeforeUpdate.Content

	// Update post with only Title set (Content not set = zero value)
	// Since Post doesn't have partial update enabled, ALL fields should be included in UPDATE
	postToUpdate := &Post{
		ID:     firstPost.ID,
		UserID: firstPost.UserID,
		Title:  "Updated Title Only",
		// Content is not set (zero value) - should be excluded from UPDATE
	}

	if err := typedb.Update(ctx, db, postToUpdate); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Reload post to verify update
	updatedPost := &Post{ID: firstPost.ID}
	if err := typedb.Load(ctx, db, updatedPost); err != nil {
		t.Fatalf("Failed to load updated post: %v", err)
	}

	// Verify title was updated
	if updatedPost.Title != "Updated Title Only" {
		t.Errorf("Expected title 'Updated Title Only', got '%s'", updatedPost.Title)
	}

	// Verify content was NOT updated (should remain unchanged since zero values are excluded)
	// This demonstrates that non-partial update still excludes zero values
	if updatedPost.Content != originalLoadedContent {
		t.Errorf("Expected content to remain unchanged '%s', got '%s'", originalLoadedContent, updatedPost.Content)
	}

	// Restore original content
	restorePost := &Post{
		ID:      firstPost.ID,
		UserID:  firstPost.UserID,
		Title:   originalLoadedTitle,
		Content: originalLoadedContent,
	}
	if err := typedb.Update(ctx, db, restorePost); err != nil {
		t.Fatalf("Failed to restore original post: %v", err)
	}
}
