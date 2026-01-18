package main

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/TheBlackHowling/typedb"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

func getTestDSN() string {
	dsn := os.Getenv("SQLITE_DSN")
	if dsn == "" {
		dsn = "typedb_examples_test.db"
	}
	return dsn
}

func setupTestDB(t *testing.T) *typedb.DB {
	dsn := getTestDSN()

	// Remove existing test database
	os.Remove(dsn)

	// Open raw database connection for migrations
	sqlDB, err := sql.Open("sqlite3", dsn)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Run migrations
	driver, err := sqlite3.WithInstance(sqlDB, &sqlite3.Config{})
	if err != nil {
		sqlDB.Close()
		t.Fatalf("Failed to create migration driver: %v", err)
	}

	// Get migrations directory path - resolve relative to test file location
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)
	migrationsPath := filepath.Join(testDir, "migrations")
	
	// Convert to absolute path for file:// URL
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		sqlDB.Close()
		t.Fatalf("Failed to resolve migrations path: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+filepath.ToSlash(absPath),
		"sqlite3", driver)
	if err != nil {
		sqlDB.Close()
		t.Fatalf("Failed to create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		sqlDB.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Close the raw connection - migrations are done
	sqlDB.Close()

	// Now open with typedb (will use the same database file)
	db, err := typedb.OpenWithoutValidation("sqlite3", dsn)
	if err != nil {
		t.Fatalf("Failed to open typedb connection: %v", err)
	}

	return db
}

func TestSQLite_QueryAll(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer os.Remove(getTestDSN())

	ctx := context.Background()
	users, err := typedb.QueryAll[*User](ctx, db, "SELECT id, name, email, created_at FROM users ORDER BY id")
	if err != nil {
		t.Fatalf("QueryAll failed: %v", err)
	}

	if len(users) == 0 {
		t.Fatal("Expected at least one user")
	}

	if users[0].ID == 0 {
		t.Error("User ID should not be zero")
	}
	if users[0].Name == "" {
		t.Error("User name should not be empty")
	}
}

func TestSQLite_QueryFirst(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer os.Remove(getTestDSN())

	ctx := context.Background()
	user, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = ?", 1)
	if err != nil {
		t.Fatalf("QueryFirst failed: %v", err)
	}

	if user == nil {
		t.Fatal("Expected a user, got nil")
	}

	if user.ID != 1 {
		t.Errorf("Expected user ID 1, got %d", user.ID)
	}
}

func TestSQLite_QueryOne(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer os.Remove(getTestDSN())

	ctx := context.Background()
	user, err := typedb.QueryOne[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = ?", 1)
	if err != nil {
		t.Fatalf("QueryOne failed: %v", err)
	}

	if user.ID != 1 {
		t.Errorf("Expected user ID 1, got %d", user.ID)
	}
}

func TestSQLite_Load(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer os.Remove(getTestDSN())

	ctx := context.Background()
	user := &User{ID: 1}
	if err := typedb.Load(ctx, db, user); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if user.Name == "" {
		t.Error("User name should be loaded")
	}
	if user.Email == "" {
		t.Error("User email should be loaded")
	}
}

func TestSQLite_LoadByField(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer os.Remove(getTestDSN())

	ctx := context.Background()
	user := &User{Email: "alice@example.com"}
	if err := typedb.LoadByField(ctx, db, user, "Email"); err != nil {
		t.Fatalf("LoadByField failed: %v", err)
	}

	if user.ID == 0 {
		t.Error("User ID should be loaded")
	}
	if user.Name == "" {
		t.Error("User name should be loaded")
	}
}

func TestSQLite_LoadByComposite(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer os.Remove(getTestDSN())

	ctx := context.Background()
	userPost := &UserPost{UserID: 1, PostID: 1}
	if err := typedb.LoadByComposite(ctx, db, userPost, "userpost"); err != nil {
		t.Fatalf("LoadByComposite failed: %v", err)
	}

	if userPost.FavoritedAt == "" {
		t.Error("FavoritedAt should be loaded")
	}
}

func TestSQLite_Transaction(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer os.Remove(getTestDSN())

	ctx := context.Background()
	err := db.WithTx(ctx, func(tx *typedb.Tx) error {
		users, err := typedb.QueryAll[*User](ctx, tx, "SELECT id, name, email, created_at FROM users WHERE id = ?", 1)
		if err != nil {
			return err
		}
		if len(users) == 0 {
			t.Error("Expected at least one user in transaction")
		}
		return nil
	}, nil)

	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}
}

func TestSQLite_ComprehensiveTypes(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer os.Remove(getTestDSN())

	ctx := context.Background()

	// Test loading comprehensive type example
	typeExample := &TypeExample{ID: 1}
	if err := typedb.Load(ctx, db, typeExample); err != nil {
		t.Fatalf("Load type example failed: %v", err)
	}

	// Verify various types are deserialized
	if typeExample.IntegerCol == 0 {
		t.Error("IntegerCol should be loaded")
	}
	if typeExample.VarcharCol == "" {
		t.Error("VarcharCol should be loaded")
	}
	if typeExample.TextCol == "" {
		t.Error("TextCol should be loaded")
	}
	if typeExample.DateCol == "" {
		t.Error("DateCol should be loaded")
	}
	if typeExample.JsonCol == "" {
		t.Error("JsonCol should be loaded")
	}

	// Test QueryAll with comprehensive types
	examples, err := typedb.QueryAll[*TypeExample](ctx, db, "SELECT id, integer_col, real_col, numeric_col, text_col, varchar_col, char_col, clob_col, blob_col, date_col, datetime_col, timestamp_col, time_col, boolean_col, json_col, created_at FROM type_examples")
	if err != nil {
		t.Fatalf("QueryAll type examples failed: %v", err)
	}

	if len(examples) == 0 {
		t.Fatal("Expected at least one type example")
	}
}

func TestSQLite_ComprehensiveTypesRoundTrip(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer os.Remove(getTestDSN())

	ctx := context.Background()

	// Migrations already create the type_examples table, so we can proceed directly

	// Create test data with all fields populated
	testID := 9999
	insertSQL := `INSERT INTO type_examples (
		id, integer_col, real_col, numeric_col, text_col, varchar_col, char_col,
		clob_col, blob_col, date_col, datetime_col, timestamp_col, time_col,
		boolean_col, json_col
	) VALUES (
		?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
	)`

	// Insert test data
	_, err := db.Exec(ctx, insertSQL,
		testID,                      // id
		987654321,                   // integer_col
		3.14159,                     // real_col
		"1234.56",                   // numeric_col
		"test text content",         // text_col
		"test_varchar",              // varchar_col
		"test_char ",                // char_col (padded)
		"test clob content",         // clob_col
		[]byte{0xDE, 0xAD, 0xBE, 0xEF}, // blob_col
		"2024-12-25",                // date_col
		"2024-12-25 15:30:45",       // datetime_col
		"2024-12-25 15:30:45",       // timestamp_col
		"15:30:45",                  // time_col
		true,                        // boolean_col
		`{"test": "json_value", "number": 42}`, // json_col
	)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Clean up after test
	defer func() {
		db.Exec(ctx, "DELETE FROM type_examples WHERE id = ?", testID)
	}()

	// Query it back
	loaded := &TypeExample{ID: testID}
	if err := typedb.Load(ctx, db, loaded); err != nil {
		t.Fatalf("Failed to load inserted data: %v", err)
	}

	// Validate every field
	if loaded.IntegerCol != 987654321 {
		t.Errorf("IntegerCol: expected 987654321, got %d", loaded.IntegerCol)
	}
	if loaded.RealCol == "" {
		t.Error("RealCol: should not be empty")
	}
	if loaded.NumericCol == "" {
		t.Error("NumericCol: should not be empty")
	}
	if loaded.TextCol != "test text content" {
		t.Errorf("TextCol: expected 'test text content', got '%s'", loaded.TextCol)
	}
	if loaded.VarcharCol != "test_varchar" {
		t.Errorf("VarcharCol: expected 'test_varchar', got '%s'", loaded.VarcharCol)
	}
	if loaded.CharCol == "" {
		t.Error("CharCol: should not be empty")
	}
	if loaded.ClobCol != "test clob content" {
		t.Errorf("ClobCol: expected 'test clob content', got '%s'", loaded.ClobCol)
	}
	if loaded.BlobCol == "" {
		t.Error("BlobCol: should not be empty")
	}
	if loaded.DateCol == "" || !strings.Contains(loaded.DateCol, "2024-12-25") {
		t.Errorf("DateCol: expected to contain '2024-12-25', got '%s'", loaded.DateCol)
	}
	if loaded.DatetimeCol == "" {
		t.Error("DatetimeCol: should not be empty")
	}
	if loaded.TimestampCol == "" {
		t.Error("TimestampCol: should not be empty")
	}
	if loaded.TimeCol == "" {
		t.Error("TimeCol: should not be empty")
	}
	if !loaded.BooleanCol {
		t.Error("BooleanCol: expected true, got false")
	}
	if loaded.JsonCol == "" {
		t.Error("JsonCol: should not be empty")
	}
	if loaded.CreatedAt == "" {
		t.Error("CreatedAt: should not be empty")
	}
}

// TestSQLite_Insert tests Insert by object functionality
func TestSQLite_Insert(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer os.Remove(getTestDSN())

	ctx := context.Background()

	// Insert new user
	newUser := &User{
		Name:  "Test Insert User",
		Email: "testinsert@example.com",
	}
	if err := typedb.Insert(ctx, db, newUser); err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Verify ID was set
	if newUser.ID == 0 {
		t.Error("User ID should be set after insert")
	}

	// Verify user was inserted
	loaded := &User{ID: newUser.ID}
	if err := typedb.Load(ctx, db, loaded); err != nil {
		t.Fatalf("Failed to load inserted user: %v", err)
	}

	if loaded.Name != "Test Insert User" {
		t.Errorf("Expected name 'Test Insert User', got '%s'", loaded.Name)
	}
	if loaded.Email != "testinsert@example.com" {
		t.Errorf("Expected email 'testinsert@example.com', got '%s'", loaded.Email)
	}
}

// TestSQLite_InsertAndReturn tests InsertAndReturn functionality
func TestSQLite_InsertAndReturn(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer os.Remove(getTestDSN())

	ctx := context.Background()

	// Get first user for foreign key
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users ORDER BY id LIMIT 1")
	if err != nil || firstUser == nil {
		t.Fatal("Need at least one user in database for foreign key")
	}

	// Insert post with RETURNING clause (SQLite supports RETURNING)
	insertedPost, err := typedb.InsertAndReturn[*Post](ctx, db,
		"INSERT INTO posts (user_id, title, content, tags, metadata, created_at) VALUES (?, ?, ?, ?, ?, ?) RETURNING id, user_id, title, content, tags, metadata, created_at",
		firstUser.ID, "Test Post", "Test content", `["go","database"]`, `{"test":true}`, "2024-01-01 00:00:00")
	if err != nil {
		t.Fatalf("InsertAndReturn failed: %v", err)
	}

	// Verify returned post
	if insertedPost.ID == 0 {
		t.Error("Post ID should be set")
	}
	if insertedPost.Title != "Test Post" {
		t.Errorf("Expected title 'Test Post', got '%s'", insertedPost.Title)
	}
	if insertedPost.UserID != firstUser.ID {
		t.Errorf("Expected UserID %d, got %d", firstUser.ID, insertedPost.UserID)
	}
}

// TestSQLite_InsertAndGetId tests InsertAndGetId functionality
func TestSQLite_InsertAndGetId(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer os.Remove(getTestDSN())

	ctx := context.Background()

	// Get first user for foreign key
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users ORDER BY id LIMIT 1")
	if err != nil || firstUser == nil {
		t.Fatal("Need at least one user in database for foreign key")
	}

	// Insert post and get ID (SQLite supports RETURNING)
	postID, err := typedb.InsertAndGetId(ctx, db,
		"INSERT INTO posts (user_id, title, content, tags, metadata, created_at) VALUES (?, ?, ?, ?, ?, ?) RETURNING id",
		firstUser.ID, "Test Post ID", "Test content", `["go"]`, `{"test":true}`, "2024-01-01 00:00:00")
	if err != nil {
		t.Fatalf("InsertAndGetId failed: %v", err)
	}

	// Verify ID was returned
	if postID == 0 {
		t.Error("Post ID should not be zero")
	}

	// Verify post exists
	loaded := &Post{ID: int(postID)}
	if err := typedb.Load(ctx, db, loaded); err != nil {
		t.Fatalf("Failed to load inserted post: %v", err)
	}

	if loaded.Title != "Test Post ID" {
		t.Errorf("Expected title 'Test Post ID', got '%s'", loaded.Title)
	}
}

// TestSQLite_Update tests Update by object functionality
func TestSQLite_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer os.Remove(getTestDSN())

	ctx := context.Background()

	// Get first user
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users ORDER BY id LIMIT 1")
	if err != nil || firstUser == nil {
		t.Fatal("Need at least one user in database")
	}

	originalName := firstUser.Name

	// Update user
	userToUpdate := &User{
		ID:   firstUser.ID,
		Name: "Updated Test Name",
	}
	if err := typedb.Update(ctx, db, userToUpdate); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	updatedUser := &User{ID: firstUser.ID}
	if err := typedb.Load(ctx, db, updatedUser); err != nil {
		t.Fatalf("Failed to load updated user: %v", err)
	}

	if updatedUser.Name != "Updated Test Name" {
		t.Errorf("Expected name 'Updated Test Name', got '%s'", updatedUser.Name)
	}

	// Restore original name
	restoreUser := &User{
		ID:   firstUser.ID,
		Name: originalName,
	}
	if err := typedb.Update(ctx, db, restoreUser); err != nil {
		t.Fatalf("Failed to restore original name: %v", err)
	}
}

// TestSQLite_QueryFirst_NoRows tests QueryFirst with no rows (should return nil, no error)
func TestSQLite_QueryFirst_NoRows(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer os.Remove(getTestDSN())

	ctx := context.Background()

	// Query for non-existent user
	user, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = ?", 99999)
	if err != nil {
		t.Fatalf("QueryFirst should not return error for no rows, got: %v", err)
	}

	if user != nil {
		t.Error("QueryFirst should return nil for no rows")
	}
}

// TestSQLite_QueryOne_NoRows tests QueryOne with no rows (should return ErrNotFound)
func TestSQLite_QueryOne_NoRows(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer os.Remove(getTestDSN())

	ctx := context.Background()

	// Query for non-existent user
	user, err := typedb.QueryOne[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = ?", 99999)
	if err == nil {
		t.Fatal("QueryOne should return error for no rows")
	}

	if err != typedb.ErrNotFound {
		t.Errorf("QueryOne should return ErrNotFound for no rows, got: %v", err)
	}

	if user != nil {
		t.Error("QueryOne should return nil when error occurs")
	}
}

// TestSQLite_QueryOne_MultipleRows tests QueryOne with multiple rows (should return error)
func TestSQLite_QueryOne_MultipleRows(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer os.Remove(getTestDSN())

	ctx := context.Background()

	// Query that returns multiple rows (no WHERE clause)
	user, err := typedb.QueryOne[*User](ctx, db, "SELECT id, name, email, created_at FROM users")
	if err == nil {
		t.Fatal("QueryOne should return error for multiple rows")
	}

	if !strings.Contains(err.Error(), "multiple rows") {
		t.Errorf("QueryOne should return error about multiple rows, got: %v", err)
	}

	if user != nil {
		t.Error("QueryOne should return nil when error occurs")
	}
}

// TestSQLite_Negative_InvalidQuery tests error handling for invalid SQL
func TestSQLite_Negative_InvalidQuery(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer os.Remove(getTestDSN())

	ctx := context.Background()

	// Invalid SQL query
	_, err := typedb.QueryAll[*User](ctx, db, "SELECT invalid_column FROM users")
	if err == nil {
		t.Fatal("QueryAll should return error for invalid SQL")
	}
}

// TestSQLite_Negative_ConstraintViolation tests error handling for constraint violations
func TestSQLite_Negative_ConstraintViolation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer os.Remove(getTestDSN())

	ctx := context.Background()

	// Get existing user email
	existingUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users ORDER BY id LIMIT 1")
	if err != nil || existingUser == nil {
		t.Fatal("Need at least one user in database")
	}

	// Try to insert user with duplicate email (unique constraint violation)
	duplicateUser := &User{
		Name:  "Duplicate Email User",
		Email: existingUser.Email, // Duplicate email
	}
	err = typedb.Insert(ctx, db, duplicateUser)
	if err == nil {
		// Clean up if insert somehow succeeded
		if duplicateUser.ID != 0 {
			db.Exec(ctx, "DELETE FROM users WHERE id = ?", duplicateUser.ID)
		}
		t.Fatal("Insert should fail with unique constraint violation")
	}

	// Verify error indicates constraint violation
	if !strings.Contains(err.Error(), "unique") && !strings.Contains(err.Error(), "duplicate") && !strings.Contains(err.Error(), "UNIQUE") {
		t.Errorf("Expected constraint violation error, got: %v", err)
	}
}

// TestSQLite_Update_AutoTimestamp tests auto-updated timestamp functionality
func TestSQLite_Update_AutoTimestamp(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer os.Remove(getTestDSN())

	ctx := context.Background()

	// Get first user
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at, updated_at FROM users ORDER BY id LIMIT 1")
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

// TestSQLite_Update_PartialUpdate tests partial update functionality
func TestSQLite_Update_PartialUpdate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer os.Remove(getTestDSN())

	ctx := context.Background()

	// Get first user
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at, updated_at FROM users ORDER BY id LIMIT 1")
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

// TestSQLite_Update_NonPartialUpdate tests that Update without partial update enabled updates all fields
func TestSQLite_Update_NonPartialUpdate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer os.Remove(getTestDSN())

	ctx := context.Background()

	// Use Post model which doesn't have partial update enabled
	firstPost, err := typedb.QueryFirst[*Post](ctx, db, "SELECT id, user_id, title, content, tags, metadata, created_at, updated_at FROM posts ORDER BY id LIMIT 1")
	if err != nil || firstPost == nil {
		t.Skip("Need at least one post in database for non-partial update test")
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
		// Content is not set - should be updated to empty string when partial update is disabled
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

	// Verify content was also updated (to empty string, since it wasn't set)
	// This demonstrates that non-partial update includes ALL fields
	if updatedPost.Content != "" {
		t.Errorf("Expected content to be empty (zero value) when not set in non-partial update, got '%s'", updatedPost.Content)
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
