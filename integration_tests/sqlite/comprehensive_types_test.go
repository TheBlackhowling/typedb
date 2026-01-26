package main

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/TheBlackHowling/typedb"
)

func TestSQLite_ComprehensiveTypes(t *testing.T) {
	db := setupTestDB(t)
	defer closeDB(t, db)
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
	defer closeDB(t, db)
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
		testID,                                 // id
		987654321,                              // integer_col
		3.14159,                                // real_col
		"1234.56",                              // numeric_col
		"test text content",                    // text_col
		"test_varchar",                         // varchar_col
		"test_char ",                           // char_col (padded)
		"test clob content",                    // clob_col
		[]byte{0xDE, 0xAD, 0xBE, 0xEF},         // blob_col
		"2024-12-25",                           // date_col
		"2024-12-25 15:30:45",                  // datetime_col
		"2024-12-25 15:30:45",                  // timestamp_col
		"15:30:45",                             // time_col
		true,                                   // boolean_col
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
