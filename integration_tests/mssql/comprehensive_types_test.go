package main

import (
	"context"
	"strings"
	"testing"

	"github.com/TheBlackHowling/typedb"
	_ "github.com/microsoft/go-mssqldb" // SQL Server driver
)

func TestMSSQL_ComprehensiveTypes(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("sqlserver", getTestDSN())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer closeDB(t, db)

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
	defer closeDB(t, db)

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
		[]byte{0xCA, 0xFE, 0xBA, 0xBE},             // varbinary_col
		[]byte{0x01, 0x02, 0x03, 0x04, 0x05},       // varbinary_max_col
		[]byte{0x10, 0x11, 0x12, 0x13, 0x14, 0x15}, // image_col
		"2024-12-25",                           // date_col
		"15:30:45",                             // time_col
		"2024-12-25 15:30:45",                  // datetime_col
		"2024-12-25 15:30:45.1234567",          // datetime2_col
		"2024-12-25 15:30:45 +00:00",           // datetimeoffset_col
		"2024-12-25 15:30:00",                  // smalldatetime_col
		"550e8400-e29b-41d4-a716-446655440000", // uniqueidentifier_col
		"<root><test>roundtrip</test></root>",  // xml_col
		"/1/",                                  // hierarchyid_col
		"POINT(-122.4194 37.7749)",             // geography_col (San Francisco)
		"POINT(10 20)",                         // geometry_col
		"SQL_VARIANT test",                     // sql_variant_col
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
