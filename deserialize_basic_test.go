package typedb

import (
	"reflect"
	"testing"
)

func TestDeserialize_BasicTypes(t *testing.T) {
	user := &DeserializeUser{}
	row := map[string]any{
		"id":         123,
		"name":       "John Doe",
		"email":      "john@example.com",
		"active":     true,
		"created_at": "2023-01-01T12:00:00Z",
	}

	err := deserialize(row, user)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if user.ID != 123 {
		t.Errorf("Expected ID 123, got %d", user.ID)
	}
	if user.Name != "John Doe" {
		t.Errorf("Expected Name 'John Doe', got %q", user.Name)
	}
	if user.Email != "john@example.com" {
		t.Errorf("Expected Email 'john@example.com', got %q", user.Email)
	}
	if !user.Active {
		t.Error("Expected Active true, got false")
	}
	if user.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
}

func TestDeserialize_IntConversions(t *testing.T) {
	user := &DeserializeUser{}

	tests := []struct {
		value any    // 16 bytes (interface{})
		name  string // 16 bytes
		want  int    // 8 bytes
	}{
		{int64(123), "int64", 123},
		{int32(456), "int32", 456},
		{float64(789), "float64", 789},
		{"999", "string", 999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := map[string]any{"id": tt.value}
			err := deserialize(row, user)
			if err != nil {
				t.Fatalf("Deserialize failed: %v", err)
			}
			if user.ID != tt.want {
				t.Errorf("Expected ID %d, got %d", tt.want, user.ID)
			}
		})
	}
}

func TestDeserialize_BoolConversions(t *testing.T) {
	post := &DeserializePost{}

	tests := []struct {
		value any    // 16 bytes (interface{})
		name  string // 16 bytes
		want  bool   // 1 byte
	}{
		{true, "bool true", true},
		{false, "bool false", false},
		{"true", "string true", true},
		{"false", "string false", false},
		{"1", "string 1", true},
		{"0", "string 0", false},
		{"t", "string t", true},
		{"f", "string f", false},
		{1, "int 1", true},
		{0, "int 0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := map[string]any{"published": tt.value}
			err := deserialize(row, post)
			if err != nil {
				t.Fatalf("Deserialize failed: %v", err)
			}
			if post.Published != tt.want {
				t.Errorf("Expected Published %v, got %v", tt.want, post.Published)
			}
		})
	}
}

func TestDeserialize_PointerFields(t *testing.T) {
	model := &DeserializeModelWithPointers{}

	row := map[string]any{
		"id":      123,
		"name":    "Test",
		"active":  true,
		"deleted": nil,
	}

	err := deserialize(row, model)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if model.ID == nil || *model.ID != 123 {
		t.Errorf("Expected ID *123, got %v", model.ID)
	}
	if model.Name == nil || *model.Name != "Test" {
		t.Errorf("Expected Name *'Test', got %v", model.Name)
	}
	if model.Active == nil || !*model.Active {
		t.Error("Expected Active *true, got false or nil")
	}
	if model.Deleted != nil {
		t.Error("Expected Deleted nil, got non-nil")
	}
}

func TestDeserialize_NilValues(t *testing.T) {
	model := &DeserializeModelWithPointers{}

	row := map[string]any{
		"id":      nil,
		"name":    nil,
		"active":  nil,
		"deleted": nil,
	}

	err := deserialize(row, model)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	// Pointer fields should be set to nil
	if model.ID != nil {
		t.Error("Expected ID nil")
	}
	if model.Name != nil {
		t.Error("Expected Name nil")
	}
	if model.Active != nil {
		t.Error("Expected Active nil")
	}
	if model.Deleted != nil {
		t.Error("Expected Deleted nil")
	}
}

func TestDeserialize_StringArrays(t *testing.T) {
	model := &DeserializeModelWithArrays{}

	tests := []struct {
		name  string
		value any
		want  []string
	}{
		{"slice", []string{"tag1", "tag2"}, []string{"tag1", "tag2"}},
		{"postgres array", "{tag1,tag2,tag3}", []string{"tag1", "tag2", "tag3"}},
		{"empty postgres array", "{}", []string{}},
		{"any slice", []any{"a", "b"}, []string{"a", "b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := map[string]any{"tags": tt.value}
			err := deserialize(row, model)
			if err != nil {
				t.Fatalf("Deserialize failed: %v", err)
			}
			if !reflect.DeepEqual(model.Tags, tt.want) {
				t.Errorf("Expected Tags %v, got %v", tt.want, model.Tags)
			}
		})
	}
}

func TestDeserialize_IntArrays(t *testing.T) {
	model := &DeserializeModelWithArrays{}

	tests := []struct {
		name  string
		value any
		want  []int
	}{
		{"slice", []int{1, 2, 3}, []int{1, 2, 3}},
		{"int64 slice", []int64{10, 20, 30}, []int{10, 20, 30}},
		{"postgres array", "{1,2,3}", []int{1, 2, 3}},
		{"empty postgres array", "{}", []int{}},
		{"any slice", []any{1, 2, 3}, []int{1, 2, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := map[string]any{"numbers": tt.value}
			err := deserialize(row, model)
			if err != nil {
				t.Fatalf("Deserialize failed: %v", err)
			}
			if !reflect.DeepEqual(model.Numbers, tt.want) {
				t.Errorf("Expected Numbers %v, got %v", tt.want, model.Numbers)
			}
		})
	}
}

func TestDeserialize_JSONB(t *testing.T) {
	model := &DeserializeModelWithJSON{}

	metadataJSON := `{"key": "value", "num": 42}`
	row := map[string]any{
		"id":       1,
		"metadata": metadataJSON,
		"config":   `{"setting": "value"}`,
	}

	err := deserialize(row, model)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if model.Metadata == nil {
		t.Fatal("Expected Metadata to be set")
	}
	if model.Metadata["key"] != "value" {
		t.Errorf("Expected Metadata['key'] 'value', got %v", model.Metadata["key"])
	}
	if model.Config == nil {
		t.Fatal("Expected Config to be set")
	}
	if model.Config["setting"] != "value" {
		t.Errorf("Expected Config['setting'] 'value', got %v", model.Config["setting"])
	}
}

func TestDeserialize_DotNotation(t *testing.T) {
	model := &DeserializeModelWithDotNotation{}

	row := map[string]any{
		"users.id":   123,
		"users.name": "John",
	}

	err := deserialize(row, model)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if model.ID != 123 {
		t.Errorf("Expected ID 123, got %d", model.ID)
	}
	if model.Name != "John" {
		t.Errorf("Expected Name 'John', got %q", model.Name)
	}
}

func TestDeserialize_EmbeddedStructs(t *testing.T) {
	model := &DerivedModel{}

	row := map[string]any{
		"id":   456,
		"name": "Derived",
	}

	err := deserialize(row, model)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if model.ID != 456 {
		t.Errorf("Expected ID 456, got %d", model.ID)
	}
	if model.Name != "Derived" {
		t.Errorf("Expected Name 'Derived', got %q", model.Name)
	}
}

func TestDeserializeForType(t *testing.T) {
	row := map[string]any{
		"id":         789,
		"name":       "Generic User",
		"email":      "generic@example.com",
		"active":     true,
		"created_at": "2023-01-01T12:00:00Z",
	}

	user, err := deserializeForType[*DeserializeUser](row)
	if err != nil {
		t.Fatalf("DeserializeForType failed: %v", err)
	}

	if user == nil {
		t.Fatal("Expected user to be non-nil")
	}
	if user.ID != 789 {
		t.Errorf("Expected ID 789, got %d", user.ID)
	}
	if user.Name != "Generic User" {
		t.Errorf("Expected Name 'Generic User', got %q", user.Name)
	}
}

func TestDeserialize_NilDest(t *testing.T) {
	var user *DeserializeUser = nil
	row := map[string]any{"id": 123}

	err := deserialize(row, user)
	if err == nil {
		t.Error("Expected error for nil dest")
	}
}

func TestDeserialize_UnknownFields(t *testing.T) {
	user := &DeserializeUser{}
	row := map[string]any{
		"id":      123,
		"unknown": "value",
		"another": 456,
	}

	// Should not error on unknown fields
	err := deserialize(row, user)
	if err != nil {
		t.Fatalf("Deserialize should ignore unknown fields, got error: %v", err)
	}

	if user.ID != 123 {
		t.Errorf("Expected ID 123, got %d", user.ID)
	}
}
