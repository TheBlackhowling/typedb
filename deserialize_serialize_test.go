package typedb

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestDeserializeToField_NilValue(t *testing.T) {
	var intPtr *int
	err := deserializeToField(&intPtr, nil)
	if err != nil {
		t.Fatalf("DeserializeToField failed: %v", err)
	}
	if intPtr != nil {
		t.Error("Expected nil pointer")
	}

	var i int
	err = deserializeToField(&i, nil)
	if err != nil {
		t.Fatalf("DeserializeToField failed: %v", err)
	}
	if i != 0 {
		t.Errorf("Expected zero value, got %d", i)
	}
}

// BadModel implements ModelInterface for testing error paths
type BadModel struct {
	Model
	ID int `db:"id"`
}

func (b *BadModel) Deserialize(row map[string]any) error {
	return deserialize(row, b)
}

func TestDeserializeForType_ErrorPath(t *testing.T) {
	// Test error path when Deserialize fails
	row := map[string]any{
		"id": "invalid", // This will cause an error in Deserialize
	}

	// BadModel implements ModelInterface via Deserialize method
	// But Deserialize will fail because "invalid" can't be converted to int
	_, err := deserializeForType[*BadModel](row)
	if err == nil {
		t.Error("Expected error for invalid deserialization")
	}
}

// Serialization tests

func TestSerializeJSONB(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{"map[string]any", map[string]any{"key": "value"}, `{"key":"value"}`},
		{"map[string]string", map[string]string{"key": "value"}, `{"key":"value"}`},
		{"already string", `{"key":"value"}`, `{"key":"value"}`},
		{"already bytes", []byte(`{"key":"value"}`), `{"key":"value"}`},
		{"nil", nil, ""},
		{"slice", []string{"a", "b"}, `["a","b"]`},
		{"int slice", []int{1, 2, 3}, `[1,2,3]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := serializeJSONB(tt.value)
			if tt.value == nil {
				if err != nil {
					t.Fatalf("SerializeJSONB failed: %v", err)
				}
				if got != nil {
					t.Errorf("Expected nil, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("SerializeJSONB failed: %v", err)
			}
			gotStr, ok := got.(string)
			if !ok {
				gotBytes, ok := got.([]byte)
				if !ok {
					t.Fatalf("Expected string or []byte, got %T", got)
				}
				gotStr = string(gotBytes)
			}
			// Parse both to compare semantically (order might differ)
			var gotMap, wantMap map[string]any
			if err := json.Unmarshal([]byte(gotStr), &gotMap); err == nil {
				if err := json.Unmarshal([]byte(tt.want), &wantMap); err == nil {
					if !reflect.DeepEqual(gotMap, wantMap) {
						t.Errorf("Expected %s, got %s", tt.want, gotStr)
					}
					return
				}
			}
			// For non-map values, compare strings directly
			if gotStr != tt.want {
				t.Errorf("Expected %s, got %s", tt.want, gotStr)
			}
		})
	}
}

func TestSerializeJSONB_Error(t *testing.T) {
	// Test with a type that can't be marshaled (channel)
	ch := make(chan int)
	_, err := serializeJSONB(ch)
	if err == nil {
		t.Error("Expected error for unmarshalable type")
	}
}

func TestSerializeIntArray(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{"[]int", []int{1, 2, 3}, "{1,2,3}"},
		{"[]int64", []int64{10, 20, 30}, "{10,20,30}"},
		{"[]int32", []int32{100, 200}, "{100,200}"},
		{"[]int16", []int16{5, 6}, "{5,6}"},
		{"[]int8", []int8{1, 2}, "{1,2}"},
		{"[]uint", []uint{7, 8}, "{7,8}"},
		{"[]uint64", []uint64{9, 10}, "{9,10}"},
		{"[]uint32", []uint32{11, 12}, "{11,12}"},
		{"[]uint16", []uint16{13, 14}, "{13,14}"},
		{"[]uint8", []uint8{15, 16}, "{15,16}"},
		{"[]any", []any{1, 2, 3}, "{1,2,3}"},
		{"empty", []int{}, "{}"},
		{"nil", nil, "{}"},
		{"single element", []int{42}, "{42}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := serializeIntArray(tt.value)
			if err != nil {
				t.Fatalf("SerializeIntArray failed: %v", err)
			}
			if got != tt.want {
				t.Errorf("Expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestSerializeIntArray_Error(t *testing.T) {
	_, err := serializeIntArray("not an array")
	if err == nil {
		t.Error("Expected error for non-array type")
	}

	_, err = serializeIntArray([]any{1, "not a number", 3})
	if err == nil {
		t.Error("Expected error for invalid element")
	}
}

func TestSerializeStringArray(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{"simple", []string{"a", "b", "c"}, "{a,b,c}"},
		{"with comma", []string{"a,b", "c"}, `{"a,b",c}`},
		{"with quote", []string{`a"b`, "c"}, `{"a\"b",c}`},
		{"with backslash", []string{`a\b`, "c"}, `{"a\\b",c}`},
		{"with brace", []string{"a{b", "c"}, `{"a{b",c}`},
		{"empty", []string{}, "{}"},
		{"nil", nil, "{}"},
		{"single element", []string{"hello"}, "{hello}"},
		{"[]any", []any{"a", "b"}, "{a,b}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := serializeStringArray(tt.value)
			if err != nil {
				t.Fatalf("serializeStringArray failed: %v", err)
			}
			if got != tt.want {
				t.Errorf("Expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestSerializeStringArray_Error(t *testing.T) {
	_, err := serializeStringArray(123)
	if err == nil {
		t.Error("Expected error for non-array type")
	}
}

func TestSerialize(t *testing.T) {
	tests := []struct {
		value any
		want  any
		name  string
	}{
		{value: 123, want: 123, name: "int"},
		{value: int64(456), want: int64(456), name: "int64"},
		{value: int32(789), want: int32(789), name: "int32"},
		{value: uint(999), want: uint(999), name: "uint"},
		{value: float64(1.5), want: float64(1.5), name: "float64"},
		{value: float32(2.5), want: float32(2.5), name: "float32"},
		{value: "hello", want: "hello", name: "string"},
		{value: true, want: true, name: "bool"},
		{value: []byte("bytes"), want: []byte("bytes"), name: "[]byte"},
		{value: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), want: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), name: "time.Time"},
		{value: map[string]any{"key": "value"}, want: `{"key":"value"}`, name: "map[string]any"},
		{value: map[string]string{"key": "value"}, want: `{"key":"value"}`, name: "map[string]string"},
		{value: []int{1, 2, 3}, want: "{1,2,3}", name: "[]int"},
		{value: []int64{10, 20}, want: "{10,20}", name: "[]int64"},
		{value: []string{"a", "b"}, want: "{a,b}", name: "[]string"},
		{value: nil, want: nil, name: "nil"},
		{value: struct{ Name string }{"test"}, want: `{"Name":"test"}`, name: "custom struct"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := serialize(tt.value)
			if err != nil {
				t.Fatalf("serialize failed: %v", err)
			}
			if tt.value == nil {
				if got != nil {
					t.Errorf("Expected nil, got %v", got)
				}
				return
			}
			// For string results (JSONB/arrays), compare as strings
			if gotStr, ok := got.(string); ok {
				if wantStr, ok := tt.want.(string); ok {
					// For JSON, parse and compare semantically
					if strings.HasPrefix(gotStr, "{") && strings.HasPrefix(wantStr, "{") && !strings.HasPrefix(gotStr, "{1") {
						var gotMap, wantMap map[string]any
						if err := json.Unmarshal([]byte(gotStr), &gotMap); err == nil {
							if err := json.Unmarshal([]byte(wantStr), &wantMap); err == nil {
								if !reflect.DeepEqual(gotMap, wantMap) {
									t.Errorf("Expected %s, got %s", wantStr, gotStr)
								}
								return
							}
						}
					}
					if gotStr != wantStr {
						t.Errorf("Expected %q, got %q", wantStr, gotStr)
					}
					return
				}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Expected %v (%T), got %v (%T)", tt.want, tt.want, got, got)
			}
		})
	}
}

func TestSerializeJSONB_MoreTypes(t *testing.T) {
	// Test types that go through JSON marshaling
	tests := []struct {
		value any
		name  string
	}{
		{"struct", struct{ Name string }{"test"}},
		{"slice of maps", []map[string]any{{"a": 1}, {"b": 2}}},
		{"nested map", map[string]any{"nested": map[string]any{"key": "value"}}},
		{"[]any", []any{"a", 1, true}},
		{"[]string", []string{"a", "b", "c"}},
		{"[]int", []int{1, 2, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := serializeJSONB(tt.value)
			if err != nil {
				t.Fatalf("SerializeJSONB failed: %v", err)
			}
			if got == nil {
				t.Error("Expected non-nil result")
			}
			// Verify it's valid JSON by unmarshaling
			gotStr, ok := got.(string)
			if !ok {
				gotBytes, ok := got.([]byte)
				if !ok {
					t.Fatalf("Expected string or []byte, got %T", got)
				}
				gotStr = string(gotBytes)
			}
			var result any
			if err := json.Unmarshal([]byte(gotStr), &result); err != nil {
				t.Errorf("Result is not valid JSON: %v", err)
			}
		})
	}
}
