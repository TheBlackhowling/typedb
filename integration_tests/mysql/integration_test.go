package main

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/TheBlackHowling/typedb"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

func getTestDSN() string {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "user:password@tcp(localhost:3306)/typedb_examples?parseTime=true"
	}
	return dsn
}

func TestMySQL_QueryAll(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
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

	// Verify first user
	if users[0].ID == 0 {
		t.Error("User ID should not be zero")
	}
	if users[0].Name == "" {
		t.Error("User name should not be empty")
	}
	if users[0].Email == "" {
		t.Error("User email should not be empty")
	}
}

func TestMySQL_QueryFirst(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

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

func TestMySQL_QueryOne(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	user, err := typedb.QueryOne[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = ?", 1)
	if err != nil {
		t.Fatalf("QueryOne failed: %v", err)
	}

	if user.ID != 1 {
		t.Errorf("Expected user ID 1, got %d", user.ID)
	}
}

func TestMySQL_Load(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
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

func TestMySQL_LoadByField(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
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

func TestMySQL_LoadByComposite(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
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

func TestMySQL_Transaction(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	err = db.WithTx(ctx, func(tx *typedb.Tx) error {
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

func TestMySQL_MySQLSpecificFeatures(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Test MySQL JSON
	posts, err := typedb.QueryAll[*Post](ctx, db, "SELECT id, user_id, title, content, tags, metadata, created_at FROM posts ORDER BY id")
	if err != nil {
		t.Fatalf("QueryAll posts failed: %v", err)
	}

	if len(posts) == 0 {
		t.Fatal("Expected at least one post")
	}

	// Verify MySQL JSON fields are deserialized as strings
	post := posts[0]
	if post.Tags == "" {
		t.Error("Tags (MySQL JSON) should be loaded")
	}
	if post.Metadata == "" {
		t.Error("Metadata (MySQL JSON) should be loaded")
	}
}

func TestMySQL_ComprehensiveTypes(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
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
	examples, err := typedb.QueryAll[*TypeExample](ctx, db, "SELECT id, tiny_int, tiny_int_unsigned, small_int, small_int_unsigned, medium_int, medium_int_unsigned, integer_col, integer_col_unsigned, big_int, big_int_unsigned, decimal_col, decimal_col_unsigned, numeric_col, numeric_col_unsigned, float_col, float_col_precision, double_col, double_col_precision, bit_col, bit_col_64, char_col, varchar_col, binary_col, varbinary_col, tinytext_col, text_col, mediumtext_col, longtext_col, enum_col, set_col, tinyblob_col, blob_col, mediumblob_col, longblob_col, date_col, time_col, datetime_col, timestamp_col, year_col, json_col, geometry_col, point_col, linestring_col, polygon_col, multipoint_col, multilinestring_col, multipolygon_col, geometrycollection_col, created_at FROM type_examples")
	if err != nil {
		t.Fatalf("QueryAll type examples failed: %v", err)
	}

	if len(examples) == 0 {
		t.Fatal("Expected at least one type example")
	}
}

func TestMySQL_ComprehensiveTypesRoundTrip(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
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
		id, tiny_int, tiny_int_unsigned, small_int, small_int_unsigned,
		medium_int, medium_int_unsigned, integer_col, integer_col_unsigned,
		big_int, big_int_unsigned,
		decimal_col, decimal_col_unsigned, numeric_col, numeric_col_unsigned,
		float_col, float_col_precision, double_col, double_col_precision,
		bit_col, bit_col_64,
		char_col, varchar_col, binary_col, varbinary_col,
		tinytext_col, text_col, mediumtext_col, longtext_col,
		tinyblob_col, blob_col, mediumblob_col, longblob_col,
		enum_col, set_col,
		date_col, time_col, datetime_col, timestamp_col, year_col,
		json_col,
		geometry_col, point_col, linestring_col, polygon_col,
		multipoint_col, multilinestring_col, multipolygon_col, geometrycollection_col
	) VALUES (
		?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
		?, ST_GeomFromText(?), ST_GeomFromText(?), ST_GeomFromText(?), ST_GeomFromText(?),
		ST_GeomFromText(?), ST_GeomFromText(?), ST_GeomFromText(?), ST_GeomFromText(?)
	)`

	// Insert test data
	_, err = db.Exec(ctx, insertSQL,
		testID,                 // id
		100,                    // tiny_int
		200,                    // tiny_int_unsigned
		12345,                  // small_int
		54321,                  // small_int_unsigned
		8388600,                // medium_int
		16777200,               // medium_int_unsigned
		987654321,              // integer_col
		4294967290,             // integer_col_unsigned
		9223372036854775800,    // big_int
		"18446744073709551600", // big_int_unsigned (MySQL returns as string)
		"1234.567890",          // decimal_col
		"9876.543210",          // decimal_col_unsigned
		"1111.222222",          // numeric_col
		"2222.333333",          // numeric_col_unsigned
		3.14159,                // float_col
		123.4567,               // float_col_precision
		2.71828,                // double_col
		12345.67890123,         // double_col_precision
		[]byte{0xAA},           // bit_col
		[]byte{0xF0, 0xF0, 0xF0, 0xF0, 0xF0, 0xF0, 0xF0, 0xF0}, // bit_col_64
		"test_char  ",                                    // char_col (padded)
		"test_varchar",                                   // varchar_col
		[]byte{0xDE, 0xAD, 0xBE, 0xEF},                   // binary_col
		[]byte{0xCA, 0xFE, 0xBA, 0xBE},                   // varbinary_col
		"tinytext test",                                  // tinytext_col
		"text test content",                              // text_col
		"mediumtext test content",                        // mediumtext_col
		"longtext test content with much longer content", // longtext_col
		[]byte{0x01, 0x02, 0x03},                         // tinyblob_col
		[]byte{0x04, 0x05, 0x06, 0x07},                   // blob_col
		[]byte{0x08, 0x09, 0x0A, 0x0B, 0x0C},             // mediumblob_col
		[]byte{0x10, 0x11, 0x12, 0x13, 0x14, 0x15},       // longblob_col
		"value1",                                // enum_col
		"option1,option2",                       // set_col
		"2024-12-25",                            // date_col
		"15:30:45",                              // time_col
		"2024-12-25 15:30:45",                   // datetime_col
		"2024-12-25 15:30:45",                   // timestamp_col
		2024,                                    // year_col
		`{"test": "json_value", "number": 42}`,  // json_col
		"POINT(5 10)",                           // geometry_col
		"POINT(10 20)",                          // point_col
		"LINESTRING(0 0,1 1,2 2)",               // linestring_col
		"POLYGON((0 0,1 0,1 1,0 1,0 0))",        // polygon_col
		"MULTIPOINT((0 0),(1 1))",               // multipoint_col
		"MULTILINESTRING((0 0,1 1),(2 2,3 3))",  // multilinestring_col
		"MULTIPOLYGON(((0 0,1 0,1 1,0 1,0 0)))", // multipolygon_col
		"GEOMETRYCOLLECTION(POINT(0 0))",        // geometrycollection_col
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
	if loaded.TinyInt != 100 {
		t.Errorf("TinyInt: expected 100, got %d", loaded.TinyInt)
	}
	if loaded.TinyIntUnsigned != 200 {
		t.Errorf("TinyIntUnsigned: expected 200, got %d", loaded.TinyIntUnsigned)
	}
	if loaded.SmallInt != 12345 {
		t.Errorf("SmallInt: expected 12345, got %d", loaded.SmallInt)
	}
	if loaded.SmallIntUnsigned != 54321 {
		t.Errorf("SmallIntUnsigned: expected 54321, got %d", loaded.SmallIntUnsigned)
	}
	if loaded.MediumInt != 8388600 {
		t.Errorf("MediumInt: expected 8388600, got %d", loaded.MediumInt)
	}
	if loaded.MediumIntUnsigned != 16777200 {
		t.Errorf("MediumIntUnsigned: expected 16777200, got %d", loaded.MediumIntUnsigned)
	}
	if loaded.IntegerCol != 987654321 {
		t.Errorf("IntegerCol: expected 987654321, got %d", loaded.IntegerCol)
	}
	if loaded.IntegerColUnsigned != 4294967290 {
		t.Errorf("IntegerColUnsigned: expected 4294967290, got %d", loaded.IntegerColUnsigned)
	}
	if loaded.BigInt != 9223372036854775800 {
		t.Errorf("BigInt: expected 9223372036854775800, got %d", loaded.BigInt)
	}
	if loaded.BigIntUnsigned != 18446744073709551600 {
		t.Errorf("BigIntUnsigned: expected 18446744073709551600, got %d", loaded.BigIntUnsigned)
	}
	if loaded.DecimalCol == "" {
		t.Error("DecimalCol: should not be empty")
	}
	if loaded.DecimalColUnsigned == "" {
		t.Error("DecimalColUnsigned: should not be empty")
	}
	if loaded.NumericCol == "" {
		t.Error("NumericCol: should not be empty")
	}
	if loaded.NumericColUnsigned == "" {
		t.Error("NumericColUnsigned: should not be empty")
	}
	if loaded.FloatCol == "" {
		t.Error("FloatCol: should not be empty")
	}
	if loaded.FloatColPrecision == "" {
		t.Error("FloatColPrecision: should not be empty")
	}
	if loaded.DoubleCol == "" {
		t.Error("DoubleCol: should not be empty")
	}
	if loaded.DoubleColPrecision == "" {
		t.Error("DoubleColPrecision: should not be empty")
	}
	if loaded.BitCol == "" {
		t.Error("BitCol: should not be empty")
	}
	if loaded.BitCol64 == "" {
		t.Error("BitCol64: should not be empty")
	}
	if loaded.CharCol == "" {
		t.Error("CharCol: should not be empty")
	}
	if loaded.VarcharCol != "test_varchar" {
		t.Errorf("VarcharCol: expected 'test_varchar', got '%s'", loaded.VarcharCol)
	}
	if loaded.BinaryCol == "" {
		t.Error("BinaryCol: should not be empty")
	}
	if loaded.VarbinaryCol == "" {
		t.Error("VarbinaryCol: should not be empty")
	}
	if loaded.TinytextCol != "tinytext test" {
		t.Errorf("TinytextCol: expected 'tinytext test', got '%s'", loaded.TinytextCol)
	}
	if loaded.TextCol != "text test content" {
		t.Errorf("TextCol: expected 'text test content', got '%s'", loaded.TextCol)
	}
	if loaded.MediumtextCol != "mediumtext test content" {
		t.Errorf("MediumtextCol: expected 'mediumtext test content', got '%s'", loaded.MediumtextCol)
	}
	if loaded.LongtextCol == "" {
		t.Error("LongtextCol: should not be empty")
	}
	if loaded.TinyblobCol == "" {
		t.Error("TinyblobCol: should not be empty")
	}
	if loaded.BlobCol == "" {
		t.Error("BlobCol: should not be empty")
	}
	if loaded.MediumblobCol == "" {
		t.Error("MediumblobCol: should not be empty")
	}
	if loaded.LongblobCol == "" {
		t.Error("LongblobCol: should not be empty")
	}
	if loaded.EnumCol != "value1" {
		t.Errorf("EnumCol: expected 'value1', got '%s'", loaded.EnumCol)
	}
	if loaded.SetCol == "" {
		t.Error("SetCol: should not be empty")
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
	if loaded.TimestampCol == "" {
		t.Error("TimestampCol: should not be empty")
	}
	if loaded.YearCol != 2024 {
		t.Errorf("YearCol: expected 2024, got %d", loaded.YearCol)
	}
	if loaded.JsonCol == "" {
		t.Error("JsonCol: should not be empty")
	}
	if loaded.GeometryCol == "" {
		t.Error("GeometryCol: should not be empty")
	}
	if loaded.PointCol == "" {
		t.Error("PointCol: should not be empty")
	}
	if loaded.LinestringCol == "" {
		t.Error("LinestringCol: should not be empty")
	}
	if loaded.PolygonCol == "" {
		t.Error("PolygonCol: should not be empty")
	}
	if loaded.MultipointCol == "" {
		t.Error("MultipointCol: should not be empty")
	}
	if loaded.MultilinestringCol == "" {
		t.Error("MultilinestringCol: should not be empty")
	}
	if loaded.MultipolygonCol == "" {
		t.Error("MultipolygonCol: should not be empty")
	}
	if loaded.GeometrycollectionCol == "" {
		t.Error("GeometrycollectionCol: should not be empty")
	}
	if loaded.CreatedAt == "" {
		t.Error("CreatedAt: should not be empty")
	}
}

// TestMySQL_Insert tests Insert by object functionality
func TestMySQL_Insert(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
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
		db.Exec(ctx, "DELETE FROM users WHERE id = ?", newUser.ID)
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

// TestMySQL_InsertAndReturn tests InsertAndReturn functionality
func TestMySQL_InsertAndReturn(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Get first user for foreign key
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users ORDER BY id LIMIT 1")
	if err != nil || firstUser == nil {
		t.Fatal("Need at least one user in database for foreign key")
	}

	// MySQL doesn't support RETURNING, so we'll test InsertAndGetId instead
	// For InsertAndReturn, we'd need to use a separate SELECT after INSERT
	// But MySQL's LastInsertId() works with InsertAndGetId
	postID, err := typedb.InsertAndGetId(ctx, db,
		"INSERT INTO posts (user_id, title, content, tags, metadata, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		firstUser.ID, "Test Post", "Test content", `["go","database"]`, `{"test":true}`, "2024-01-01 00:00:00")
	if err != nil {
		t.Fatalf("InsertAndGetId failed: %v", err)
	}

	// Clean up
	defer func() {
		db.Exec(ctx, "DELETE FROM posts WHERE id = ?", postID)
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

	if loaded.Title != "Test Post" {
		t.Errorf("Expected title 'Test Post', got '%s'", loaded.Title)
	}
}

// TestMySQL_InsertAndGetId tests InsertAndGetId functionality
func TestMySQL_InsertAndGetId(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Get first user for foreign key
	firstUser, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users ORDER BY id LIMIT 1")
	if err != nil || firstUser == nil {
		t.Fatal("Need at least one user in database for foreign key")
	}

	// Insert post and get ID (MySQL uses LastInsertId)
	postID, err := typedb.InsertAndGetId(ctx, db,
		"INSERT INTO posts (user_id, title, content, tags, metadata, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		firstUser.ID, "Test Post ID", "Test content", `["go"]`, `{"test":true}`, "2024-01-01 00:00:00")
	if err != nil {
		t.Fatalf("InsertAndGetId failed: %v", err)
	}

	// Clean up
	defer func() {
		db.Exec(ctx, "DELETE FROM posts WHERE id = ?", postID)
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

// TestMySQL_Update tests Update by object functionality
func TestMySQL_Update(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

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

// TestMySQL_QueryFirst_NoRows tests QueryFirst with no rows (should return nil, no error)
func TestMySQL_QueryFirst_NoRows(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Query for non-existent user
	user, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = ?", 99999)
	if err != nil {
		t.Fatalf("QueryFirst should not return error for no rows, got: %v", err)
	}

	if user != nil {
		t.Error("QueryFirst should return nil for no rows")
	}
}

// TestMySQL_QueryOne_NoRows tests QueryOne with no rows (should return ErrNotFound)
func TestMySQL_QueryOne_NoRows(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

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

// TestMySQL_QueryOne_MultipleRows tests QueryOne with multiple rows (should return error)
func TestMySQL_QueryOne_MultipleRows(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
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

// TestMySQL_Negative_InvalidQuery tests error handling for invalid SQL
func TestMySQL_Negative_InvalidQuery(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
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

// TestMySQL_Negative_ConstraintViolation tests error handling for constraint violations
func TestMySQL_Negative_ConstraintViolation(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

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
	if !strings.Contains(err.Error(), "unique") && !strings.Contains(err.Error(), "duplicate") && !strings.Contains(err.Error(), "Duplicate") {
		t.Errorf("Expected constraint violation error, got: %v", err)
	}
}

// TestMySQL_Update_AutoTimestamp tests auto-updated timestamp functionality
func TestMySQL_Update_AutoTimestamp(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("mysql", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

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

	// Wait a moment before update to ensure timestamp will be different
	time.Sleep(2 * time.Second)

	// Update user - UpdatedAt should be auto-populated
	userToUpdate := &User{
		ID:   firstUser.ID,
		Name: "Updated Name for Timestamp Test",
		// UpdatedAt is not set - should be auto-populated with database timestamp
	}
	if err := typedb.Update(ctx, db, userToUpdate); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

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
	// If original had a value, it should have changed (new >= original after delay)
	if originalUpdatedAt == "" {
		// Original was empty, verify it's now set
		if updatedUser.UpdatedAt == "" {
			t.Error("UpdatedAt should be populated after update")
		}
	} else {
		// Original had a value, verify it changed
		if updatedUser.UpdatedAt == originalUpdatedAt {
			t.Errorf("UpdatedAt should have changed after update. Original: %q, New: %q", originalUpdatedAt, updatedUser.UpdatedAt)
		}
	}
}
