package main

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/TheBlackHowling/typedb"
	_ "github.com/microsoft/go-mssqldb" // SQL Server driver
)

func getTestDSN() string {
	dsn := os.Getenv("MSSQL_DSN")
	if dsn == "" {
		dsn = "server=localhost;user id=sa;password=YourPassword123;database=typedb_examples"
	}
	return dsn
}

func TestMSSQL_QueryAll(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

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

func TestMSSQL_Load(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

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

func TestMSSQL_LoadByField(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

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

func TestMSSQL_LoadByComposite(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	userPost := &UserPost{UserID: 1, PostID: 1}
	if err := typedb.LoadByComposite(ctx, db, userPost, "userpost"); err != nil {
		t.Fatalf("LoadByComposite failed: %v", err)
	}

	if userPost.FavoritedAt == "" {
		t.Error("FavoritedAt should be loaded")
	}
}

func TestMSSQL_Transaction(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	err = db.WithTx(ctx, func(tx *typedb.Tx) error {
		users, err := typedb.QueryAll[*User](ctx, tx, "SELECT id, name, email, created_at FROM users WHERE id = @p1", 1)
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

func TestMSSQL_ComprehensiveTypes(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Test loading comprehensive type example
	typeExample := &TypeExample{ID: 1}
	if err := typedb.Load(ctx, db, typeExample); err != nil {
		t.Fatalf("Load type example failed: %v", err)
	}

	// Verify various types are deserialized
	if typeExample.IntegerCol == 0 {
		t.Error("IntegerCol should be loaded")
	}
	if typeExample.BigInt == 0 {
		t.Error("BigInt should be loaded")
	}
	if typeExample.VarcharCol == "" {
		t.Error("VarcharCol should be loaded")
	}
	if typeExample.NvarcharCol == "" {
		t.Error("NvarcharCol should be loaded")
	}
	if typeExample.DateCol == "" {
		t.Error("DateCol should be loaded")
	}
	if typeExample.Datetime2Col == "" {
		t.Error("Datetime2Col should be loaded")
	}
	if typeExample.XmlCol == "" {
		t.Error("XmlCol should be loaded")
	}

	// Test QueryAll with comprehensive types
	examples, err := typedb.QueryAll[*TypeExample](ctx, db, "SELECT id, tiny_int, small_int, integer_col, big_int, decimal_col, numeric_col, money_col, smallmoney_col, bit_col, float_col, real_col, char_col, varchar_col, varchar_max_col, nchar_col, nvarchar_col, nvarchar_max_col, text_col, ntext_col, binary_col, varbinary_col, varbinary_max_col, image_col, date_col, time_col, datetime_col, datetime2_col, datetimeoffset_col, smalldatetime_col, timestamp_col, uniqueidentifier_col, xml_col, hierarchyid_col, geography_col, geometry_col, sql_variant_col, created_at FROM type_examples")
	if err != nil {
		t.Fatalf("QueryAll type examples failed: %v", err)
	}

	if len(examples) == 0 {
		t.Fatal("Expected at least one type example")
	}
}

func TestMSSQL_ComprehensiveTypesRoundTrip(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Create test data with all fields populated
	testID := 9999
	insertSQL := `SET IDENTITY_INSERT type_examples ON;
	INSERT INTO type_examples (
		id, tiny_int, small_int, integer_col, big_int,
		decimal_col, numeric_col, money_col, smallmoney_col, bit_col,
		float_col, real_col,
		char_col, varchar_col, varchar_max_col,
		nchar_col, nvarchar_col, nvarchar_max_col,
		text_col, ntext_col,
		binary_col, varbinary_col, varbinary_max_col, image_col,
		date_col, time_col, datetime_col, datetime2_col, datetimeoffset_col, smalldatetime_col,
		uniqueidentifier_col,
		xml_col,
		hierarchyid_col,
		geography_col,
		geometry_col,
		sql_variant_col
	) VALUES (
		@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8, @p9, @p10,
		@p11, @p12, @p13, @p14, @p15, @p16, @p17, @p18, @p19, @p20,
		@p21, @p22, @p23, @p24, @p25, @p26, @p27, @p28, @p29, @p30,
		@p31, @p32, @p33, geography::STGeomFromText(@p34, 4326), geometry::STGeomFromText(@p35, 0), CAST(@p36 AS SQL_VARIANT)
	);
	SET IDENTITY_INSERT type_examples OFF;`

	// Insert test data
	_, err = db.Exec(ctx, insertSQL,
		testID,                      // id
		100,                         // tiny_int
		12345,                       // small_int
		987654321,                   // integer_col
		9223372036854775800,         // big_int
		"1234.56",                   // decimal_col
		"9876.54",                   // numeric_col
		"$1234.56",                  // money_col
		"$50.25",                    // smallmoney_col
		true,                        // bit_col
		3.14159,                     // float_col
		2.71828,                     // real_col
		"test_char  ",               // char_col (padded)
		"test_varchar",              // varchar_col
		"test varchar max content",  // varchar_max_col
		"test_nchar ",               // nchar_col (padded)
		"test_nvarchar",             // nvarchar_col
		"test nvarchar max content", // nvarchar_max_col
		"test text content",         // text_col
		"test ntext content",        // ntext_col
		[]byte{0xDE, 0xAD, 0xBE, 0xEF, 0xCA, 0xFE, 0xBA, 0xBE, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}, // binary_col
		[]byte{0xCA, 0xFE, 0xBA, 0xBE}, // varbinary_col
		[]byte{0x01, 0x02, 0x03, 0x04, 0x05}, // varbinary_max_col
		[]byte{0x10, 0x11, 0x12, 0x13, 0x14, 0x15}, // image_col
		"2024-12-25",                // date_col
		"15:30:45",                  // time_col
		"2024-12-25 15:30:45",       // datetime_col
		"2024-12-25 15:30:45.1234567", // datetime2_col
		"2024-12-25 15:30:45 +00:00", // datetimeoffset_col
		"2024-12-25 15:30:00",       // smalldatetime_col
		"550e8400-e29b-41d4-a716-446655440000", // uniqueidentifier_col
		"<root><test>roundtrip</test></root>", // xml_col
		"/1/",                       // hierarchyid_col
		"POINT(-122.4194 37.7749)",  // geography_col (San Francisco)
		"POINT(10 20)",              // geometry_col
		"SQL_VARIANT test",          // sql_variant_col
	)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Clean up after test
	defer func() {
		db.Exec(ctx, "DELETE FROM type_examples WHERE id = @p1", testID)
	}()

	// Query it back
	loaded := &TypeExample{ID: testID}
	if err := typedb.Load(ctx, db, loaded); err != nil {
		t.Fatalf("Failed to load inserted data: %v", err)
	}

	// Validate every field
	if loaded.TinyInt != 100 {
		t.Errorf("TinyInt: expected 100, got %d", loaded.TinyInt)
	}
	if loaded.SmallInt != 12345 {
		t.Errorf("SmallInt: expected 12345, got %d", loaded.SmallInt)
	}
	if loaded.IntegerCol != 987654321 {
		t.Errorf("IntegerCol: expected 987654321, got %d", loaded.IntegerCol)
	}
	if loaded.BigInt != 9223372036854775800 {
		t.Errorf("BigInt: expected 9223372036854775800, got %d", loaded.BigInt)
	}
	if loaded.DecimalCol == "" {
		t.Error("DecimalCol: should not be empty")
	}
	if loaded.NumericCol == "" {
		t.Error("NumericCol: should not be empty")
	}
	if loaded.MoneyCol == "" {
		t.Error("MoneyCol: should not be empty")
	}
	if loaded.SmallmoneyCol == "" {
		t.Error("SmallmoneyCol: should not be empty")
	}
	if !loaded.BitCol {
		t.Error("BitCol: expected true, got false")
	}
	if loaded.FloatCol == "" {
		t.Error("FloatCol: should not be empty")
	}
	if loaded.RealCol == "" {
		t.Error("RealCol: should not be empty")
	}
	if loaded.CharCol == "" {
		t.Error("CharCol: should not be empty")
	}
	if loaded.VarcharCol != "test_varchar" {
		t.Errorf("VarcharCol: expected 'test_varchar', got '%s'", loaded.VarcharCol)
	}
	if loaded.VarcharMaxCol == "" {
		t.Error("VarcharMaxCol: should not be empty")
	}
	if loaded.NcharCol == "" {
		t.Error("NcharCol: should not be empty")
	}
	if loaded.NvarcharCol != "test_nvarchar" {
		t.Errorf("NvarcharCol: expected 'test_nvarchar', got '%s'", loaded.NvarcharCol)
	}
	if loaded.NvarcharMaxCol == "" {
		t.Error("NvarcharMaxCol: should not be empty")
	}
	if loaded.TextCol == "" {
		t.Error("TextCol: should not be empty")
	}
	if loaded.NtextCol == "" {
		t.Error("NtextCol: should not be empty")
	}
	if loaded.BinaryCol == "" {
		t.Error("BinaryCol: should not be empty")
	}
	if loaded.VarbinaryCol == "" {
		t.Error("VarbinaryCol: should not be empty")
	}
	if loaded.VarbinaryMaxCol == "" {
		t.Error("VarbinaryMaxCol: should not be empty")
	}
	if loaded.ImageCol == "" {
		t.Error("ImageCol: should not be empty")
	}
	if loaded.DateCol == "" || !strings.Contains(loaded.DateCol, "2024-12-25") {
		t.Errorf("DateCol: expected to contain '2024-12-25', got '%s'", loaded.DateCol)
	}
	if loaded.TimeCol == "" {
		t.Error("TimeCol: should not be empty")
	}
	if loaded.DatetimeCol == "" {
		t.Error("DatetimeCol: should not be empty")
	}
	if loaded.Datetime2Col == "" {
		t.Error("Datetime2Col: should not be empty")
	}
	if loaded.DatetimeoffsetCol == "" {
		t.Error("DatetimeoffsetCol: should not be empty")
	}
	if loaded.SmalldatetimeCol == "" {
		t.Error("SmalldatetimeCol: should not be empty")
	}
	if loaded.TimestampCol == "" {
		t.Error("TimestampCol: should not be empty")
	}
	if loaded.UniqueidentifierCol == "" {
		t.Error("UniqueidentifierCol: should not be empty")
	}
	if loaded.XmlCol == "" {
		t.Error("XmlCol: should not be empty")
	}
	if loaded.HierarchyidCol == "" {
		t.Error("HierarchyidCol: should not be empty")
	}
	if loaded.GeographyCol == "" {
		t.Error("GeographyCol: should not be empty")
	}
	if loaded.GeometryCol == "" {
		t.Error("GeometryCol: should not be empty")
	}
	if loaded.SqlVariantCol == "" {
		t.Error("SqlVariantCol: should not be empty")
	}
	if loaded.CreatedAt == "" {
		t.Error("CreatedAt: should not be empty")
	}
}

// TestMSSQL_QueryFirst tests QueryFirst functionality
func TestMSSQL_QueryFirst(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	user, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = @p1", 1)
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

// TestMSSQL_QueryOne tests QueryOne functionality
func TestMSSQL_QueryOne(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	user, err := typedb.QueryOne[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = @p1", 1)
	if err != nil {
		t.Fatalf("QueryOne failed: %v", err)
	}

	if user.ID != 1 {
		t.Errorf("Expected user ID 1, got %d", user.ID)
	}
}

// TestMSSQL_Insert tests Insert by object functionality
func TestMSSQL_Insert(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

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

	// Clean up
	defer func() {
		db.Exec(ctx, "DELETE FROM users WHERE id = @p1", newUser.ID)
	}()

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

// TestMSSQL_InsertAndReturn tests InsertAndReturn functionality
func TestMSSQL_InsertAndReturn(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Get first user for foreign key
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT TOP 1 id, name, email, created_at FROM users ORDER BY id")
	if err != nil || firstUser == nil {
		t.Fatal("Need at least one user in database for foreign key")
	}

	// Insert post with OUTPUT clause (MSSQL equivalent of RETURNING)
	insertedPost, err := typedb.InsertAndReturn[*Post](ctx, db,
		"INSERT INTO posts (user_id, title, content, tags, metadata, created_at) OUTPUT INSERTED.id, INSERTED.user_id, INSERTED.title, INSERTED.content, INSERTED.tags, INSERTED.metadata, INSERTED.created_at VALUES (@p1, @p2, @p3, @p4, @p5, @p6)",
		firstUser.ID, "Test Post", "Test content", `["go","database"]`, `{"test":true}`, "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("InsertAndReturn failed: %v", err)
	}

	// Clean up
	defer func() {
		db.Exec(ctx, "DELETE FROM posts WHERE id = @p1", insertedPost.ID)
	}()

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

// TestMSSQL_InsertAndGetId tests InsertAndGetId functionality
func TestMSSQL_InsertAndGetId(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Get first user for foreign key
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT TOP 1 id, name, email, created_at FROM users ORDER BY id")
	if err != nil || firstUser == nil {
		t.Fatal("Need at least one user in database for foreign key")
	}

	// Insert post and get ID (MSSQL uses OUTPUT clause)
	postID, err := typedb.InsertAndGetId(ctx, db,
		"INSERT INTO posts (user_id, title, content, tags, metadata, created_at) OUTPUT INSERTED.id VALUES (@p1, @p2, @p3, @p4, @p5, @p6)",
		firstUser.ID, "Test Post ID", "Test content", `["go"]`, `{"test":true}`, "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("InsertAndGetId failed: %v", err)
	}

	// Clean up
	defer func() {
		db.Exec(ctx, "DELETE FROM posts WHERE id = @p1", postID)
	}()

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

// TestMSSQL_Update tests Update by object functionality
func TestMSSQL_Update(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Get first user
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT TOP 1 id, name, email, created_at FROM users ORDER BY id")
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

// TestMSSQL_QueryFirst_NoRows tests QueryFirst with no rows (should return nil, no error)
func TestMSSQL_QueryFirst_NoRows(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Query for non-existent user
	user, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = @p1", 99999)
	if err != nil {
		t.Fatalf("QueryFirst should not return error for no rows, got: %v", err)
	}

	if user != nil {
		t.Error("QueryFirst should return nil for no rows")
	}
}

// TestMSSQL_QueryOne_NoRows tests QueryOne with no rows (should return ErrNotFound)
func TestMSSQL_QueryOne_NoRows(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Query for non-existent user
	user, err := typedb.QueryOne[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = @p1", 99999)
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

// TestMSSQL_QueryOne_MultipleRows tests QueryOne with multiple rows (should return error)
func TestMSSQL_QueryOne_MultipleRows(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

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

// TestMSSQL_Negative_InvalidQuery tests error handling for invalid SQL
func TestMSSQL_Negative_InvalidQuery(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Invalid SQL query
	_, err = typedb.QueryAll[*User](ctx, db, "SELECT invalid_column FROM users")
	if err == nil {
		t.Fatal("QueryAll should return error for invalid SQL")
	}
}

// TestMSSQL_Negative_ConstraintViolation tests error handling for constraint violations
func TestMSSQL_Negative_ConstraintViolation(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Get existing user email
	existingUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT TOP 1 id, name, email, created_at FROM users ORDER BY id")
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
			db.Exec(ctx, "DELETE FROM users WHERE id = @p1", duplicateUser.ID)
		}
		t.Fatal("Insert should fail with unique constraint violation")
	}

	// Verify error indicates constraint violation
	if !strings.Contains(err.Error(), "unique") && !strings.Contains(err.Error(), "duplicate") && !strings.Contains(err.Error(), "UNIQUE") {
		t.Errorf("Expected constraint violation error, got: %v", err)
	}
}

// TestMSSQL_Update_AutoTimestamp tests auto-updated timestamp functionality
func TestMSSQL_Update_AutoTimestamp(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Get first user
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT TOP 1 id, name, email, created_at, updated_at FROM users ORDER BY id")
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

// TestMSSQL_Update_PartialUpdate tests partial update functionality
func TestMSSQL_Update_PartialUpdate(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Get first user
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT TOP 1 id, name, email, created_at, updated_at FROM users ORDER BY id")
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

// TestMSSQL_Update_NonPartialUpdate tests that Update without partial update enabled updates all fields
func TestMSSQL_Update_NonPartialUpdate(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Use Post model which doesn't have partial update enabled
	firstPost, err := typedb.QueryFirst[*Post](ctx, db, "SELECT TOP 1 id, user_id, title, content, tags, metadata, created_at, updated_at FROM posts ORDER BY id")
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
