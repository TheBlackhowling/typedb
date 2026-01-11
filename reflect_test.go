package typedb

import (
	"errors"
	"reflect"
	"testing"
)

// Test models for reflection tests
type TestReflectUser struct {
	Model
	ID    int    `db:"id" load:"primary"`
	Email string `db:"email" load:"unique"`
	Name  string `db:"name"`
}

func (u *TestReflectUser) Deserialize(row map[string]any) error {
	// Stub implementation for testing
	return nil
}

type TestReflectPost struct {
	Model
	ID     int `db:"id" load:"primary"`
	UserID int `db:"user_id" load:"composite:userpost"`
	PostID int `db:"post_id" load:"composite:userpost"`
}

func (p *TestReflectPost) Deserialize(row map[string]any) error {
	// Stub implementation for testing
	return nil
}

func TestGetModelType(t *testing.T) {
	user := TestReflectUser{}
	userPtr := &TestReflectUser{}

	// Test value type
	t1 := GetModelType(user)
	if t1.Kind() != reflect.Struct {
		t.Errorf("GetModelType(value) should return struct type, got %v", t1.Kind())
	}
	if t1 != reflect.TypeOf(TestReflectUser{}) {
		t.Errorf("GetModelType(value) = %v, want %v", t1, reflect.TypeOf(TestReflectUser{}))
	}

	// Test pointer type
	t2 := GetModelType(userPtr)
	if t2.Kind() != reflect.Struct {
		t.Errorf("GetModelType(pointer) should return struct type, got %v", t2.Kind())
	}
	if t2 != reflect.TypeOf(TestReflectUser{}) {
		t.Errorf("GetModelType(pointer) = %v, want %v", t2, reflect.TypeOf(TestReflectUser{}))
	}
}

func TestFindFieldByTag(t *testing.T) {
	user := TestReflectUser{}

	// Test finding primary field
	field, found := FindFieldByTag(user, "load", "primary")
	if !found {
		t.Fatal("Expected to find field with load:\"primary\" tag")
	}
	if field.Name != "ID" {
		t.Errorf("Expected field name 'ID', got %q", field.Name)
	}

	// Test finding unique field
	field, found = FindFieldByTag(user, "load", "unique")
	if !found {
		t.Fatal("Expected to find field with load:\"unique\" tag")
	}
	if field.Name != "Email" {
		t.Errorf("Expected field name 'Email', got %q", field.Name)
	}

	// Test finding non-existent tag
	field, found = FindFieldByTag(user, "load", "nonexistent")
	if found {
		t.Error("Expected not to find field with load:\"nonexistent\" tag")
	}

	// Test finding composite tag
	post := TestReflectPost{}
	field, found = FindFieldByTag(post, "load", "composite:userpost")
	if !found {
		t.Fatal("Expected to find field with load:\"composite:userpost\" tag")
	}
	// Should find one of the composite fields
	if field.Name != "UserID" && field.Name != "PostID" {
		t.Errorf("Expected field name 'UserID' or 'PostID', got %q", field.Name)
	}
}

func TestFindFieldByTag_EmbeddedStruct(t *testing.T) {
	// Test that we can find fields through embedded Model struct
	// (though Model doesn't have any tags, this tests the recursive traversal)
	user := TestReflectUser{}

	// Should still find ID field even though Model is embedded
	field, found := FindFieldByTag(user, "load", "primary")
	if !found {
		t.Fatal("Expected to find field through embedded struct")
	}
	if field.Name != "ID" {
		t.Errorf("Expected field name 'ID', got %q", field.Name)
	}
}

func TestGetFieldValue(t *testing.T) {
	user := TestReflectUser{
		ID:    42,
		Email: "test@example.com",
		Name:  "Test User",
	}

	// Test getting value from struct
	value, err := GetFieldValue(user, "ID")
	if err != nil {
		t.Fatalf("GetFieldValue failed: %v", err)
	}
	if value.Int() != 42 {
		t.Errorf("GetFieldValue(ID) = %v, want 42", value.Int())
	}

	// Test getting value from pointer
	value, err = GetFieldValue(&user, "Email")
	if err != nil {
		t.Fatalf("GetFieldValue failed: %v", err)
	}
	if value.String() != "test@example.com" {
		t.Errorf("GetFieldValue(Email) = %q, want %q", value.String(), "test@example.com")
	}

	// Test non-existent field
	_, err = GetFieldValue(user, "NonExistent")
	if err == nil {
		t.Error("Expected error for non-existent field")
	}
	if !errors.Is(err, ErrFieldNotFound) {
		t.Errorf("Expected ErrFieldNotFound, got %v", err)
	}

	// Test nil pointer
	var nilUser *TestReflectUser
	_, err = GetFieldValue(nilUser, "ID")
	if err == nil {
		t.Error("Expected error for nil pointer")
	}
}

func TestSetFieldValue(t *testing.T) {
	user := &TestReflectUser{
		ID:    42,
		Email: "test@example.com",
	}

	// Test setting value
	err := SetFieldValue(user, "ID", 100)
	if err != nil {
		t.Fatalf("SetFieldValue failed: %v", err)
	}
	if user.ID != 100 {
		t.Errorf("SetFieldValue(ID) = %d, want 100", user.ID)
	}

	// Test setting string value
	err = SetFieldValue(user, "Email", "new@example.com")
	if err != nil {
		t.Fatalf("SetFieldValue failed: %v", err)
	}
	if user.Email != "new@example.com" {
		t.Errorf("SetFieldValue(Email) = %q, want %q", user.Email, "new@example.com")
	}

	// Test non-existent field
	err = SetFieldValue(user, "NonExistent", "value")
	if err == nil {
		t.Error("Expected error for non-existent field")
	}
	if !errors.Is(err, ErrFieldNotFound) {
		t.Errorf("Expected ErrFieldNotFound, got %v", err)
	}

	// Test value type (should fail)
	userValue := TestReflectUser{}
	err = SetFieldValue(userValue, "ID", 100)
	if err == nil {
		t.Error("Expected error for value type (not pointer)")
	}

	// Test nil pointer
	var nilUser *TestReflectUser
	err = SetFieldValue(nilUser, "ID", 100)
	if err == nil {
		t.Error("Expected error for nil pointer")
	}

	// Test incompatible type
	err = SetFieldValue(user, "ID", "not an int")
	if err == nil {
		t.Error("Expected error for incompatible type")
	}
}

func TestFindMethod(t *testing.T) {
	user := &TestReflectUser{}

	// Test finding Deserialize method (from ModelInterface)
	method, found := FindMethod(user, "Deserialize")
	if !found {
		t.Fatal("Expected to find Deserialize method")
	}
	if method == nil {
		t.Error("Method should not be nil")
	}
	if method.Name != "Deserialize" {
		t.Errorf("Expected method name 'Deserialize', got %q", method.Name)
	}

	// Test non-existent method
	method, found = FindMethod(user, "NonExistentMethod")
	if found {
		t.Error("Expected not to find non-existent method")
	}
	if method != nil {
		t.Error("Method should be nil for non-existent method")
	}
}

func TestCallMethod(t *testing.T) {
	user := &TestReflectUser{}

	// Test calling Deserialize method
	row := map[string]any{
		"id":    42,
		"email": "test@example.com",
		"name":  "Test User",
	}

	results, err := CallMethod(user, "Deserialize", row)
	if err != nil {
		t.Fatalf("CallMethod failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	// Deserialize should return error (nil in this case since it's a stub)
	if !results[0].IsNil() {
		t.Error("Expected Deserialize to return nil error")
	}

	// Test non-existent method
	_, err = CallMethod(user, "NonExistentMethod")
	if err == nil {
		t.Error("Expected error for non-existent method")
	}
	if !errors.Is(err, ErrMethodNotFound) {
		t.Errorf("Expected ErrMethodNotFound, got %v", err)
	}
}

func TestFindFieldByTag_CompositeTag(t *testing.T) {
	post := TestReflectPost{}

	// Test finding composite tag with colon
	field, found := FindFieldByTag(post, "load", "composite:userpost")
	if !found {
		t.Fatal("Expected to find field with composite:userpost tag")
	}
	if field.Name != "UserID" && field.Name != "PostID" {
		t.Errorf("Expected UserID or PostID, got %q", field.Name)
	}
}

func TestContainsTagValue(t *testing.T) {
	tests := []struct {
		tag    string
		value  string
		result bool
	}{
		{"primary", "primary", true},
		{"primary,unique", "primary", true},
		{"primary,unique", "unique", true},
		{"primary,unique", "composite", false},
		{"composite:userpost", "composite:userpost", true},
		{"primary,composite:userpost", "primary", true},
		{"primary,composite:userpost", "composite:userpost", true},
		{"", "primary", false},
		{"primary", "", false}, // Empty value
	}

	for _, tt := range tests {
		result := containsTagValue(tt.tag, tt.value)
		if result != tt.result {
			t.Errorf("containsTagValue(%q, %q) = %v, want %v", tt.tag, tt.value, result, tt.result)
		}
	}
}

func TestSplitTag(t *testing.T) {
	tests := []struct {
		tag    string
		result []string
	}{
		{"", nil},
		{"primary", []string{"primary"}},
		{"primary,unique", []string{"primary", "unique"}},
		{"primary, unique", []string{"primary", "unique"}}, // With spaces
		{"primary,unique,composite:userpost", []string{"primary", "unique", "composite:userpost"}},
		{"  primary  ,  unique  ", []string{"primary", "unique"}}, // Extra spaces
	}

	for _, tt := range tests {
		result := splitTag(tt.tag)
		if len(result) != len(tt.result) {
			t.Errorf("splitTag(%q) = %v, want %v", tt.tag, result, tt.result)
			continue
		}
		for i := range result {
			if result[i] != tt.result[i] {
				t.Errorf("splitTag(%q)[%d] = %q, want %q", tt.tag, i, result[i], tt.result[i])
			}
		}
	}
}

func TestFindFieldByTagRecursive_NonStructType(t *testing.T) {
	// Test that non-struct types return false
	var notStruct int
	field, found := FindFieldByTag(notStruct, "load", "primary")
	if found {
		t.Error("Expected not to find field in non-struct type")
	}
	if field != nil {
		t.Error("Expected nil field for non-struct type")
	}
}

func TestFindFieldByTagRecursive_EmbeddedPointer(t *testing.T) {
	// Test model with embedded pointer struct
	type EmbeddedStruct struct {
		Value int `load:"embedded"`
	}
	type TestModel struct {
		*EmbeddedStruct     // Embedded pointer
		ID              int `load:"primary"`
	}

	model := TestModel{
		EmbeddedStruct: &EmbeddedStruct{Value: 42},
		ID:             100,
	}

	// Should find field in embedded pointer struct
	field, found := FindFieldByTag(model, "load", "embedded")
	if !found {
		t.Fatal("Expected to find field in embedded pointer struct")
	}
	if field.Name != "Value" {
		t.Errorf("Expected field name 'Value', got %q", field.Name)
	}

	// Should also find primary field
	field, found = FindFieldByTag(model, "load", "primary")
	if !found {
		t.Fatal("Expected to find primary field")
	}
	if field.Name != "ID" {
		t.Errorf("Expected field name 'ID', got %q", field.Name)
	}
}

func TestGetFieldValue_NonStructType(t *testing.T) {
	// Test that non-struct types return error
	var notStruct int
	_, err := GetFieldValue(notStruct, "Field")
	if err == nil {
		t.Error("Expected error for non-struct type")
	}

	// Test pointer to non-struct
	var intPtr *int
	val := 42
	intPtr = &val
	_, err = GetFieldValue(intPtr, "Field")
	if err == nil {
		t.Error("Expected error for pointer to non-struct type")
	}
}

func TestSetFieldValue_NonStructType(t *testing.T) {
	// Test that non-struct types return error
	var notStruct int
	err := SetFieldValue(notStruct, "Field", 42)
	if err == nil {
		t.Error("Expected error for non-struct type")
	}

	// Test pointer to non-struct
	var intPtr *int
	val := 42
	intPtr = &val
	err = SetFieldValue(intPtr, "Field", 42)
	if err == nil {
		t.Error("Expected error for pointer to non-struct type")
	}
}

func TestSetFieldValue_UnsettableField(t *testing.T) {
	// Test model with unexported field (cannot be set)
	type TestModel struct {
		unexported int // lowercase = unexported, cannot be set from outside package
		Exported   int
	}

	model := &TestModel{
		unexported: 1,
		Exported:   2,
	}

	// Should be able to set exported field
	err := SetFieldValue(model, "Exported", 100)
	if err != nil {
		t.Fatalf("SetFieldValue failed: %v", err)
	}
	if model.Exported != 100 {
		t.Errorf("SetFieldValue(Exported) = %d, want 100", model.Exported)
	}

	// Should fail to set unexported field (CanSet() returns false)
	err = SetFieldValue(model, "unexported", 200)
	if err == nil {
		t.Error("Expected error for unsettable (unexported) field")
	}
	if err.Error() != "typedb: field unexported cannot be set" {
		t.Errorf("Expected 'cannot be set' error, got %q", err.Error())
	}
}

func TestFindFieldByNameRecursive_NonStructType(t *testing.T) {
	// Test that non-struct types return false
	var notStruct int
	_, err := GetFieldValue(notStruct, "Field")
	if err == nil {
		t.Error("Expected error for non-struct type")
	}
	if !errors.Is(err, ErrFieldNotFound) {
		// Should fail before checking field, but let's verify the error path
	}
}

func TestFindFieldByNameRecursive_EmbeddedPointer(t *testing.T) {
	// Test model with embedded pointer struct
	type EmbeddedStruct struct {
		EmbeddedField string
	}
	type TestModel struct {
		*EmbeddedStruct // Embedded pointer
		ID              int
	}

	model := &TestModel{
		EmbeddedStruct: &EmbeddedStruct{EmbeddedField: "test"},
		ID:             100,
	}

	// Should find field in embedded pointer struct
	value, err := GetFieldValue(model, "EmbeddedField")
	if err != nil {
		t.Fatalf("GetFieldValue failed: %v", err)
	}
	if value.String() != "test" {
		t.Errorf("GetFieldValue(EmbeddedField) = %q, want %q", value.String(), "test")
	}

	// Should also find direct field
	value, err = GetFieldValue(model, "ID")
	if err != nil {
		t.Fatalf("GetFieldValue failed: %v", err)
	}
	if value.Int() != 100 {
		t.Errorf("GetFieldValue(ID) = %v, want 100", value.Int())
	}

	// Should be able to set field in embedded pointer struct
	err = SetFieldValue(model, "EmbeddedField", "updated")
	if err != nil {
		t.Fatalf("SetFieldValue failed: %v", err)
	}
	if model.EmbeddedStruct.EmbeddedField != "updated" {
		t.Errorf("SetFieldValue(EmbeddedField) = %q, want %q", model.EmbeddedStruct.EmbeddedField, "updated")
	}
}

func TestFindFieldByNameRecursive_NestedEmbedding(t *testing.T) {
	// Test deeply nested embedded structs
	type Level3 struct {
		Level3Field string
	}
	type Level2 struct {
		Level3      // Embedded
		Level2Field int
	}
	type Level1 struct {
		Level2      // Embedded
		Level1Field bool
	}
	type TestModel struct {
		Level1 // Embedded
		ID     int
	}

	model := &TestModel{
		Level1: Level1{
			Level2: Level2{
				Level3:      Level3{Level3Field: "deep"},
				Level2Field: 42,
			},
			Level1Field: true,
		},
		ID: 100,
	}

	// Should find field in deeply nested embedded struct
	value, err := GetFieldValue(model, "Level3Field")
	if err != nil {
		t.Fatalf("GetFieldValue failed: %v", err)
	}
	if value.String() != "deep" {
		t.Errorf("GetFieldValue(Level3Field) = %q, want %q", value.String(), "deep")
	}

	// Should find field in intermediate level
	value, err = GetFieldValue(model, "Level2Field")
	if err != nil {
		t.Fatalf("GetFieldValue failed: %v", err)
	}
	if value.Int() != 42 {
		t.Errorf("GetFieldValue(Level2Field) = %v, want 42", value.Int())
	}

	// Should find field in first level
	value, err = GetFieldValue(model, "Level1Field")
	if err != nil {
		t.Fatalf("GetFieldValue failed: %v", err)
	}
	if value.Bool() != true {
		t.Errorf("GetFieldValue(Level1Field) = %v, want true", value.Bool())
	}

	// Should find direct field
	value, err = GetFieldValue(model, "ID")
	if err != nil {
		t.Fatalf("GetFieldValue failed: %v", err)
	}
	if value.Int() != 100 {
		t.Errorf("GetFieldValue(ID) = %v, want 100", value.Int())
	}
}

func TestFindFieldByNameRecursive_MultipleEmbeddedStructs(t *testing.T) {
	// Test model with multiple embedded structs
	type EmbeddedA struct {
		FieldA string
	}
	type EmbeddedB struct {
		FieldB int
	}
	type TestModel struct {
		EmbeddedA // First embedded
		EmbeddedB // Second embedded
		Direct    bool
	}

	model := &TestModel{
		EmbeddedA: EmbeddedA{FieldA: "a"},
		EmbeddedB: EmbeddedB{FieldB: 42},
		Direct:    true,
	}

	// Should find field in first embedded struct
	value, err := GetFieldValue(model, "FieldA")
	if err != nil {
		t.Fatalf("GetFieldValue failed: %v", err)
	}
	if value.String() != "a" {
		t.Errorf("GetFieldValue(FieldA) = %q, want %q", value.String(), "a")
	}

	// Should find field in second embedded struct
	value, err = GetFieldValue(model, "FieldB")
	if err != nil {
		t.Fatalf("GetFieldValue failed: %v", err)
	}
	if value.Int() != 42 {
		t.Errorf("GetFieldValue(FieldB) = %v, want 42", value.Int())
	}

	// Should find direct field
	value, err = GetFieldValue(model, "Direct")
	if err != nil {
		t.Fatalf("GetFieldValue failed: %v", err)
	}
	if value.Bool() != true {
		t.Errorf("GetFieldValue(Direct) = %v, want true", value.Bool())
	}

	// Should be able to set fields in both embedded structs
	err = SetFieldValue(model, "FieldA", "updated")
	if err != nil {
		t.Fatalf("SetFieldValue failed: %v", err)
	}
	if model.EmbeddedA.FieldA != "updated" {
		t.Errorf("SetFieldValue(FieldA) = %q, want %q", model.EmbeddedA.FieldA, "updated")
	}

	err = SetFieldValue(model, "FieldB", 100)
	if err != nil {
		t.Fatalf("SetFieldValue failed: %v", err)
	}
	if model.EmbeddedB.FieldB != 100 {
		t.Errorf("SetFieldValue(FieldB) = %v, want 100", model.EmbeddedB.FieldB)
	}
}
