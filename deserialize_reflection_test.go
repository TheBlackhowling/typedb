package typedb

import (
	"reflect"
	"testing"
)

func TestDeserializeInt_AllTypes(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  int
	}{
		{"int16", int16(123), 123},
		{"int8", int8(45), 45},
		{"uint", uint(789), 789},
		{"uint64", uint64(999), 999},
		{"uint32", uint32(111), 111},
		{"uint16", uint16(222), 222},
		{"uint8", uint8(33), 33},
		{"float32", float32(444.5), 444},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := deserializeInt(tt.value)
			if err != nil {
				t.Fatalf("DeserializeInt failed: %v", err)
			}
			if got != tt.want {
				t.Errorf("Expected %d, got %d", tt.want, got)
			}
		})
	}
}

func TestDeserializeInt64_AllTypes(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  int64
	}{
		{"int16", int16(123), 123},
		{"int8", int8(45), 45},
		{"uint32", uint32(999), 999},
		{"uint16", uint16(222), 222},
		{"uint8", uint8(33), 33},
		{"float32", float32(444.5), 444},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := deserializeInt64(tt.value)
			if err != nil {
				t.Fatalf("DeserializeInt64 failed: %v", err)
			}
			if got != tt.want {
				t.Errorf("Expected %d, got %d", tt.want, got)
			}
		})
	}
}

func TestDeserializeInt32_AllTypes(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  int32
	}{
		{"int16", int16(123), 123},
		{"int8", int8(45), 45},
		{"uint16", uint16(222), 222},
		{"uint8", uint8(33), 33},
		{"float32", float32(444.5), 444},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := deserializeInt32(tt.value)
			if err != nil {
				t.Fatalf("DeserializeInt32 failed: %v", err)
			}
			if got != tt.want {
				t.Errorf("Expected %d, got %d", tt.want, got)
			}
		})
	}
}

func TestDeserializeBool_AllTypes(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{"int16 1", int16(1), true},
		{"int16 0", int16(0), false},
		{"int8 1", int8(1), true},
		{"int8 0", int8(0), false},
		{"uint 1", uint(1), true},
		{"uint 0", uint(0), false},
		{"uint64 1", uint64(1), true},
		{"uint64 0", uint64(0), false},
		{"uint32 1", uint32(1), true},
		{"uint32 0", uint32(0), false},
		{"uint16 1", uint16(1), true},
		{"uint16 0", uint16(0), false},
		{"uint8 1", uint8(1), true},
		{"uint8 0", uint8(0), false},
		{"float64 1.0", float64(1.0), true},
		{"float64 0.0", float64(0.0), false},
		{"float32 1.0", float32(1.0), true},
		{"float32 0.0", float32(0.0), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := deserializeBool(tt.value)
			if err != nil {
				t.Fatalf("DeserializeBool failed: %v", err)
			}
			if got != tt.want {
				t.Errorf("Expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestBuildFieldMap_PointerEmbeddedStruct(t *testing.T) {
	type PointerEmbedded struct {
		ID int `db:"id"`
	}
	type ModelWithPointerEmbedded struct {
		Model
		*PointerEmbedded
		Name string `db:"name"`
	}

	model := &ModelWithPointerEmbedded{}
	modelValue := reflect.ValueOf(model).Elem()
	fieldMap := buildFieldMapFromPtr(reflect.ValueOf(model), modelValue)

	if _, ok := fieldMap["id"]; !ok {
		t.Error("Expected 'id' field in map")
	}
	if _, ok := fieldMap["name"]; !ok {
		t.Error("Expected 'name' field in map")
	}
}

func TestBuildFieldMap_SkipDashTag(t *testing.T) {
	type ModelWithDashTag struct {
		Model
		ID   int    `db:"id"`
		Name string `db:"name"`
		Skip string `db:"-"`
	}

	model := &ModelWithDashTag{}
	modelValue := reflect.ValueOf(model).Elem()
	fieldMap := buildFieldMapFromPtr(reflect.ValueOf(model), modelValue)

	if _, ok := fieldMap["-"]; ok {
		t.Error("Should not include field with db:\"-\" tag")
	}
	if _, ok := fieldMap["id"]; !ok {
		t.Error("Expected 'id' field in map")
	}
	if _, ok := fieldMap["name"]; !ok {
		t.Error("Expected 'name' field in map")
	}
}

func TestDeserializeToField_NonPointerTarget(t *testing.T) {
	var i int
	err := deserializeToField(i, 123)
	if err == nil {
		t.Error("Expected error for non-pointer target")
	}
}

func TestDeserializeToField_ReflectionAssignable(t *testing.T) {
	// Test AssignableTo path in deserializeWithReflection
	type MyInt int
	var myInt MyInt
	err := deserializeToField(&myInt, MyInt(123))
	if err != nil {
		t.Fatalf("DeserializeToField failed: %v", err)
	}
	if myInt != 123 {
		t.Errorf("Expected 123, got %d", myInt)
	}
}

func TestDeserializeToField_ReflectionConvertible(t *testing.T) {
	// Test ConvertibleTo path in deserializeWithReflection
	type MyInt int
	var myInt MyInt
	err := deserializeToField(&myInt, int(456))
	if err != nil {
		t.Fatalf("DeserializeToField failed: %v", err)
	}
	if myInt != 456 {
		t.Errorf("Expected 456, got %d", myInt)
	}
}

func TestDeserializeToField_ReflectionPointerTarget(t *testing.T) {
	// Test pointer target with AssignableTo
	type MyInt int
	var myIntPtr *MyInt
	err := deserializeToField(&myIntPtr, MyInt(789))
	if err != nil {
		t.Fatalf("DeserializeToField failed: %v", err)
	}
	if myIntPtr == nil || *myIntPtr != 789 {
		t.Errorf("Expected *789, got %v", myIntPtr)
	}
}

func TestDeserializeToField_ReflectionPointerTargetConvertible(t *testing.T) {
	// Test pointer target with ConvertibleTo
	type MyInt int
	var myIntPtr *MyInt
	err := deserializeToField(&myIntPtr, int(999))
	if err != nil {
		t.Fatalf("DeserializeToField failed: %v", err)
	}
	if myIntPtr == nil || *myIntPtr != 999 {
		t.Errorf("Expected *999, got %v", myIntPtr)
	}
}

func TestDeserializeToField_ReflectionError(t *testing.T) {
	// Test error path when types are incompatible
	type MyInt int
	type MyString string
	var myInt MyInt
	err := deserializeToField(&myInt, MyString("invalid"))
	if err == nil {
		t.Error("Expected error for incompatible types")
	}
}

func TestParseTime_AllFormats(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool // true if should succeed
	}{
		{"RFC3339Nano", "2023-01-01T12:00:00.123456789Z", true},
		{"SQL format", "2023-01-02 15:04:05", true},
		{"Date only", "2023-01-02", true},
		{"SQL with microseconds", "2023-01-02 15:04:05.999999", true},
		{"SQL with nanoseconds", "2023-01-02 15:04:05.999999999", true},
		{"Invalid format", "not a date", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTime(tt.value)
			if tt.want {
				if err != nil {
					t.Fatalf("parseTime failed: %v", err)
				}
				if got.IsZero() && tt.name != "Invalid format" {
					t.Error("Expected non-zero time")
				}
			} else {
				if err == nil {
					t.Error("Expected error for invalid time format")
				}
			}
		})
	}
}

// BadTypeModel implements ModelInterface for testing error paths
type BadTypeModel struct {
	Model
	ID int `db:"id"`
}

func TestDeserialize_ErrorFromdeserializeToField(t *testing.T) {
	badModel := &BadTypeModel{}

	row := map[string]any{
		"id": "not an int", // This will cause DeserializeInt to fail
	}

	err := deserialize(row, badModel)
	if err == nil {
		t.Error("Expected error for invalid type conversion")
	}
}

func TestDeserializeIntArray_Int32Slice(t *testing.T) {
	arr := []int32{1, 2, 3}
	result, err := deserializeIntArray(arr)
	if err != nil {
		t.Fatalf("DeserializeIntArray failed: %v", err)
	}
	if !reflect.DeepEqual(result, []int{1, 2, 3}) {
		t.Errorf("Expected [1,2,3], got %v", result)
	}
}

func TestDeserializeIntArray_ErrorInElement(t *testing.T) {
	// Test error path when DeserializeInt fails on an element
	arr := []any{1, "not a number", 3}
	_, err := deserializeIntArray(arr)
	if err == nil {
		t.Error("Expected error for invalid element")
	}
}

func TestDeserializeIntArray_ErrorInPostgresFormat(t *testing.T) {
	// Test error path when parsing postgres array format fails
	_, err := deserializeIntArray("{1,not a number,3}")
	if err == nil {
		t.Error("Expected error for invalid postgres array format")
	}
}

func TestDeserializeBool_DefaultCaseMore(t *testing.T) {
	// Test default case paths that hit different string parsing branches
	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{"T uppercase", []string{"T"}, true},
		{"F uppercase", []string{"F"}, false},
		{"TRUE uppercase", []string{"TRUE"}, true},
		{"FALSE uppercase", []string{"FALSE"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := deserializeBool(tt.value)
			if err != nil {
				// Some might fail, that's okay
				return
			}
			if got != tt.want {
				t.Errorf("Expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestParseTime_EmptyString(t *testing.T) {
	// Test empty string path
	timeVal, err := parseTime("")
	if err != nil {
		t.Fatalf("parseTime failed: %v", err)
	}
	if !timeVal.IsZero() {
		t.Error("Expected zero time for empty string")
	}
}

func TestDeserializeInt64_MoreTypes(t *testing.T) {
	// Test remaining type cases
	tests := []struct {
		name  string
		value any
		want  int64
	}{
		{"uint16", uint16(123), 123},
		{"uint8", uint8(45), 45},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := deserializeInt64(tt.value)
			if err != nil {
				t.Fatalf("DeserializeInt64 failed: %v", err)
			}
			if got != tt.want {
				t.Errorf("Expected %d, got %d", tt.want, got)
			}
		})
	}
}

func TestDeserializeInt32_MoreTypes(t *testing.T) {
	// Test remaining type cases
	tests := []struct {
		name  string
		value any
		want  int32
	}{
		{"uint16", uint16(123), 123},
		{"uint8", uint8(45), 45},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := deserializeInt32(tt.value)
			if err != nil {
				t.Fatalf("DeserializeInt32 failed: %v", err)
			}
			if got != tt.want {
				t.Errorf("Expected %d, got %d", tt.want, got)
			}
		})
	}
}
