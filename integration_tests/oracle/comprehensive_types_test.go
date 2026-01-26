package main

import (
	"context"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/TheBlackHowling/typedb"
	_ "github.com/sijms/go-ora/v2" // Oracle driver
)

func TestOracle_ComprehensiveTypes(t *testing.T) {
	ctx := context.Background()
	db, err := typedb.Open("oracle", getTestDSN())
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
	defer closeDB(t, db)

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
		testID,                         // id
		"1234.56",                      // number_col
		"9876.54",                      // number_precision_col
		"1111.22",                      // number_scale_col
		3.14159,                        // float_col
		123.4567,                       // float_precision_col
		2.71828,                        // binary_float_col
		1.41421,                        // binary_double_col
		"test_char ",                   // char_col (padded)
		"test_varchar2",                // varchar2_col
		"test_varchar",                 // varchar_col
		"test_nchar",                   // nchar_col (NCHAR(10), will be padded)
		"test_nvarchar2",               // nvarchar2_col
		"test clob content",            // clob_col
		"test nclob content",           // nclob_col
		"test long content",            // long_col
		[]byte{0xDE, 0xAD, 0xBE, 0xEF}, // raw_col
		[]byte{0x01, 0x02, 0x03, 0x04}, // blob_col
		"2024-12-25",                   // date_col (used in TO_DATE)
		"2024-12-25 15:30:45",          // timestamp_col (used in TO_TIMESTAMP)
		"2024-12-25 15:30:45.123456",   // timestamp_precision_col (used in TO_TIMESTAMP)
		"2024-12-25 15:30:45 +00:00",   // timestamp_tz_col (used in TO_TIMESTAMP_TZ)
		"2024-12-25 15:30:45",          // timestamp_ltz_col (used in TO_TIMESTAMP)
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
	defer closeDB(t, db)

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
