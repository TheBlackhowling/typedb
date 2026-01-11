package typedb

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

// Test models for deserialization
type DeserializeUser struct {
	Model
	ID        int       `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Active    bool      `db:"active"`
	CreatedAt time.Time `db:"created_at"`
}

func (u *DeserializeUser) Deserialize(row map[string]any) error {
	return Deserialize(row, u)
}

type DeserializePost struct {
	Model
	ID        int    `db:"id"`
	UserID    int    `db:"user_id"`
	Title     string `db:"title"`
	Content   string `db:"content"`
	Published bool   `db:"published"`
}

func (p *DeserializePost) Deserialize(row map[string]any) error {
	return Deserialize(row, p)
}

type DeserializeModelWithPointers struct {
	Model
	ID      *int    `db:"id"`
	Name    *string `db:"name"`
	Active  *bool   `db:"active"`
	Deleted *bool   `db:"deleted"`
}

func (m *DeserializeModelWithPointers) Deserialize(row map[string]any) error {
	return Deserialize(row, m)
}

type DeserializeModelWithArrays struct {
	Model
	ID      int      `db:"id"`
	Tags    []string `db:"tags"`
	Numbers []int    `db:"numbers"`
}

func (m *DeserializeModelWithArrays) Deserialize(row map[string]any) error {
	return Deserialize(row, m)
}

type DeserializeModelWithJSON struct {
	Model
	ID      int            `db:"id"`
	Metadata map[string]any `db:"metadata"`
	Config   map[string]string `db:"config"`
}

func (m *DeserializeModelWithJSON) Deserialize(row map[string]any) error {
	return Deserialize(row, m)
}

type DeserializeModelWithDotNotation struct {
	Model
	ID   int    `db:"users.id"`
	Name string `db:"users.name"`
}

func (m *DeserializeModelWithDotNotation) Deserialize(row map[string]any) error {
	return Deserialize(row, m)
}

type BaseModel struct {
	Model
	ID int `db:"id"`
}

type DerivedModel struct {
	BaseModel
	Name string `db:"name"`
}

func (m *DerivedModel) Deserialize(row map[string]any) error {
	return Deserialize(row, m)
}

func TestDeserialize_BasicTypes(t *testing.T) {
	user := &DeserializeUser{}
	row := map[string]any{
		"id":         123,
		"name":       "John Doe",
		"email":      "john@example.com",
		"active":     true,
		"created_at": "2023-01-01T12:00:00Z",
	}

	err := Deserialize(row, user)
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
		name  string
		value any
		want  int
	}{
		{"int64", int64(123), 123},
		{"int32", int32(456), 456},
		{"float64", float64(789), 789},
		{"string", "999", 999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := map[string]any{"id": tt.value}
			err := Deserialize(row, user)
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
		{"string t", "t", true},
		{"string f", "f", false},
		{"int 1", 1, true},
		{"int 0", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := map[string]any{"published": tt.value}
			err := Deserialize(row, post)
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

	err := Deserialize(row, model)
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

	err := Deserialize(row, model)
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
			err := Deserialize(row, model)
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
			err := Deserialize(row, model)
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

	err := Deserialize(row, model)
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

	err := Deserialize(row, model)
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

	err := Deserialize(row, model)
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

	user, err := DeserializeForType[*DeserializeUser](row)
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

func TestDeserialize_NonPointerDest(t *testing.T) {
	// This test can't actually be written because Deserialize requires ModelInterface,
	// and all models implement it with pointer receivers. The check for pointer type
	// happens at the start of Deserialize, but we can't pass a non-pointer that satisfies
	// ModelInterface. The type system enforces this at compile time.
	t.Skip("Cannot test non-pointer dest due to ModelInterface requiring pointer receivers")
}

func TestDeserialize_NilDest(t *testing.T) {
	var user *DeserializeUser = nil
	row := map[string]any{"id": 123}

	err := Deserialize(row, user)
	if err == nil {
		t.Error("Expected error for nil dest")
	}
}

func TestDeserialize_UnknownFields(t *testing.T) {
	user := &DeserializeUser{}
	row := map[string]any{
		"id":        123,
		"unknown":   "value",
		"another":   456,
	}

	// Should not error on unknown fields
	err := Deserialize(row, user)
	if err != nil {
		t.Fatalf("Deserialize should ignore unknown fields, got error: %v", err)
	}

	if user.ID != 123 {
		t.Errorf("Expected ID 123, got %d", user.ID)
	}
}

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
			err := Deserialize(row, user)
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
			got, err := DeserializeInt(tt.value)
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
			got, err := DeserializeBool(tt.value)
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
			got := DeserializeString(tt.value)
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
			got, err := DeserializeJSONB(tt.value)
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
	_, err := DeserializeJSONB("invalid json")
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestDeserializeJSONB_UnsupportedType(t *testing.T) {
	_, err := DeserializeJSONB(123)
	if err == nil {
		t.Error("Expected error for unsupported type")
	}
}

func TestDeserializeIntArray_InvalidFormat(t *testing.T) {
	_, err := DeserializeIntArray("invalid")
	if err == nil {
		t.Error("Expected error for invalid array format")
	}
}

func TestDeserializeStringArray_InvalidFormat(t *testing.T) {
	_, err := DeserializeStringArray(123)
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
			got, err := DeserializeMap(tt.value)
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
	_, err := DeserializeMap(123)
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
			got, err := DeserializeInt64(tt.value)
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
			got, err := DeserializeInt32(tt.value)
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
			got, err := DeserializeTime(tt.value)
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

	err := DeserializeToField(&customInt, 123)
	if err != nil {
		t.Fatalf("DeserializeToField failed: %v", err)
	}
	if customInt != 123 {
		t.Errorf("Expected 123, got %d", customInt)
	}
}

func TestDeserializeToField_PointerType(t *testing.T) {
	var intPtr *int
	err := DeserializeToField(&intPtr, 456)
	if err != nil {
		t.Fatalf("DeserializeToField failed: %v", err)
	}
	if intPtr == nil || *intPtr != 456 {
		t.Errorf("Expected *456, got %v", intPtr)
	}
}

func TestDeserializeToField_NilValue(t *testing.T) {
	var intPtr *int
	err := DeserializeToField(&intPtr, nil)
	if err != nil {
		t.Fatalf("DeserializeToField failed: %v", err)
	}
	if intPtr != nil {
		t.Error("Expected nil pointer")
	}

	var i int
	err = DeserializeToField(&i, nil)
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
	return Deserialize(row, b)
}

func TestDeserializeForType_ErrorPath(t *testing.T) {
	// Test error path when Deserialize fails
	row := map[string]any{
		"id": "invalid", // This will cause an error in Deserialize
	}

	// BadModel implements ModelInterface via Deserialize method
	// But Deserialize will fail because "invalid" can't be converted to int
	_, err := DeserializeForType[*BadModel](row)
	if err == nil {
		t.Error("Expected error for invalid deserialization")
	}
}

