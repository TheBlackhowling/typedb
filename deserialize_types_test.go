package typedb

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestDeserialize_TimeFormats(t *testing.T) {
	user := &DeserializeUser{}

	tests := []struct {
		name  string
		value string
	}{
		{name: "RFC3339", value: "2023-01-01T12:00:00Z"},
		{name: "RFC3339Nano", value: "2023-01-01T12:00:00.123456789Z"},
		{name: "SQL format", value: "2023-01-02 15:04:05"},
		{name: "Date only", value: "2023-01-02"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := map[string]any{"created_at": tt.value}
			err := deserialize(row, user)
			if err != nil {
				t.Fatalf("Deserialize failed: %v", err)
			}
			if user.CreatedAt.IsZero() {
				t.Error("Expected CreatedAt to be set")
			}
		})
	}
}

func TestDeserializeInt(t *testing.T) {
	tests := []struct {
		value any
		name  string
		want  int
	}{
		{value: 123, name: "int", want: 123},
		{value: int64(456), name: "int64", want: 456},
		{value: int32(789), name: "int32", want: 789},
		{value: float64(999), name: "float64", want: 999},
		{value: "111", name: "string", want: 111},
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

func TestDeserializeBool(t *testing.T) {
	tests := []struct {
		value any
		name  string
		want  bool
	}{
		{value: true, name: "bool true", want: true},
		{value: false, name: "bool false", want: false},
		{value: "true", name: "string true", want: true},
		{value: "false", name: "string false", want: false},
		{value: "1", name: "string 1", want: true},
		{value: "0", name: "string 0", want: false},
		{value: 1, name: "int 1", want: true},
		{value: 0, name: "int 0", want: false},
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

func TestDeserializeString(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{"string", "test", "test"},
		{"int", 123, "123"},
		{"bool", true, "true"},
		{"nil", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deserializeString(tt.value)
			if got != tt.want {
				t.Errorf("Expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestDeserializeJSONB(t *testing.T) {
	tests := []struct {
		value any
		want  map[string]any
		name  string
	}{
		{value: map[string]any{"key": "value"}, want: map[string]any{"key": "value"}, name: "map"},
		{value: `{"key": "value"}`, want: map[string]any{"key": "value"}, name: "json string"},
		{value: []byte(`{"key": "value"}`), want: map[string]any{"key": "value"}, name: "json bytes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := deserializeJSONB(tt.value)
			if err != nil {
				t.Fatalf("DeserializeJSONB failed: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				gotJSON, _ := json.Marshal(got)
				wantJSON, _ := json.Marshal(tt.want)
				t.Errorf("Expected %s, got %s", wantJSON, gotJSON)
			}
		})
	}
}

func TestDeserializeJSONB_InvalidJSON(t *testing.T) {
	_, err := deserializeJSONB("invalid json")
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestDeserializeJSONB_UnsupportedType(t *testing.T) {
	_, err := deserializeJSONB(123)
	if err == nil {
		t.Error("Expected error for unsupported type")
	}
}

func TestDeserializeIntArray_InvalidFormat(t *testing.T) {
	_, err := deserializeIntArray("invalid")
	if err == nil {
		t.Error("Expected error for invalid array format")
	}
}

func TestDeserializeStringArray_InvalidFormat(t *testing.T) {
	_, err := deserializeStringArray(123)
	if err == nil {
		t.Error("Expected error for invalid array format")
	}
}

func TestDeserializeMap(t *testing.T) {
	tests := []struct {
		value any
		want  map[string]string
		name  string
	}{
		{value: map[string]string{"key": "value"}, want: map[string]string{"key": "value"}, name: "map[string]string"},
		{value: map[string]any{"key": 123}, want: map[string]string{"key": "123"}, name: "map[string]any"},
		{value: `{"key": "value"}`, want: map[string]string{"key": "value"}, name: "json string"},
		{value: []byte(`{"key": "value"}`), want: map[string]string{"key": "value"}, name: "json bytes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := deserializeMap(tt.value)
			if err != nil {
				t.Fatalf("DeserializeMap failed: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestDeserializeMap_UnsupportedType(t *testing.T) {
	_, err := deserializeMap(123)
	if err == nil {
		t.Error("Expected error for unsupported type")
	}
}

func TestDeserializeInt64(t *testing.T) {
	tests := []struct {
		value any
		name  string
		want  int64
	}{
		{value: int64(123), name: "int64", want: 123},
		{value: int(456), name: "int", want: 456},
		{value: int32(789), name: "int32", want: 789},
		{value: uint64(999), name: "uint64", want: 999},
		{value: float64(111), name: "float64", want: 111},
		{value: "222", name: "string", want: 222},
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

func TestDeserializeInt32(t *testing.T) {
	tests := []struct {
		value any
		name  string
		want  int32
	}{
		{value: int32(123), name: "int32", want: 123},
		{value: int(456), name: "int", want: 456},
		{value: int64(789), name: "int64", want: 789},
		{value: uint32(999), name: "uint32", want: 999},
		{value: float64(111), name: "float64", want: 111},
		{value: "222", name: "string", want: 222},
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

func TestDeserializeTime(t *testing.T) {
	tests := []struct {
		value any
		name  string
		want  bool
	}{
		{value: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), name: "time.Time", want: true},
		{value: "2023-01-01T12:00:00Z", name: "RFC3339 string", want: true},
		{value: []byte("2023-01-01T12:00:00Z"), name: "bytes", want: true},
		{value: 123, name: "invalid", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := deserializeTime(tt.value)
			if tt.want {
				if err != nil {
					t.Fatalf("DeserializeTime failed: %v", err)
				}
				if got.IsZero() && tt.name != "invalid" {
					t.Error("Expected non-zero time")
				}
			} else {
				if err == nil {
					t.Error("Expected error for invalid time")
				}
			}
		})
	}
}

func TestDeserializeToField_Reflection(t *testing.T) {
	// Test reflection path for types not in the type switch
	type CustomInt int
	var customInt CustomInt

	err := deserializeToField(&customInt, 123)
	if err != nil {
		t.Fatalf("DeserializeToField failed: %v", err)
	}
	if customInt != 123 {
		t.Errorf("Expected 123, got %d", customInt)
	}
}

func TestDeserializeToField_PointerType(t *testing.T) {
	var intPtr *int
	err := deserializeToField(&intPtr, 456)
	if err != nil {
		t.Fatalf("DeserializeToField failed: %v", err)
	}
	if intPtr == nil || *intPtr != 456 {
		t.Errorf("Expected *456, got %v", intPtr)
	}
}

// TestDeserialize_PointerToPrimitive tests that *bool, *int, *string deserialize correctly.
// Pointer-to-primitive types allow nil (omit/NULL) vs explicit zero values in Update.
func TestDeserialize_PointerToPrimitive(t *testing.T) {
	type ModelWithPointerPrimitives struct {
		Model
		IsActive *bool   `db:"is_active"`
		Count    *int    `db:"count"`
		Nickname *string `db:"nickname"`
		Name     string  `db:"name"`
		ID       int64   `db:"id"`
	}

	tests := []struct {
		row   map[string]any
		check func(t *testing.T, m *ModelWithPointerPrimitives)
		name  string
	}{
		{
			name: "*bool true",
			row:  map[string]any{"id": int64(1), "name": "Alice", "is_active": true},
			check: func(t *testing.T, m *ModelWithPointerPrimitives) {
				if m.IsActive == nil || *m.IsActive != true {
					t.Errorf("IsActive = %v, want *true", m.IsActive)
				}
			},
		},
		{
			name: "*bool false",
			row:  map[string]any{"id": int64(2), "name": "Bob", "is_active": false},
			check: func(t *testing.T, m *ModelWithPointerPrimitives) {
				if m.IsActive == nil || *m.IsActive != false {
					t.Errorf("IsActive = %v, want *false", m.IsActive)
				}
			},
		},
		{
			name: "*bool nil",
			row:  map[string]any{"id": int64(3), "name": "Carol", "is_active": nil},
			check: func(t *testing.T, m *ModelWithPointerPrimitives) {
				if m.IsActive != nil {
					t.Errorf("IsActive = %v, want nil", m.IsActive)
				}
			},
		},
		{
			name: "*int zero",
			row:  map[string]any{"id": int64(4), "name": "Dave", "count": 0},
			check: func(t *testing.T, m *ModelWithPointerPrimitives) {
				if m.Count == nil || *m.Count != 0 {
					t.Errorf("Count = %v, want *0", m.Count)
				}
			},
		},
		{
			name: "*int nonzero",
			row:  map[string]any{"id": int64(5), "name": "Eve", "count": 42},
			check: func(t *testing.T, m *ModelWithPointerPrimitives) {
				if m.Count == nil || *m.Count != 42 {
					t.Errorf("Count = %v, want *42", m.Count)
				}
			},
		},
		{
			name: "*string nil",
			row:  map[string]any{"id": int64(6), "name": "Frank", "nickname": nil},
			check: func(t *testing.T, m *ModelWithPointerPrimitives) {
				if m.Nickname != nil {
					t.Errorf("Nickname = %v, want nil", m.Nickname)
				}
			},
		},
		{
			name: "*string empty",
			row:  map[string]any{"id": int64(7), "name": "Grace", "nickname": ""},
			check: func(t *testing.T, m *ModelWithPointerPrimitives) {
				if m.Nickname == nil || *m.Nickname != "" {
					t.Errorf("Nickname = %v, want *\"\"", m.Nickname)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ModelWithPointerPrimitives{}
			err := deserialize(tt.row, m)
			if err != nil {
				t.Fatalf("deserialize failed: %v", err)
			}
			tt.check(t, m)
		})
	}
}
