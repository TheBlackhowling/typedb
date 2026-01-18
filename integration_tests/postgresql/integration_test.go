package main

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/TheBlackHowling/typedb"
	_ "github.com/lib/pq" // PostgreSQL driver
)

func getTestDSN() string {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://user:password@localhost/typedb_examples?sslmode=disable"
	}
	return dsn
}

func TestPostgreSQL_QueryAll(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
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

func TestPostgreSQL_QueryFirst(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	user, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = $1", 1)
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

func TestPostgreSQL_QueryOne(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	user, err := typedb.QueryOne[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = $1", 1)
	if err != nil {
		t.Fatalf("QueryOne failed: %v", err)
	}

	if user.ID != 1 {
		t.Errorf("Expected user ID 1, got %d", user.ID)
	}
}

func TestPostgreSQL_Load(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
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

func TestPostgreSQL_LoadByField(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
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

func TestPostgreSQL_LoadByComposite(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
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

func TestPostgreSQL_Transaction(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	err = db.WithTx(ctx, func(tx *typedb.Tx) error {
		users, err := typedb.QueryAll[*User](ctx, tx, "SELECT id, name, email, created_at FROM users WHERE id = $1", 1)
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

func TestPostgreSQL_PostgreSQLSpecificFeatures(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Test PostgreSQL arrays and JSONB
	posts, err := typedb.QueryAll[*Post](ctx, db, "SELECT id, user_id, title, content, tags, metadata, created_at FROM posts ORDER BY id")
	if err != nil {
		t.Fatalf("QueryAll posts failed: %v", err)
	}

	if len(posts) == 0 {
		t.Fatal("Expected at least one post")
	}

	// Verify PostgreSQL-specific fields are deserialized as strings
	post := posts[0]
	if post.Tags == "" {
		t.Error("Tags (PostgreSQL array) should be loaded")
	}
	if post.Metadata == "" {
		t.Error("Metadata (JSONB) should be loaded")
	}
}

func TestPostgreSQL_ComprehensiveTypes(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
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
	if typeExample.TimestampCol == "" {
		t.Error("TimestampCol should be loaded")
	}
	if typeExample.JsonCol == "" {
		t.Error("JsonCol should be loaded")
	}
	if typeExample.JsonbCol == "" {
		t.Error("JsonbCol should be loaded")
	}
	if typeExample.UuidCol == "" {
		t.Error("UuidCol should be loaded")
	}

	// Test QueryAll with comprehensive types
	examples, err := typedb.QueryAll[*TypeExample](ctx, db, "SELECT id, small_int, integer_col, big_int, decimal_col, numeric_col, real_col, double_precision_col, money_col, varchar_col, char_col, text_col, bytea_col, date_col, time_col, timestamp_col, timestamptz_col, interval_col, boolean_col, json_col, jsonb_col, int_array, text_array, jsonb_array, uuid_col, inet_col, cidr_col, macaddr_col, point_col, bit_col, varbit_col, xml_col, created_at FROM type_examples")
	if err != nil {
		t.Fatalf("QueryAll type examples failed: %v", err)
	}

	if len(examples) == 0 {
		t.Fatal("Expected at least one type example")
	}
}

func TestPostgreSQL_ComprehensiveTypesRoundTrip(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
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
		id, small_int, integer_col, big_int, decimal_col, numeric_col,
		real_col, double_precision_col, money_col,
		varchar_col, char_col, text_col,
		bytea_col,
		date_col, time_col, time_tz_col, timestamp_col, timestamptz_col, interval_col,
		boolean_col,
		json_col, jsonb_col,
		smallint_array, int_array, bigint_array, real_array, double_precision_array,
		numeric_array, varchar_array, text_array, boolean_array, date_array,
		timestamp_array, json_array, jsonb_array, uuid_array, bytea_array,
		uuid_col,
		inet_col, cidr_col, macaddr_col, macaddr8_col,
		point_col, line_col, lseg_col, box_col, path_col, polygon_col, circle_col,
		int4range_col, int8range_col, numrange_col, tsrange_col, tstzrange_col, daterange_col,
		bit_col, varbit_col,
		tsvector_col, tsquery_col,
		xml_col
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
		$21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38,
		$39, $40, $41, $42, $43, $44, $45, $46, $47, $48, $49, $50, $51, $52, $53, $54, $55, $56,
		$57, $58, $59, $60
	)`

	// Insert test data
	_, err = db.Exec(ctx, insertSQL,
		testID,                         // id
		12345,                          // small_int
		987654321,                      // integer_col
		9223372036854775807,            // big_int
		"1234.567890",                  // decimal_col
		"9876.543210",                  // numeric_col
		"3.14159",                      // real_col
		"2.71828",                      // double_precision_col
		"$1,234.56",                    // money_col
		"test_varchar",                 // varchar_col
		"test_char  ",                  // char_col (padded)
		"test text content",            // text_col
		[]byte{0xDE, 0xAD, 0xBE, 0xEF}, // bytea_col
		"2024-12-25",                   // date_col
		"15:30:45",                     // time_col
		"15:30:45+00",                  // time_tz_col
		"2024-12-25 15:30:45",          // timestamp_col
		"2024-12-25 15:30:45+00",       // timestamptz_col
		"2 days 3 hours",               // interval_col
		true,                           // boolean_col
		`{"test": "json_value"}`,       // json_col
		`{"test": "jsonb_value", "nested": {"key": "value"}}`, // jsonb_col
		"{100,200,300}",                             // smallint_array
		"{1000,2000,3000}",                          // int_array
		"{9000000000000000000}",                     // bigint_array
		"{1.1,2.2,3.3}",                             // real_array
		"{1.11,2.22,3.33}",                          // double_precision_array
		"{10.5,20.5,30.5}",                          // numeric_array
		"{a,b,c}",                                   // varchar_array
		"{text1,text2,text3}",                       // text_array
		"{true,false,true}",                         // boolean_array
		"{2024-01-01,2024-01-02}",                   // date_array
		"{2024-01-01 10:00:00,2024-01-02 11:00:00}", // timestamp_array
		`{"{\"x\":1}","{\"y\":2}"}`,                 // json_array
		`{"{\"x\":1}","{\"y\":2}"}`,                 // jsonb_array
		"{550e8400-e29b-41d4-a716-446655440000,6ba7b810-9dad-11d1-80b4-00c04fd430c8}", // uuid_array
		`{"\\xDEADBEEF","\\xCAFEBABE"}`,                                               // bytea_array
		"550e8400-e29b-41d4-a716-446655440000",                                        // uuid_col
		"10.0.0.1",                                                                    // inet_col
		"10.0.0.0/8",                                                                  // cidr_col
		"08:00:2b:01:02:03",                                                           // macaddr_col
		"08:00:2b:ff:fe:01:02:03",                                                     // macaddr8_col
		"(5,10)",                                                                      // point_col
		"{1,-1,0}",                                                                    // line_col (LINE: {A,B,C} where Ax+By+C=0, line through (0,0) and (1,1))
		"[(0,0),(1,1)]",                                                               // lseg_col
		"(1,1),(0,0)",                                                                 // box_col
		"((0,0),(1,1),(2,2))",                                                         // path_col
		"((0,0),(1,0),(1,1),(0,1))",                                                   // polygon_col
		"<(0,0),5>",                                                                   // circle_col
		"[1,100)",                                                                     // int4range_col
		"[1000,2000)",                                                                 // int8range_col
		"[1.5,99.5)",                                                                  // numrange_col
		"[2024-01-01 00:00:00,2024-01-02 00:00:00)",       // tsrange_col
		"[2024-01-01 00:00:00+00,2024-01-02 00:00:00+00)", // tstzrange_col
		"[2024-01-01,2024-01-02)",                         // daterange_col
		"10101010",                                        // bit_col
		"11110000",                                        // varbit_col
		"test roundtrip",                                  // tsvector_col
		"test & roundtrip",                                // tsquery_col
		"<root><test>roundtrip</test></root>",             // xml_col
	)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Clean up after test
	defer func() {
		db.Exec(ctx, "DELETE FROM type_examples WHERE id = $1", testID)
	}()

	// Query it back
	loaded := &TypeExample{ID: testID}
	if err := typedb.Load(ctx, db, loaded); err != nil {
		t.Fatalf("Failed to load inserted data: %v", err)
	}

	// Validate every field
	if loaded.SmallInt != 12345 {
		t.Errorf("SmallInt: expected 12345, got %d", loaded.SmallInt)
	}
	if loaded.IntegerCol != 987654321 {
		t.Errorf("IntegerCol: expected 987654321, got %d", loaded.IntegerCol)
	}
	if loaded.BigInt != 9223372036854775807 {
		t.Errorf("BigInt: expected 9223372036854775807, got %d", loaded.BigInt)
	}
	if loaded.DecimalCol == "" {
		t.Error("DecimalCol: should not be empty")
	}
	if loaded.NumericCol == "" {
		t.Error("NumericCol: should not be empty")
	}
	if loaded.RealCol == "" {
		t.Error("RealCol: should not be empty")
	}
	if loaded.DoublePrecisionCol == "" {
		t.Error("DoublePrecisionCol: should not be empty")
	}
	if loaded.MoneyCol == "" {
		t.Error("MoneyCol: should not be empty")
	}
	if loaded.VarcharCol != "test_varchar" {
		t.Errorf("VarcharCol: expected 'test_varchar', got '%s'", loaded.VarcharCol)
	}
	if loaded.CharCol == "" {
		t.Error("CharCol: should not be empty")
	}
	if loaded.TextCol != "test text content" {
		t.Errorf("TextCol: expected 'test text content', got '%s'", loaded.TextCol)
	}
	if loaded.ByteaCol == "" {
		t.Error("ByteaCol: should not be empty")
	}
	if loaded.DateCol == "" || !strings.Contains(loaded.DateCol, "2024-12-25") {
		t.Errorf("DateCol: expected to contain '2024-12-25', got '%s'", loaded.DateCol)
	}
	if loaded.TimeCol == "" {
		t.Error("TimeCol: should not be empty")
	}
	if loaded.TimeTzCol == "" {
		t.Error("TimeTzCol: should not be empty")
	}
	if loaded.TimestampCol == "" {
		t.Error("TimestampCol: should not be empty")
	}
	if loaded.TimestamptzCol == "" {
		t.Error("TimestamptzCol: should not be empty")
	}
	if loaded.IntervalCol == "" {
		t.Error("IntervalCol: should not be empty")
	}
	if !loaded.BooleanCol {
		t.Error("BooleanCol: expected true, got false")
	}
	if loaded.JsonCol == "" {
		t.Error("JsonCol: should not be empty")
	}
	if loaded.JsonbCol == "" {
		t.Error("JsonbCol: should not be empty")
	}
	if loaded.SmallintArray == "" {
		t.Error("SmallintArray: should not be empty")
	}
	if loaded.IntArray == "" {
		t.Error("IntArray: should not be empty")
	}
	if loaded.BigintArray == "" {
		t.Error("BigintArray: should not be empty")
	}
	if loaded.RealArray == "" {
		t.Error("RealArray: should not be empty")
	}
	if loaded.DoublePrecisionArray == "" {
		t.Error("DoublePrecisionArray: should not be empty")
	}
	if loaded.NumericArray == "" {
		t.Error("NumericArray: should not be empty")
	}
	if loaded.VarcharArray == "" {
		t.Error("VarcharArray: should not be empty")
	}
	if loaded.TextArray == "" {
		t.Error("TextArray: should not be empty")
	}
	if loaded.BooleanArray == "" {
		t.Error("BooleanArray: should not be empty")
	}
	if loaded.DateArray == "" {
		t.Error("DateArray: should not be empty")
	}
	if loaded.TimestampArray == "" {
		t.Error("TimestampArray: should not be empty")
	}
	if loaded.JsonArray == "" {
		t.Error("JsonArray: should not be empty")
	}
	if loaded.JsonbArray == "" {
		t.Error("JsonbArray: should not be empty")
	}
	if loaded.UuidArray == "" {
		t.Error("UuidArray: should not be empty")
	}
	if loaded.ByteaArray == "" {
		t.Error("ByteaArray: should not be empty")
	}
	if loaded.UuidCol == "" {
		t.Error("UuidCol: should not be empty")
	}
	if loaded.InetCol == "" {
		t.Error("InetCol: should not be empty")
	}
	if loaded.CidrCol == "" {
		t.Error("CidrCol: should not be empty")
	}
	if loaded.MacaddrCol == "" {
		t.Error("MacaddrCol: should not be empty")
	}
	if loaded.Macaddr8Col == "" {
		t.Error("Macaddr8Col: should not be empty")
	}
	if loaded.PointCol == "" {
		t.Error("PointCol: should not be empty")
	}
	if loaded.LineCol == "" {
		t.Error("LineCol: should not be empty")
	}
	if loaded.LsegCol == "" {
		t.Error("LsegCol: should not be empty")
	}
	if loaded.BoxCol == "" {
		t.Error("BoxCol: should not be empty")
	}
	if loaded.PathCol == "" {
		t.Error("PathCol: should not be empty")
	}
	if loaded.PolygonCol == "" {
		t.Error("PolygonCol: should not be empty")
	}
	if loaded.CircleCol == "" {
		t.Error("CircleCol: should not be empty")
	}
	if loaded.Int4rangeCol == "" {
		t.Error("Int4rangeCol: should not be empty")
	}
	if loaded.Int8rangeCol == "" {
		t.Error("Int8rangeCol: should not be empty")
	}
	if loaded.NumrangeCol == "" {
		t.Error("NumrangeCol: should not be empty")
	}
	if loaded.TsrangeCol == "" {
		t.Error("TsrangeCol: should not be empty")
	}
	if loaded.TstzrangeCol == "" {
		t.Error("TstzrangeCol: should not be empty")
	}
	if loaded.DaterangeCol == "" {
		t.Error("DaterangeCol: should not be empty")
	}
	if loaded.BitCol == "" {
		t.Error("BitCol: should not be empty")
	}
	if loaded.VarbitCol == "" {
		t.Error("VarbitCol: should not be empty")
	}
	if loaded.TsvectorCol == "" {
		t.Error("TsvectorCol: should not be empty")
	}
	if loaded.TsqueryCol == "" {
		t.Error("TsqueryCol: should not be empty")
	}
	if loaded.XmlCol == "" {
		t.Error("XmlCol: should not be empty")
	}
	if loaded.CreatedAt == "" {
		t.Error("CreatedAt: should not be empty")
	}
}

// TestPostgreSQL_Insert tests Insert by object functionality
func TestPostgreSQL_Insert(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
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
		db.Exec(ctx, "DELETE FROM users WHERE id = $1", newUser.ID)
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

// TestPostgreSQL_InsertAndReturn tests InsertAndReturn functionality
func TestPostgreSQL_InsertAndReturn(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
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

	// Insert post with RETURNING clause
	insertedPost, err := typedb.InsertAndReturn[*Post](ctx, db,
		"INSERT INTO posts (user_id, title, content, tags, metadata, created_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, user_id, title, content, tags, metadata, created_at",
		firstUser.ID, "Test Post", "Test content", "{\"go\",\"database\"}", "{\"test\":true}", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("InsertAndReturn failed: %v", err)
	}

	// Clean up
	defer func() {
		db.Exec(ctx, "DELETE FROM posts WHERE id = $1", insertedPost.ID)
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

// TestPostgreSQL_InsertAndGetId tests InsertAndGetId functionality
func TestPostgreSQL_InsertAndGetId(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
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

	// Insert post and get ID
	postID, err := typedb.InsertAndGetId(ctx, db,
		"INSERT INTO posts (user_id, title, content, tags, metadata, created_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		firstUser.ID, "Test Post ID", "Test content", "{\"go\"}", "{\"test\":true}", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("InsertAndGetId failed: %v", err)
	}

	// Clean up
	defer func() {
		db.Exec(ctx, "DELETE FROM posts WHERE id = $1", postID)
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

// TestPostgreSQL_Update tests Update by object functionality
func TestPostgreSQL_Update(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
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

// TestPostgreSQL_Update_AutoTimestamp tests auto-updated timestamp functionality
func TestPostgreSQL_Update_AutoTimestamp(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
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
	// If original was empty/NULL, that's fine - just verify it's now set
	// If original had a value, verify it changed
	if originalUpdatedAt != "" && updatedUser.UpdatedAt == originalUpdatedAt {
		t.Error("UpdatedAt should change after update")
	}
}

// TestPostgreSQL_QueryFirst_NoRows tests QueryFirst with no rows (should return nil, no error)
func TestPostgreSQL_QueryFirst_NoRows(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Query for non-existent user
	user, err := typedb.QueryFirst[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = $1", 99999)
	if err != nil {
		t.Fatalf("QueryFirst should not return error for no rows, got: %v", err)
	}

	if user != nil {
		t.Error("QueryFirst should return nil for no rows")
	}
}

// TestPostgreSQL_QueryOne_NoRows tests QueryOne with no rows (should return ErrNotFound)
func TestPostgreSQL_QueryOne_NoRows(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Query for non-existent user
	user, err := typedb.QueryOne[*User](ctx, db, "SELECT id, name, email, created_at FROM users WHERE id = $1", 99999)
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

// TestPostgreSQL_QueryOne_MultipleRows tests QueryOne with multiple rows (should return error)
func TestPostgreSQL_QueryOne_MultipleRows(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
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

// TestPostgreSQL_Negative_InvalidQuery tests error handling for invalid SQL
func TestPostgreSQL_Negative_InvalidQuery(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
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

// TestPostgreSQL_Negative_ConstraintViolation tests error handling for constraint violations
func TestPostgreSQL_Negative_ConstraintViolation(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("postgres", getTestDSN())
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
			db.Exec(ctx, "DELETE FROM users WHERE id = $1", duplicateUser.ID)
		}
		t.Fatal("Insert should fail with unique constraint violation")
	}

	// Verify error indicates constraint violation
	if !strings.Contains(err.Error(), "unique") && !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("Expected constraint violation error, got: %v", err)
	}
}
