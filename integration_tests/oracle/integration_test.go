package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/TheBlackHowling/typedb"
	_ "github.com/sijms/go-ora/v2" // Oracle driver
)

func getTestDSN() string {
	dsn := os.Getenv("ORACLE_DSN")
	if dsn == "" {
		dsn = "oracle://user:password@localhost:1521/XE"
	}
	return dsn
}

func TestOracle_QueryAll(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
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

func TestOracle_Load(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
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

func TestOracle_LoadByField(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
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

func TestOracle_LoadByComposite(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
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

func TestOracle_Transaction(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	err = db.WithTx(ctx, func(tx *typedb.Tx) error {
		users, err := typedb.QueryAll[*User](ctx, tx, "SELECT id, name, email, created_at FROM users WHERE id = :1", 1)
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

func TestOracle_ComprehensiveTypes(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
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
	if typeExample.Varchar2Col == "" {
		t.Error("Varchar2Col should be loaded")
	}
	if typeExample.Nvarchar2Col == "" {
		t.Error("Nvarchar2Col should be loaded")
	}
	if typeExample.ClobCol == "" {
		t.Error("ClobCol should be loaded")
	}
	if typeExample.DateCol == "" {
		t.Error("DateCol should be loaded")
	}
	if typeExample.TimestampCol == "" {
		t.Error("TimestampCol should be loaded")
	}
	if typeExample.XmltypeCol == "" {
		t.Error("XmltypeCol should be loaded")
	}

	// Test QueryAll with comprehensive types
	examples, err := typedb.QueryAll[*TypeExample](ctx, db, "SELECT id, number_col, number_precision_col, number_scale_col, float_col, float_precision_col, binary_float_col, binary_double_col, char_col, varchar2_col, varchar_col, nchar_col, nvarchar2_col, clob_col, nclob_col, long_col, raw_col, blob_col, bfile_col, date_col, timestamp_col, timestamp_precision_col, timestamp_tz_col, timestamp_ltz_col, interval_year_col, interval_day_col, rowid_col, urowid_col, xmltype_col, created_at FROM type_examples")
	if err != nil {
		t.Fatalf("QueryAll type examples failed: %v", err)
	}

	if len(examples) == 0 {
		t.Fatal("Expected at least one type example")
	}
}

func TestOracle_ComprehensiveTypesRoundTrip(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Create test data with all fields populated
	testID := 9999
	insertSQL := `INSERT INTO type_examples (
		id, number_col, number_precision_col, number_scale_col,
		float_col, float_precision_col, binary_float_col, binary_double_col,
		char_col, varchar2_col, varchar_col, nchar_col, nvarchar2_col,
		clob_col, nclob_col, long_col,
		raw_col, blob_col,
		date_col, timestamp_col, timestamp_precision_col,
		timestamp_tz_col, timestamp_ltz_col,
		interval_year_col, interval_day_col,
		xmltype_col
	) VALUES (
		:1, :2, :3, :4, :5, :6, :7, :8, :9, :10, :11, :12, :13,
		:14, :15, :16, :17, :18, TO_DATE(:19, 'YYYY-MM-DD'), TO_TIMESTAMP(:20, 'YYYY-MM-DD HH24:MI:SS'), TO_TIMESTAMP(:21, 'YYYY-MM-DD HH24:MI:SS.FF6'),
		TO_TIMESTAMP_TZ(:22, 'YYYY-MM-DD HH24:MI:SS TZH:TZM'), TO_TIMESTAMP(:23, 'YYYY-MM-DD HH24:MI:SS'),
		INTERVAL '2-0' YEAR TO MONTH, INTERVAL '3 00:00:00.000000' DAY TO SECOND,
		XMLTYPE(:24)
	)`

	// Insert test data
	_, err = db.Exec(ctx, insertSQL,
		testID,                      // id
		"1234.56",                   // number_col
		"9876.54",                   // number_precision_col
		"1111.22",                   // number_scale_col
		3.14159,                     // float_col
		123.4567,                    // float_precision_col
		2.71828,                     // binary_float_col
		1.41421,                     // binary_double_col
		"test_char ",                // char_col (padded)
		"test_varchar2",             // varchar2_col
		"test_varchar",              // varchar_col
		"test_nchar",                // nchar_col (NCHAR(10), will be padded)
		"test_nvarchar2",            // nvarchar2_col
		"test clob content",         // clob_col
		"test nclob content",        // nclob_col
		"test long content",         // long_col
		[]byte{0xDE, 0xAD, 0xBE, 0xEF}, // raw_col
		[]byte{0x01, 0x02, 0x03, 0x04}, // blob_col
		"2024-12-25",                // date_col (used in TO_DATE)
		"2024-12-25 15:30:45",       // timestamp_col (used in TO_TIMESTAMP)
		"2024-12-25 15:30:45.123456", // timestamp_precision_col (used in TO_TIMESTAMP)
		"2024-12-25 15:30:45 +00:00", // timestamp_tz_col (used in TO_TIMESTAMP_TZ)
		"2024-12-25 15:30:45",       // timestamp_ltz_col (used in TO_TIMESTAMP)
		// interval_year_col and interval_day_col are hardcoded in SQL (can't parameterize INTERVAL literals)
		"<root><test>roundtrip</test></root>", // xmltype_col (used in XMLTYPE)
	)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Clean up after test
	defer func() {
		db.Exec(ctx, "DELETE FROM type_examples WHERE id = :1", testID)
	}()

	// Query it back
	loaded := &TypeExample{ID: testID}
	if err := typedb.Load(ctx, db, loaded); err != nil {
		t.Fatalf("Failed to load inserted data: %v", err)
	}

	// Validate every field
	if loaded.NumberCol == "" {
		t.Error("NumberCol: should not be empty")
	}
	if loaded.NumberPrecisionCol == "" {
		t.Error("NumberPrecisionCol: should not be empty")
	}
	if loaded.NumberScaleCol == "" {
		t.Error("NumberScaleCol: should not be empty")
	}
	if loaded.FloatCol == "" {
		t.Error("FloatCol: should not be empty")
	}
	if loaded.FloatPrecisionCol == "" {
		t.Error("FloatPrecisionCol: should not be empty")
	}
	if loaded.BinaryFloatCol == "" {
		t.Error("BinaryFloatCol: should not be empty")
	}
	if loaded.BinaryDoubleCol == "" {
		t.Error("BinaryDoubleCol: should not be empty")
	}
	if loaded.CharCol == "" {
		t.Error("CharCol: should not be empty")
	}
	if loaded.Varchar2Col != "test_varchar2" {
		t.Errorf("Varchar2Col: expected 'test_varchar2', got '%s'", loaded.Varchar2Col)
	}
	if loaded.VarcharCol != "test_varchar" {
		t.Errorf("VarcharCol: expected 'test_varchar', got '%s'", loaded.VarcharCol)
	}
	if loaded.NcharCol == "" {
		t.Error("NcharCol: should not be empty")
	}
	if loaded.Nvarchar2Col != "test_nvarchar2" {
		t.Errorf("Nvarchar2Col: expected 'test_nvarchar2', got '%s'", loaded.Nvarchar2Col)
	}
	if loaded.ClobCol != "test clob content" {
		t.Errorf("ClobCol: expected 'test clob content', got '%s'", loaded.ClobCol)
	}
	if loaded.NclobCol != "test nclob content" {
		t.Errorf("NclobCol: expected 'test nclob content', got '%s'", loaded.NclobCol)
	}
	if loaded.LongCol == "" {
		t.Error("LongCol: should not be empty")
	}
	if loaded.RawCol == "" {
		t.Error("RawCol: should not be empty")
	}
	// LONG RAW is tested separately in TestOracle_LongRawType due to Oracle's limitation
	if loaded.BlobCol == "" {
		t.Error("BlobCol: should not be empty")
	}
	if loaded.DateCol == "" || !strings.Contains(loaded.DateCol, "2024-12-25") {
		t.Errorf("DateCol: expected to contain '2024-12-25', got '%s'", loaded.DateCol)
	}
	if loaded.TimestampCol == "" {
		t.Error("TimestampCol: should not be empty")
	}
	if loaded.TimestampPrecisionCol == "" {
		t.Error("TimestampPrecisionCol: should not be empty")
	}
	if loaded.TimestampTzCol == "" {
		t.Error("TimestampTzCol: should not be empty")
	}
	if loaded.TimestampLtzCol == "" {
		t.Error("TimestampLtzCol: should not be empty")
	}
	if loaded.IntervalYearCol == "" {
		t.Error("IntervalYearCol: should not be empty")
	}
	if loaded.IntervalDayCol == "" {
		t.Error("IntervalDayCol: should not be empty")
	}
	if loaded.XmltypeCol == "" {
		t.Error("XmltypeCol: should not be empty")
	}
	if loaded.CreatedAt == "" {
		t.Error("CreatedAt: should not be empty")
	}
}

func TestOracle_LongRawType(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Test loading LONG RAW example from migration data
	longRawExample := &LongRawExample{ID: 1}
	if err := typedb.Load(ctx, db, longRawExample); err != nil {
		t.Fatalf("Load long raw example failed: %v", err)
	}

	// Verify LONG RAW is deserialized
	if longRawExample.LongRawCol == "" {
		t.Error("LongRawCol should be loaded")
	}

	// Test round-trip: insert and query back, validating exact match
	testID := 9999
	testHexValue := "DEADBEEFCAFEBABE" // Longer test value for better validation
	
	// Convert hex string to bytes for validation
	expectedBytes, err := hex.DecodeString(testHexValue)
	if err != nil {
		t.Fatalf("Failed to decode test hex value: %v", err)
	}
	
	insertSQL := `INSERT INTO long_raw_examples (id, long_raw_col) VALUES (:1, HEXTORAW(:2))`
	_, err = db.Exec(ctx, insertSQL, testID, testHexValue)
	if err != nil {
		t.Fatalf("Failed to insert LONG RAW test data: %v", err)
	}

	defer func() {
		db.Exec(ctx, "DELETE FROM long_raw_examples WHERE id = :1", testID)
	}()

	// Query it back
	loaded := &LongRawExample{ID: testID}
	if err := typedb.Load(ctx, db, loaded); err != nil {
		t.Fatalf("Failed to load inserted LONG RAW data: %v", err)
	}

	if loaded.LongRawCol == "" {
		t.Error("LongRawCol: should not be empty after round-trip")
	}

	// Oracle returns LONG RAW as binary bytes converted to string
	// Convert the loaded string back to bytes and then to hex for comparison
	loadedBytes := []byte(loaded.LongRawCol)
	loadedHex := hex.EncodeToString(loadedBytes)
	
	// Validate that the retrieved bytes match what we inserted
	if len(loadedBytes) != len(expectedBytes) {
		t.Errorf("LongRawCol byte length mismatch: expected %d bytes, got %d bytes", len(expectedBytes), len(loadedBytes))
	}
	
	// Compare hex representations (case-insensitive)
	loadedHexUpper := strings.ToUpper(loadedHex)
	testHexUpper := strings.ToUpper(testHexValue)
	if loadedHexUpper != testHexUpper {
		t.Errorf("LongRawCol round-trip validation failed: expected hex '%s', got hex '%s' (raw: %q)", testHexValue, loadedHex, loaded.LongRawCol)
	}
	
	// Also verify byte-by-byte match
	for i := range expectedBytes {
		if i >= len(loadedBytes) {
			t.Errorf("LongRawCol: byte array shorter than expected at index %d", i)
			break
		}
		if loadedBytes[i] != expectedBytes[i] {
			t.Errorf("LongRawCol: byte mismatch at index %d: expected 0x%02X, got 0x%02X", i, expectedBytes[i], loadedBytes[i])
			break
		}
	}
}

// TestOracle_QueryFirst tests QueryFirst functionality
func TestOracle_QueryFirst(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	user, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = :1", 1)
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

// TestOracle_QueryOne tests QueryOne functionality
func TestOracle_QueryOne(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	user, err := typedb.QueryOne[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = :1", 1)
	if err != nil {
		t.Fatalf("QueryOne failed: %v", err)
	}

	if user.ID != 1 {
		t.Errorf("Expected user ID 1, got %d", user.ID)
	}
}

// TestOracle_Insert tests Insert by object functionality
func TestOracle_Insert(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Insert new user with unique email
	uniqueEmail := fmt.Sprintf("testinsert-%d@example.com", time.Now().UnixNano())
	newUser := &User{
		Name:  "Test Insert User",
		Email: uniqueEmail,
	}
	if err := typedb.Insert(ctx, db, newUser); err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Verify ID was set
	if newUser.ID == 0 {
		t.Error("User ID should be set after insert")
	}

	// Register cleanup that runs even on failure
	t.Cleanup(func() {
		if newUser.ID != 0 {
			db.Exec(ctx, "DELETE FROM users WHERE id = :1", newUser.ID)
		}
	})

	// Verify user was inserted
	loaded := &User{ID: newUser.ID}
	if err := typedb.Load(ctx, db, loaded); err != nil {
		t.Fatalf("Failed to load inserted user: %v", err)
	}

	if loaded.Name != "Test Insert User" {
		t.Errorf("Expected name 'Test Insert User', got '%s'", loaded.Name)
	}
	if loaded.Email != uniqueEmail {
		t.Errorf("Expected email '%s', got '%s'", uniqueEmail, loaded.Email)
	}
}

// TestOracle_InsertAndReturn tests InsertAndReturn functionality
func TestOracle_InsertAndReturn(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Get first user for foreign key
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE ROWNUM <= 1 ORDER BY id")
	if err != nil || firstUser == nil {
		t.Fatal("Need at least one user in database for foreign key")
	}

	// Insert post with RETURNING clause using unique title
	uniqueTitle := fmt.Sprintf("Test Post %d", time.Now().UnixNano())
	insertedPost, err := typedb.InsertAndReturn[*Post](ctx, db,
		"INSERT INTO posts (user_id, title, content, tags, metadata, created_at) VALUES (:1, :2, :3, :4, :5, TO_TIMESTAMP(:6, 'YYYY-MM-DD\"T\"HH24:MI:SS\"Z\"')) RETURNING id, user_id, title, content, tags, metadata, created_at",
		firstUser.ID, uniqueTitle, "Test content", `["go","database"]`, `{"test":true}`, "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("InsertAndReturn failed: %v", err)
	}

	// Register cleanup that runs even on failure
	t.Cleanup(func() {
		if insertedPost.ID != 0 {
			db.Exec(ctx, "DELETE FROM posts WHERE id = :1", insertedPost.ID)
		}
	})

	// Verify returned post
	if insertedPost.ID == 0 {
		t.Error("Post ID should be set")
	}
	if insertedPost.Title != uniqueTitle {
		t.Errorf("Expected title '%s', got '%s'", uniqueTitle, insertedPost.Title)
	}
	if insertedPost.UserID != firstUser.ID {
		t.Errorf("Expected UserID %d, got %d", firstUser.ID, insertedPost.UserID)
	}
}

// TestOracle_InsertAndGetId tests InsertAndGetId functionality
func TestOracle_InsertAndGetId(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Get first user for foreign key
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE ROWNUM <= 1 ORDER BY id")
	if err != nil || firstUser == nil {
		t.Fatal("Need at least one user in database for foreign key")
	}

	// Insert post and get ID (Oracle uses RETURNING) with unique title
	uniqueTitle := fmt.Sprintf("Test Post ID %d", time.Now().UnixNano())
	postID, err := typedb.InsertAndGetId(ctx, db,
		"INSERT INTO posts (user_id, title, content, tags, metadata, created_at) VALUES (:1, :2, :3, :4, :5, TO_TIMESTAMP(:6, 'YYYY-MM-DD\"T\"HH24:MI:SS\"Z\"')) RETURNING id",
		firstUser.ID, uniqueTitle, "Test content", `["go"]`, `{"test":true}`, "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("InsertAndGetId failed: %v", err)
	}

	// Register cleanup that runs even on failure
	t.Cleanup(func() {
		if postID != 0 {
			db.Exec(ctx, "DELETE FROM posts WHERE id = :1", postID)
		}
	})

	// Verify ID was returned
	if postID == 0 {
		t.Error("Post ID should not be zero")
	}

	// Verify post exists
	loaded := &Post{ID: int(postID)}
	if err := typedb.Load(ctx, db, loaded); err != nil {
		t.Fatalf("Failed to load inserted post: %v", err)
	}

	if loaded.Title != uniqueTitle {
		t.Errorf("Expected title '%s', got '%s'", uniqueTitle, loaded.Title)
	}
}

// TestOracle_Update tests Update by object functionality
func TestOracle_Update(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

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

// TestOracle_QueryFirst_NoRows tests QueryFirst with no rows (should return nil, no error)
func TestOracle_QueryFirst_NoRows(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Query for non-existent user
	user, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = :1", 99999)
	if err != nil {
		t.Fatalf("QueryFirst should not return error for no rows, got: %v", err)
	}

	if user != nil {
		t.Error("QueryFirst should return nil for no rows")
	}
}

// TestOracle_QueryOne_NoRows tests QueryOne with no rows (should return ErrNotFound)
func TestOracle_QueryOne_NoRows(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Query for non-existent user
	user, err := typedb.QueryOne[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = :1", 99999)
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

// TestOracle_QueryOne_MultipleRows tests QueryOne with multiple rows (should return error)
func TestOracle_QueryOne_MultipleRows(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
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

// TestOracle_Negative_InvalidQuery tests error handling for invalid SQL
func TestOracle_Negative_InvalidQuery(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
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

// TestOracle_Negative_ConstraintViolation tests error handling for constraint violations
func TestOracle_Negative_ConstraintViolation(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Get existing user email
	existingUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE ROWNUM <= 1 ORDER BY id")
	if err != nil || existingUser == nil {
		t.Fatal("Need at least one user in database")
	}

	// Try to insert user with duplicate email (unique constraint violation)
	duplicateUser := &User{
		Name:  "Duplicate Email User",
		Email: existingUser.Email, // Duplicate email
	}
	
	// Register cleanup using returned ID even if insert somehow succeeds
	t.Cleanup(func() {
		if duplicateUser.ID != 0 {
			db.Exec(ctx, "DELETE FROM users WHERE id = :1", duplicateUser.ID)
		}
	})
	
	err = typedb.Insert(ctx, db, duplicateUser)
	if err == nil {
		t.Fatal("Insert should fail with unique constraint violation")
	}

	// Verify error indicates constraint violation
	if !strings.Contains(err.Error(), "unique") && !strings.Contains(err.Error(), "duplicate") && !strings.Contains(err.Error(), "UNIQUE") {
		t.Errorf("Expected constraint violation error, got: %v", err)
	}
}

// TestOracle_Update_AutoTimestamp tests auto-updated timestamp functionality
func TestOracle_Update_AutoTimestamp(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

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

// TestOracle_Update_PartialUpdate tests partial update functionality
func TestOracle_Update_PartialUpdate(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

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

// TestOracle_Update_NonPartialUpdate tests that Update without partial update enabled updates all fields
func TestOracle_Update_NonPartialUpdate(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Use Post model which doesn't have partial update enabled
	firstPost, err := typedb.QueryFirst[*Post](ctx, db, "SELECT id, user_id, title, content, tags, metadata, created_at, updated_at FROM posts WHERE ROWNUM <= 1 ORDER BY id")
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
