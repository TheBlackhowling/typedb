package typedb

import (
	"reflect"
	"strings"
	"testing"
)

func TestIsZeroOrNil(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		// Nil pointer types
		{"nil_ptr", (*int)(nil), true},
		{"nil_slice", ([]int)(nil), true},
		{"nil_map", (map[string]int)(nil), true},
		{"nil_chan", (chan int)(nil), true},
		{"nil_func", (func())(nil), true},

		// Non-nil pointer types
		{"non_nil_ptr", intPtr(42), false},
		{"non_nil_slice", []int{1, 2, 3}, false},
		{"non_nil_map", map[string]int{"a": 1}, false},
		{"non_nil_chan", make(chan int), false},
		{"non_nil_func", func() {}, false},

		// Empty but non-nil
		{"empty_slice", []int{}, false},        // Empty slice is not nil
		{"empty_map", map[string]int{}, false}, // Empty map is not nil

		// String types
		{"empty_string", "", true},
		{"non_empty_string", "hello", false},

		// Integer types
		{"int_zero", 0, true},
		{"int_non_zero", 42, false},
		{"int8_zero", int8(0), true},
		{"int8_non_zero", int8(42), false},
		{"int16_zero", int16(0), true},
		{"int16_non_zero", int16(42), false},
		{"int32_zero", int32(0), true},
		{"int32_non_zero", int32(42), false},
		{"int64_zero", int64(0), true},
		{"int64_non_zero", int64(42), false},

		// Unsigned integer types
		{"uint_zero", uint(0), true},
		{"uint_non_zero", uint(42), false},
		{"uint8_zero", uint8(0), true},
		{"uint8_non_zero", uint8(42), false},
		{"uint16_zero", uint16(0), true},
		{"uint16_non_zero", uint16(42), false},
		{"uint32_zero", uint32(0), true},
		{"uint32_non_zero", uint32(42), false},
		{"uint64_zero", uint64(0), true},
		{"uint64_non_zero", uint64(42), false},
		{"uintptr_zero", uintptr(0), true},
		{"uintptr_non_zero", uintptr(42), false},

		// Float types
		{"float32_zero", float32(0), true},
		{"float32_non_zero", float32(3.14), false},
		{"float64_zero", float64(0), true},
		{"float64_non_zero", float64(3.14), false},

		// Bool types
		{"bool_false", false, true},
		{"bool_true", true, false},

		// Struct types (default case - should return false)
		{"struct_zero", struct{}{}, false},
		{"struct_with_fields", struct{ X int }{X: 0}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := reflect.ValueOf(tt.value)
			result := isZeroOrNil(v)
			if result != tt.expected {
				t.Errorf("isZeroOrNil(%v) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}

	// Test nil interface separately - reflect.ValueOf((interface{})(nil)) returns invalid value
	// We need to test it through a struct field or variable
	t.Run("nil_interface", func(t *testing.T) {
		var iface interface{} = (*int)(nil) // Interface containing nil pointer
		v := reflect.ValueOf(iface)
		result := isZeroOrNil(v)
		// Interface containing nil pointer should be considered nil
		if !result {
			t.Errorf("isZeroOrNil(interface{}(nil pointer)) = %v, want true", result)
		}
	})

	// Test interface containing non-nil value
	t.Run("interface_with_value", func(t *testing.T) {
		var iface interface{} = 42
		v := reflect.ValueOf(iface)
		result := isZeroOrNil(v)
		// Interface containing non-zero value should not be nil
		if result {
			t.Errorf("isZeroOrNil(interface{}(42)) = %v, want false", result)
		}
	})
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}

// BadTableNameModel1 has TableName() returning wrong number of results
type BadTableNameModel1 struct {
	Model
	ID int `db:"id" load:"primary"`
}

func (m *BadTableNameModel1) TableName() (string, error) {
	return "users", nil // Returns 2 values instead of 1
}

// BadTableNameModel2 has TableName() returning empty string
type BadTableNameModel2 struct {
	Model
	ID int `db:"id" load:"primary"`
}

func (m *BadTableNameModel2) TableName() string {
	return "" // Returns empty string
}

// BadTableNameModel3 has TableName() returning no values
type BadTableNameModel3 struct {
	Model
	ID int `db:"id" load:"primary"`
}

func (m *BadTableNameModel3) TableName() {
	// Returns nothing
}

func TestGetTableName_Success(t *testing.T) {
	model := &InsertModel{}
	tableName, err := getTableName(model)
	if err != nil {
		t.Fatalf("getTableName failed: %v", err)
	}
	if tableName != "users" {
		t.Errorf("Expected table name 'users', got %q", tableName)
	}
}

func TestGetTableName_NoMethod(t *testing.T) {
	model := &NoTableNameModel{}
	_, err := getTableName(model)
	if err == nil {
		t.Fatal("Expected error for missing TableName() method")
	}
	if !strings.Contains(err.Error(), "TableName()") {
		t.Errorf("Expected error about TableName(), got: %v", err)
	}
}

func TestGetTableName_EmptyString(t *testing.T) {
	model := &BadTableNameModel2{}
	_, err := getTableName(model)
	if err == nil {
		t.Fatal("Expected error for empty TableName() return")
	}
	if !strings.Contains(err.Error(), "empty string") {
		t.Errorf("Expected error about empty string, got: %v", err)
	}
}

func TestGetTableName_WrongReturnCount(t *testing.T) {
	model := &BadTableNameModel1{}
	_, err := getTableName(model)
	if err == nil {
		t.Fatal("Expected error for wrong return count")
	}
	if !strings.Contains(err.Error(), "exactly one value") {
		t.Errorf("Expected error about return count, got: %v", err)
	}
}

func TestGetTableName_NoReturnValue(t *testing.T) {
	model := &BadTableNameModel3{}
	_, err := getTableName(model)
	if err == nil {
		t.Fatal("Expected error for no return value")
	}
	if !strings.Contains(err.Error(), "exactly one value") {
		t.Errorf("Expected error about return count, got: %v", err)
	}
}

func TestBuildReturningClause(t *testing.T) {
	tests := []struct {
		name             string
		driverName       string
		primaryKeyColumn string
		expected         string
	}{
		// PostgreSQL
		{"postgres_lowercase", "postgres", "id", ` RETURNING "id"`},
		{"postgres_uppercase", "POSTGRES", "id", ` RETURNING "id"`},
		{"postgres_mixed_case", "Postgres", "id", ` RETURNING "id"`},
		{"postgres_custom_column", "postgres", "user_id", ` RETURNING "user_id"`},

		// SQLite
		{"sqlite3_lowercase", "sqlite3", "id", ` RETURNING "id"`},
		{"sqlite3_uppercase", "SQLITE3", "id", ` RETURNING "id"`},
		{"sqlite3_custom_column", "sqlite3", "record_id", ` RETURNING "record_id"`},

		// SQL Server
		{"sqlserver_lowercase", "sqlserver", "id", ` OUTPUT INSERTED.[id]`},
		{"sqlserver_uppercase", "SQLSERVER", "id", ` OUTPUT INSERTED.[id]`},
		{"sqlserver_custom_column", "sqlserver", "user_id", ` OUTPUT INSERTED.[user_id]`},

		// MSSQL (alias for SQL Server)
		{"mssql_lowercase", "mssql", "id", ` OUTPUT INSERTED.[id]`},
		{"mssql_uppercase", "MSSQL", "id", ` OUTPUT INSERTED.[id]`},
		{"mssql_custom_column", "mssql", "record_id", ` OUTPUT INSERTED.[record_id]`},

		// Oracle
		{"oracle_lowercase", "oracle", "id", ` RETURNING "ID"`},
		{"oracle_uppercase", "ORACLE", "id", ` RETURNING "ID"`},
		{"oracle_custom_column", "oracle", "user_id", ` RETURNING "USER_ID"`},
		{"oracle_mixed_case", "oracle", "UserId", ` RETURNING "USERID"`},

		// MySQL (no RETURNING)
		{"mysql_lowercase", "mysql", "id", ""},
		{"mysql_uppercase", "MYSQL", "id", ""},
		{"mysql_custom_column", "mysql", "user_id", ""},

		// Default/Unknown (defaults to PostgreSQL style)
		{"unknown_driver", "unknown", "id", ` RETURNING "id"`},
		{"empty_driver", "", "id", ` RETURNING "id"`},
		{"custom_driver", "customdb", "id", ` RETURNING "id"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildReturningClause(tt.driverName, tt.primaryKeyColumn)
			if result != tt.expected {
				t.Errorf("buildReturningClause(%q, %q) = %q, want %q", tt.driverName, tt.primaryKeyColumn, result, tt.expected)
			}
		})
	}
}

func TestQuoteIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		driverName string
		identifier string
		expected   string
	}{
		// PostgreSQL
		{"postgres_lowercase", "postgres", "id", `"id"`},
		{"postgres_uppercase", "POSTGRES", "id", `"id"`},
		{"postgres_mixed_case", "Postgres", "id", `"id"`},
		{"postgres_simple_column", "postgres", "name", `"name"`},
		{"postgres_with_underscore", "postgres", "user_id", `"user_id"`},
		{"postgres_mixed_case_column", "postgres", "UserId", `"UserId"`},
		{"postgres_table_name", "postgres", "users", `"users"`},

		// SQLite
		{"sqlite3_lowercase", "sqlite3", "id", `"id"`},
		{"sqlite3_uppercase", "SQLITE3", "id", `"id"`},
		{"sqlite3_mixed_case", "Sqlite3", "id", `"id"`},
		{"sqlite3_simple_column", "sqlite3", "name", `"name"`},
		{"sqlite3_with_underscore", "sqlite3", "record_id", `"record_id"`},
		{"sqlite3_mixed_case_column", "sqlite3", "RecordId", `"RecordId"`},
		{"sqlite3_table_name", "sqlite3", "users", `"users"`},

		// MySQL
		{"mysql_lowercase", "mysql", "id", "`id`"},
		{"mysql_uppercase", "MYSQL", "id", "`id`"},
		{"mysql_mixed_case", "MySql", "id", "`id`"},
		{"mysql_simple_column", "mysql", "name", "`name`"},
		{"mysql_with_underscore", "mysql", "user_id", "`user_id`"},
		{"mysql_mixed_case_column", "mysql", "UserId", "`UserId`"},
		{"mysql_table_name", "mysql", "users", "`users`"},
		{"mysql_reserved_word", "mysql", "order", "`order`"},

		// SQL Server
		{"sqlserver_lowercase", "sqlserver", "id", "[id]"},
		{"sqlserver_uppercase", "SQLSERVER", "id", "[id]"},
		{"sqlserver_mixed_case", "SqlServer", "id", "[id]"},
		{"sqlserver_simple_column", "sqlserver", "name", "[name]"},
		{"sqlserver_with_underscore", "sqlserver", "user_id", "[user_id]"},
		{"sqlserver_mixed_case_column", "sqlserver", "UserId", "[UserId]"},
		{"sqlserver_table_name", "sqlserver", "users", "[users]"},
		{"sqlserver_reserved_word", "sqlserver", "order", "[order]"},

		// MSSQL (alias for SQL Server)
		{"mssql_lowercase", "mssql", "id", "[id]"},
		{"mssql_uppercase", "MSSQL", "id", "[id]"},
		{"mssql_mixed_case", "MsSql", "id", "[id]"},
		{"mssql_simple_column", "mssql", "name", "[name]"},
		{"mssql_with_underscore", "mssql", "record_id", "[record_id]"},
		{"mssql_mixed_case_column", "mssql", "RecordId", "[RecordId]"},
		{"mssql_table_name", "mssql", "users", "[users]"},

		// Oracle
		{"oracle_lowercase", "oracle", "id", `"ID"`},
		{"oracle_uppercase", "ORACLE", "id", `"ID"`},
		{"oracle_mixed_case", "Oracle", "id", `"ID"`},
		{"oracle_simple_column", "oracle", "name", `"NAME"`},
		{"oracle_with_underscore", "oracle", "user_id", `"USER_ID"`},
		{"oracle_mixed_case_column", "oracle", "UserId", `"USERID"`},
		{"oracle_already_uppercase", "oracle", "ID", `"ID"`},
		{"oracle_table_name", "oracle", "users", `"USERS"`},
		{"oracle_complex_name", "oracle", "user_profile_id", `"USER_PROFILE_ID"`},

		// Default/Unknown (defaults to PostgreSQL style)
		{"unknown_driver", "unknown", "id", `"id"`},
		{"empty_driver", "", "id", `"id"`},
		{"custom_driver", "customdb", "id", `"id"`},
		{"custom_driver_column", "customdb", "user_id", `"user_id"`},
		{"custom_driver_mixed_case", "customdb", "UserId", `"UserId"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := quoteIdentifier(tt.driverName, tt.identifier)
			if result != tt.expected {
				t.Errorf("quoteIdentifier(%q, %q) = %q, want %q", tt.driverName, tt.identifier, result, tt.expected)
			}
		})
	}
}

func TestGeneratePlaceholder(t *testing.T) {
	tests := []struct {
		name       string
		driverName string
		position   int
		expected   string
	}{
		// PostgreSQL
		{"postgres_lowercase_1", "postgres", 1, "$1"},
		{"postgres_lowercase_2", "postgres", 2, "$2"},
		{"postgres_lowercase_10", "postgres", 10, "$10"},
		{"postgres_uppercase", "POSTGRES", 1, "$1"},
		{"postgres_mixed_case", "Postgres", 3, "$3"},

		// MySQL
		{"mysql_lowercase_1", "mysql", 1, "?"},
		{"mysql_lowercase_2", "mysql", 2, "?"},
		{"mysql_lowercase_10", "mysql", 10, "?"},
		{"mysql_uppercase", "MYSQL", 1, "?"},
		{"mysql_mixed_case", "MySql", 5, "?"},

		// SQLite
		{"sqlite3_lowercase_1", "sqlite3", 1, "?"},
		{"sqlite3_lowercase_2", "sqlite3", 2, "?"},
		{"sqlite3_lowercase_10", "sqlite3", 10, "?"},
		{"sqlite3_uppercase", "SQLITE3", 1, "?"},
		{"sqlite3_mixed_case", "Sqlite3", 7, "?"},

		// SQL Server
		{"sqlserver_lowercase_1", "sqlserver", 1, "@p1"},
		{"sqlserver_lowercase_2", "sqlserver", 2, "@p2"},
		{"sqlserver_lowercase_10", "sqlserver", 10, "@p10"},
		{"sqlserver_uppercase", "SQLSERVER", 1, "@p1"},
		{"sqlserver_mixed_case", "SqlServer", 3, "@p3"},

		// MSSQL (alias for SQL Server)
		{"mssql_lowercase_1", "mssql", 1, "@p1"},
		{"mssql_lowercase_2", "mssql", 2, "@p2"},
		{"mssql_lowercase_10", "mssql", 10, "@p10"},
		{"mssql_uppercase", "MSSQL", 1, "@p1"},
		{"mssql_mixed_case", "MsSql", 4, "@p4"},

		// Oracle
		{"oracle_lowercase_1", "oracle", 1, ":1"},
		{"oracle_lowercase_2", "oracle", 2, ":2"},
		{"oracle_lowercase_10", "oracle", 10, ":10"},
		{"oracle_uppercase", "ORACLE", 1, ":1"},
		{"oracle_mixed_case", "Oracle", 6, ":6"},

		// Default/Unknown (defaults to PostgreSQL style)
		{"unknown_driver_1", "unknown", 1, "$1"},
		{"unknown_driver_2", "unknown", 2, "$2"},
		{"empty_driver_1", "", 1, "$1"},
		{"empty_driver_5", "", 5, "$5"},
		{"custom_driver_1", "customdb", 1, "$1"},
		{"custom_driver_10", "customdb", 10, "$10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generatePlaceholder(tt.driverName, tt.position)
			if result != tt.expected {
				t.Errorf("generatePlaceholder(%q, %d) = %q, want %q", tt.driverName, tt.position, result, tt.expected)
			}
		})
	}
}

// Test models for checkDotNotationRecursive

// SimpleStructNoDot has no dot notation
type SimpleStructNoDot struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

// SimpleStructWithDot has dot notation in direct field
type SimpleStructWithDot struct {
	UserID int `db:"users.id"`
}

// EmbeddedStructNoDot is embedded without dot notation
type EmbeddedStructNoDot struct {
	ID int `db:"id"`
}

// EmbeddedStructWithDot is embedded with dot notation
type EmbeddedStructWithDot struct {
	UserID int `db:"users.id"`
}

// PointerEmbeddedStructWithDot is a pointer embedded struct with dot notation
type PointerEmbeddedStructWithDot struct {
	ProfileID int `db:"profiles.id"`
}

// Level1Struct has embedded struct (level 1)
type Level1Struct struct {
	EmbeddedStructNoDot
	Field string `db:"field"`
}

// Level2Struct has embedded struct that embeds another struct (level 2)
type Level2Struct struct {
	Level1Struct
	Extra string `db:"extra"`
}

// Level3Struct has 3 levels of embedding
type Level3Struct struct {
	Level2Struct
	Final string `db:"final"`
}

// StructWithEmbeddedDot has embedded struct with dot notation
type StructWithEmbeddedDot struct {
	EmbeddedStructWithDot
	Field string `db:"field"`
}

// StructWithPointerEmbeddedDot has pointer embedded struct with dot notation
type StructWithPointerEmbeddedDot struct {
	*PointerEmbeddedStructWithDot
	Field string `db:"field"`
}

// StructWithDeepEmbeddedDot has 2-level embedding where second level has dot notation
type StructWithDeepEmbeddedDot struct {
	Level1Struct
	DeepField string `db:"deep.field"`
}

// StructWithUnexportedField has unexported field with dot notation (should be skipped)
type StructWithUnexportedField struct {
	ID   int    `db:"id"`
	name string `db:"users.name"` //nolint:unused // unexported, should be skipped - intentionally unused for testing
}

// StructWithDashTag has db:"-" tag (should be skipped)
type StructWithDashTag struct {
	ID   int    `db:"id"`
	Name string `db:"-"`
	User int    `db:"users.id"`
}

func TestCheckDotNotationRecursive(t *testing.T) {
	tests := []struct {
		name     string
		typ      reflect.Type
		expected bool
	}{
		// Non-struct types
		{"int_type", reflect.TypeOf(0), false},
		{"string_type", reflect.TypeOf(""), false},
		{"slice_type", reflect.TypeOf([]int{}), false},
		{"map_type", reflect.TypeOf(map[string]int{}), false},

		// Structs with no dot notation
		{"simple_no_dot", reflect.TypeOf(SimpleStructNoDot{}), false},
		{"insert_model", reflect.TypeOf(InsertModel{}), false},

		// Structs with dot notation in direct fields
		{"simple_with_dot", reflect.TypeOf(SimpleStructWithDot{}), true},
		{"joined_model", reflect.TypeOf(JoinedModel{}), true},

		// Embedded structs without dot notation
		{"embedded_no_dot", reflect.TypeOf(EmbeddedStructNoDot{}), false},
		{"level1_no_dot", reflect.TypeOf(Level1Struct{}), false},
		{"level2_no_dot", reflect.TypeOf(Level2Struct{}), false},
		{"level3_no_dot", reflect.TypeOf(Level3Struct{}), false},

		// Embedded structs with dot notation (non-pointer)
		{"embedded_with_dot", reflect.TypeOf(EmbeddedStructWithDot{}), true},
		{"struct_with_embedded_dot", reflect.TypeOf(StructWithEmbeddedDot{}), true},

		// Pointer embedded structs with dot notation
		{"pointer_embedded_with_dot", reflect.TypeOf(PointerEmbeddedStructWithDot{}), true},
		{"struct_with_pointer_embedded_dot", reflect.TypeOf(StructWithPointerEmbeddedDot{}), true},

		// Deep embedding with dot notation (2-3 levels)
		{"deep_embedded_dot", reflect.TypeOf(StructWithDeepEmbeddedDot{}), true},

		// Unexported fields (should be skipped, so no dot notation detected)
		{"unexported_field", reflect.TypeOf(StructWithUnexportedField{}), false},

		// Dash tag (should be skipped)
		{"dash_tag", reflect.TypeOf(StructWithDashTag{}), true}, // Has users.id, so should return true
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkDotNotationRecursive(tt.typ)
			if result != tt.expected {
				t.Errorf("checkDotNotationRecursive(%v) = %v, want %v", tt.typ, result, tt.expected)
			}
		})
	}
}

// Test models for serializeModelFields

// SerializeTestModel is a simple model for testing serializeModelFields
type SerializeTestModel struct {
	Model
	ID    int64  `db:"id" load:"primary"`
	Name  string `db:"name"`
	Email string `db:"email"`
	Age   int    `db:"age"`
}

// SerializeModelWithDashTag has db:"-" tag
type SerializeModelWithDashTag struct {
	Model
	ID   int    `db:"id" load:"primary"`
	Name string `db:"name"`
	Skip string `db:"-"`
	Age  int    `db:"age"`
}

// SerializeModelWithEmptyTag has empty db tag
type SerializeModelWithEmptyTag struct {
	Model
	ID   int    `db:"id" load:"primary"`
	Name string `db:"name"`
	Skip string // No db tag
	Age  int    `db:"age"`
}

// SerializeModelWithDotNotation has dot notation in db tag
type SerializeModelWithDotNotation struct {
	Model
	ID     int    `db:"id" load:"primary"`
	UserID int    `db:"users.id"`
	Bio    string `db:"profiles.bio"`
}

// SerializeEmbeddedStruct is embedded in other models
type SerializeEmbeddedStruct struct {
	EmbeddedField string `db:"embedded_field"`
}

// SerializeModelWithEmbedded has non-pointer embedded struct
type SerializeModelWithEmbedded struct {
	Model
	SerializeEmbeddedStruct
	ID   int    `db:"id" load:"primary"`
	Name string `db:"name"`
}

// SerializeModelWithPointerEmbedded has pointer embedded struct
type SerializeModelWithPointerEmbedded struct {
	Model
	*SerializeEmbeddedStruct
	ID   int    `db:"id" load:"primary"`
	Name string `db:"name"`
}

// SerializeModelWithUnexported has unexported field
type SerializeModelWithUnexported struct {
	Model
	ID   int    `db:"id" load:"primary"`
	Name string `db:"name"`
	age  int    `db:"age"` //nolint:unused // unexported - intentionally unused for testing
}
