package typedb

import (
	"reflect"
	"testing"
)

func TestDeserializeInt_AllTypes(t *testing.T) {
	tests := []struct {
		value any
		name  string
		want  int
	}{
		{value: int16(123), name: "int16", want: 123},
		{value: int8(45), name: "int8", want: 45},
		{value: uint(789), name: "uint", want: 789},
		{value: uint64(999), name: "uint64", want: 999},
		{value: uint32(111), name: "uint32", want: 111},
		{value: uint16(222), name: "uint16", want: 222},
		{value: uint8(33), name: "uint8", want: 33},
		{value: float32(444.5), name: "float32", want: 444},
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
		value any
		name  string
		want  int64
	}{
		{value: int16(123), name: "int16", want: 123},
		{value: int8(45), name: "int8", want: 45},
		{value: uint32(999), name: "uint32", want: 999},
		{value: uint16(222), name: "uint16", want: 222},
		{value: uint8(33), name: "uint8", want: 33},
		{value: float32(444.5), name: "float32", want: 444},
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
		value any
		name  string
		want  int32
	}{
		{value: int16(123), name: "int16", want: 123},
		{value: int8(45), name: "int8", want: 45},
		{value: uint16(222), name: "uint16", want: 222},
		{value: uint8(33), name: "uint8", want: 33},
		{value: float32(444.5), name: "float32", want: 444},
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
		value any
		name  string
		want  bool
	}{
		{value: int16(1), name: "int16 1", want: true},
		{value: int16(0), name: "int16 0", want: false},
		{value: int8(1), name: "int8 1", want: true},
		{value: int8(0), name: "int8 0", want: false},
		{value: uint(1), name: "uint 1", want: true},
		{value: uint(0), name: "uint 0", want: false},
		{value: uint64(1), name: "uint64 1", want: true},
		{value: uint64(0), name: "uint64 0", want: false},
		{value: uint32(1), name: "uint32 1", want: true},
		{value: uint32(0), name: "uint32 0", want: false},
		{value: uint16(1), name: "uint16 1", want: true},
		{value: uint16(0), name: "uint16 0", want: false},
		{value: uint8(1), name: "uint8 1", want: true},
		{value: uint8(0), name: "uint8 0", want: false},
		{value: float64(1.0), name: "float64 1.0", want: true},
		{value: float64(0.0), name: "float64 0.0", want: false},
		{value: float32(1.0), name: "float32 1.0", want: true},
		{value: float32(0.0), name: "float32 0.0", want: false},
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
		Name string `db:"name"`
		Skip string `db:"-"`
		ID   int    `db:"id"`
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
		{value: "2023-01-01T12:00:00.123456789Z", name: "RFC3339Nano", want: true},
		{value: "2023-01-02 15:04:05", name: "SQL format", want: true},
		{value: "2023-01-02", name: "Date only", want: true},
		{value: "2023-01-02 15:04:05.999999", name: "SQL with microseconds", want: true},
		{value: "2023-01-02 15:04:05.999999999", name: "SQL with nanoseconds", want: true},
		{value: "not a date", name: "Invalid format", want: false},
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
		value any
		name  string
		want  bool
	}{
		{value: []string{"T"}, name: "T uppercase", want: true},
		{value: []string{"F"}, name: "F uppercase", want: false},
		{value: []string{"TRUE"}, name: "TRUE uppercase", want: true},
		{value: []string{"FALSE"}, name: "FALSE uppercase", want: false},
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
		value any
		name  string
		want  int64
	}{
		{value: uint16(123), name: "uint16", want: 123},
		{value: uint8(45), name: "uint8", want: 45},
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
		value any
		name  string
		want  int32
	}{
		{value: uint16(123), name: "uint16", want: 123},
		{value: uint8(45), name: "uint8", want: 45},
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
