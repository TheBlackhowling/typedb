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
		{"RFC3339", "2023-01-01T12:00:00Z"},
		{"RFC3339Nano", "2023-01-01T12:00:00.123456789Z"},
		{"SQL format", "2023-01-02 15:04:05"},
		{"Date only", "2023-01-02"},
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
		name  string
		value any
		want  int
	}{
		{"int", 123, 123},
		{"int64", int64(456), 456},
		{"int32", int32(789), 789},
		{"float64", float64(999), 999},
		{"string", "111", 111},
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
		name  string
		value any
		want  bool
	}{
		{"bool true", true, true},
		{"bool false", false, false},
		{"string true", "true", true},
		{"string false", "false", false},
		{"string 1", "1", true},
		{"string 0", "0", false},
		{"int 1", 1, true},
		{"int 0", 0, false},
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
		name  string
		value any
		want  map[string]any
	}{
		{"map", map[string]any{"key": "value"}, map[string]any{"key": "value"}},
		{"json string", `{"key": "value"}`, map[string]any{"key": "value"}},
		{"json bytes", []byte(`{"key": "value"}`), map[string]any{"key": "value"}},
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
		name  string
		value any
		want  map[string]string
	}{
		{"map[string]string", map[string]string{"key": "value"}, map[string]string{"key": "value"}},
		{"map[string]any", map[string]any{"key": 123}, map[string]string{"key": "123"}},
		{"json string", `{"key": "value"}`, map[string]string{"key": "value"}},
		{"json bytes", []byte(`{"key": "value"}`), map[string]string{"key": "value"}},
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
		name  string
		value any
		want  int64
	}{
		{"int64", int64(123), 123},
		{"int", int(456), 456},
		{"int32", int32(789), 789},
		{"uint64", uint64(999), 999},
		{"float64", float64(111), 111},
		{"string", "222", 222},
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
		name  string
		value any
		want  int32
	}{
		{"int32", int32(123), 123},
		{"int", int(456), 456},
		{"int64", int64(789), 789},
		{"uint32", uint32(999), 999},
		{"float64", float64(111), 111},
		{"string", "222", 222},
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
		name  string
		value any
		want  bool // true if should succeed
	}{
		{"time.Time", time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), true},
		{"RFC3339 string", "2023-01-01T12:00:00Z", true},
		{"bytes", []byte("2023-01-01T12:00:00Z"), true},
		{"invalid", 123, false},
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
