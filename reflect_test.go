package typedb

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

// Test models for reflection tests
type TestReflectUser struct {
	Model
	Email string `db:"email" load:"unique"`
	Name  string `db:"name"`
	ID    int    `db:"id" load:"primary"`
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
	userPtr := &TestReflectUser{}

	// Test pointer type (required)
	modelType := getModelType(userPtr)
	if modelType.Kind() != reflect.Struct {
		t.Errorf("getModelType(pointer) should return struct type, got %v", modelType.Kind())
	}
	if modelType != reflect.TypeOf(TestReflectUser{}) {
		t.Errorf("getModelType(pointer) = %v, want %v", modelType, reflect.TypeOf(TestReflectUser{}))
	}

	// Test that value type panics
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for value type")
		}
	}()
	getModelType(TestReflectUser{})
}

func TestFindFieldByTag(t *testing.T) {
	user := &TestReflectUser{}

	// Test finding primary field
	field, found := findFieldByTag(user, "load", "primary")
	if !found {
		t.Fatal("Expected to find field with load:\"primary\" tag")
	}
	if field.Name != "ID" {
		t.Errorf("Expected field name 'ID', got %q", field.Name)
	}

	// Test finding unique field
	field, found = findFieldByTag(user, "load", "unique")
	if !found {
		t.Fatal("Expected to find field with load:\"unique\" tag")
	}
	if field.Name != "Email" {
		t.Errorf("Expected field name 'Email', got %q", field.Name)
	}

	// Test finding non-existent tag
	_, found = findFieldByTag(user, "load", "nonexistent")
	if found {
		t.Error("Expected not to find field with load:\"nonexistent\" tag")
	}

	// Test finding composite tag
	post := &TestReflectPost{}
	field, found = findFieldByTag(post, "load", "composite:userpost")
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
	user := &TestReflectUser{}

	// Should still find ID field even though Model is embedded
	field, found := findFieldByTag(user, "load", "primary")
	if !found {
		t.Fatal("Expected to find field through embedded struct")
	}
	if field.Name != "ID" {
		t.Errorf("Expected field name 'ID', got %q", field.Name)
	}
}

func TestGetFieldValue(t *testing.T) {
	user := &TestReflectUser{
		ID:    42,
		Email: "test@example.com",
		Name:  "Test User",
	}

	// Test getting value from pointer (required)
	value, err := getFieldValue(user, "ID")
	if err != nil {
		t.Fatalf("getFieldValue failed: %v", err)
	}
	if value.Int() != 42 {
		t.Errorf("getFieldValue(ID) = %v, want 42", value.Int())
	}

	// Test getting value from pointer
	value, err = getFieldValue(user, "Email")
	if err != nil {
		t.Fatalf("getFieldValue failed: %v", err)
	}
	if value.String() != "test@example.com" {
		t.Errorf("getFieldValue(Email) = %q, want %q", value.String(), "test@example.com")
	}

	// Test non-existent field
	_, err = getFieldValue(user, "NonExistent")
	if err == nil {
		t.Error("Expected error for non-existent field")
	}
	if !errors.Is(err, ErrFieldNotFound) {
		t.Errorf("Expected ErrFieldNotFound, got %v", err)
	}

	// Test nil pointer
	var nilUser *TestReflectUser
	_, err = getFieldValue(nilUser, "ID")
	if err == nil {
		t.Error("Expected error for nil pointer")
	}

	// Test value type (should fail)
	userValue := TestReflectUser{ID: 42}
	_, err = getFieldValue(userValue, "ID")
	if err == nil {
		t.Error("Expected error for value type")
	}
}

func TestSetFieldValue(t *testing.T) {
	user := &TestReflectUser{
		ID:    42,
		Email: "test@example.com",
	}

	// Test setting value
	err := setFieldValue(user, "ID", 100)
	if err != nil {
		t.Fatalf("SetFieldValue failed: %v", err)
	}
	if user.ID != 100 {
		t.Errorf("setFieldValue(ID) = %d, want 100", user.ID)
	}

	// Test setting string value
	err = setFieldValue(user, "Email", "new@example.com")
	if err != nil {
		t.Fatalf("SetFieldValue failed: %v", err)
	}
	if user.Email != "new@example.com" {
		t.Errorf("setFieldValue(Email) = %q, want %q", user.Email, "new@example.com")
	}

	// Test non-existent field
	err = setFieldValue(user, "NonExistent", "value")
	if err == nil {
		t.Error("Expected error for non-existent field")
	}
	if !errors.Is(err, ErrFieldNotFound) {
		t.Errorf("Expected ErrFieldNotFound, got %v", err)
	}

	// Test value type (should fail)
	userValue := TestReflectUser{}
	err = setFieldValue(userValue, "ID", 100)
	if err == nil {
		t.Error("Expected error for value type (not pointer)")
	}

	// Test nil pointer
	var nilUser *TestReflectUser
	err = setFieldValue(nilUser, "ID", 100)
	if err == nil {
		t.Error("Expected error for nil pointer")
	}

	// Test incompatible type
	err = setFieldValue(user, "ID", "not an int")
	if err == nil {
		t.Error("Expected error for incompatible type")
	}
}

func TestFindMethod(t *testing.T) {
	user := &TestReflectUser{}

	// Test finding Deserialize method (from ModelInterface)
	method, found := findMethod(user, "Deserialize")
	if !found {
		t.Fatal("Expected to find Deserialize method")
	}
	// method cannot be nil if found is true (findMethod guarantees this)
	if method.Name != "Deserialize" {
		t.Errorf("Expected method name 'Deserialize', got %q", method.Name)
	}

	// Test non-existent method
	method, found = findMethod(user, "NonExistentMethod")
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

	results, err := callMethod(user, "Deserialize", row)
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
	_, err = callMethod(user, "NonExistentMethod")
	if err == nil {
		t.Error("Expected error for non-existent method")
	}
	if !errors.Is(err, ErrMethodNotFound) {
		t.Errorf("Expected ErrMethodNotFound, got %v", err)
	}
}

func TestFindFieldByTag_CompositeTag(t *testing.T) {
	post := &TestReflectPost{}

	// Test finding composite tag with colon
	field, found := findFieldByTag(post, "load", "composite:userpost")
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
	// Test that non-pointer types panic (GetModelType requires pointer)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for non-pointer type")
		}
	}()
	var notStruct int
	findFieldByTag(notStruct, "load", "primary")
}

func TestFindFieldByTagRecursive_NonStructTypeDefensive(t *testing.T) {
	// Test the defensive check for non-struct types in findFieldByTagRecursive
	// This shouldn't happen in practice since GetModelType ensures struct type,
	// but we test it for completeness by calling the internal function directly

	// Create a reflect.Type that's not a struct (int type)
	intType := reflect.TypeOf(0)

	// Call the internal function directly to test the defensive check
	field, found := findFieldByTagRecursive(intType, "load", "primary")
	if found {
		t.Error("Expected not to find field in non-struct type")
	}
	if field != nil {
		t.Error("Expected nil field for non-struct type")
	}
}

func TestFindFieldByNameRecursive_NonStructTypeDefensive(t *testing.T) {
	// Test the defensive check for non-struct types in findFieldByNameRecursive
	// This shouldn't happen in practice since getFieldValue ensures struct type,
	// but we test it via getFieldValue with a pointer to non-struct type
	var intPtr *int
	val := 42
	intPtr = &val

	_, err := getFieldValue(intPtr, "Field")
	if err == nil {
		t.Error("Expected error for pointer to non-struct type")
	}
	// The error should be about pointer to struct, not about non-struct type in recursive function
	// But this exercises the path where we check v.Kind() != reflect.Struct after Elem()
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

	model := &TestModel{
		EmbeddedStruct: &EmbeddedStruct{Value: 42},
		ID:             100,
	}

	// Should find field in embedded pointer struct
	field, found := findFieldByTag(model, "load", "embedded")
	if !found {
		t.Fatal("Expected to find field in embedded pointer struct")
	}
	if field.Name != "Value" {
		t.Errorf("Expected field name 'Value', got %q", field.Name)
	}

	// Should also find primary field
	field, found = findFieldByTag(model, "load", "primary")
	if !found {
		t.Fatal("Expected to find primary field")
	}
	if field.Name != "ID" {
		t.Errorf("Expected field name 'ID', got %q", field.Name)
	}
}

func TestGetFieldValue_NonPointerType(t *testing.T) {
	// Test that non-pointer types return error
	var notStruct int
	_, err := getFieldValue(notStruct, "Field")
	if err == nil {
		t.Error("Expected error for non-pointer type")
	}

	// Test pointer to non-struct
	var intPtr *int
	val := 42
	intPtr = &val
	_, err = getFieldValue(intPtr, "Field")
	if err == nil {
		t.Error("Expected error for pointer to non-struct type")
	}
}

func TestSetFieldValue_NonPointerType(t *testing.T) {
	// Test that non-pointer types return error
	var notStruct int
	err := setFieldValue(notStruct, "Field", 42)
	if err == nil {
		t.Error("Expected error for non-pointer type")
	}

	// Test pointer to non-struct
	var intPtr *int
	val := 42
	intPtr = &val
	err = setFieldValue(intPtr, "Field", 42)
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
	err := setFieldValue(model, "Exported", 100)
	if err != nil {
		t.Fatalf("SetFieldValue failed: %v", err)
	}
	if model.Exported != 100 {
		t.Errorf("setFieldValue(Exported) = %d, want 100", model.Exported)
	}

	// Should fail to set unexported field (CanSet() returns false)
	err = setFieldValue(model, "unexported", 200)
	if err == nil {
		t.Error("Expected error for unsettable (unexported) field")
	}
	if err.Error() != "typedb: field unexported cannot be set" {
		t.Errorf("Expected 'cannot be set' error, got %q", err.Error())
	}
}

func TestFindFieldByNameRecursive_NonPointerType(t *testing.T) {
	// Test that non-pointer types return error
	var notStruct int
	_, err := getFieldValue(notStruct, "Field")
	if err == nil {
		t.Error("Expected error for non-pointer type")
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
	value, err := getFieldValue(model, "EmbeddedField")
	if err != nil {
		t.Fatalf("getFieldValue failed: %v", err)
	}
	if value.String() != "test" {
		t.Errorf("getFieldValue(EmbeddedField) = %q, want %q", value.String(), "test")
	}

	// Should also find direct field
	value, err = getFieldValue(model, "ID")
	if err != nil {
		t.Fatalf("getFieldValue failed: %v", err)
	}
	if value.Int() != 100 {
		t.Errorf("getFieldValue(ID) = %v, want 100", value.Int())
	}

	// Should be able to set field in embedded pointer struct
	err = setFieldValue(model, "EmbeddedField", "updated")
	if err != nil {
		t.Fatalf("SetFieldValue failed: %v", err)
	}
	if model.EmbeddedStruct.EmbeddedField != "updated" {
		t.Errorf("setFieldValue(EmbeddedField) = %q, want %q", model.EmbeddedStruct.EmbeddedField, "updated")
	}
}

func TestFindFieldByNameRecursive_BaseModelPattern(t *testing.T) {
	// Test realistic base model pattern: AdminUser -> BaseUser -> typedb.Model
	// This represents a common inheritance pattern where multiple user types share a base
	type BaseUser struct {
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
		Model
		ID int `db:"id" load:"primary"`
	}
	type AdminUser struct {
		BaseUser
		Department string `db:"department"`
		AdminLevel int    `db:"admin_level"`
	}

	model := &AdminUser{
		BaseUser: BaseUser{
			ID:        123,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		AdminLevel: 5,
		Department: "IT",
	}

	// Should find field in BaseUser (one level deep)
	value, err := getFieldValue(model, "ID")
	if err != nil {
		t.Fatalf("getFieldValue failed: %v", err)
	}
	if value.Int() != 123 {
		t.Errorf("getFieldValue(ID) = %v, want 123", value.Int())
	}

	// Should find field in AdminUser (direct)
	value, err = getFieldValue(model, "AdminLevel")
	if err != nil {
		t.Fatalf("getFieldValue failed: %v", err)
	}
	if value.Int() != 5 {
		t.Errorf("getFieldValue(AdminLevel) = %v, want 5", value.Int())
	}

	// Should find field in BaseUser via tag
	field, found := findFieldByTag(model, "load", "primary")
	if !found {
		t.Fatal("Expected to find primary field in BaseUser")
	}
	if field.Name != "ID" {
		t.Errorf("Expected field name 'ID', got %q", field.Name)
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
	value, err := getFieldValue(model, "FieldA")
	if err != nil {
		t.Fatalf("getFieldValue failed: %v", err)
	}
	if value.String() != "a" {
		t.Errorf("getFieldValue(FieldA) = %q, want %q", value.String(), "a")
	}

	// Should find field in second embedded struct
	value, err = getFieldValue(model, "FieldB")
	if err != nil {
		t.Fatalf("getFieldValue failed: %v", err)
	}
	if value.Int() != 42 {
		t.Errorf("getFieldValue(FieldB) = %v, want 42", value.Int())
	}

	// Should find direct field
	value, err = getFieldValue(model, "Direct")
	if err != nil {
		t.Fatalf("getFieldValue failed: %v", err)
	}
	if value.Bool() != true {
		t.Errorf("getFieldValue(Direct) = %v, want true", value.Bool())
	}

	// Should be able to set fields in both embedded structs
	err = setFieldValue(model, "FieldA", "updated")
	if err != nil {
		t.Fatalf("SetFieldValue failed: %v", err)
	}
	if model.EmbeddedA.FieldA != "updated" {
		t.Errorf("setFieldValue(FieldA) = %q, want %q", model.EmbeddedA.FieldA, "updated")
	}

	err = setFieldValue(model, "FieldB", 100)
	if err != nil {
		t.Fatalf("SetFieldValue failed: %v", err)
	}
	if model.EmbeddedB.FieldB != 100 {
		t.Errorf("setFieldValue(FieldB) = %v, want 100", model.EmbeddedB.FieldB)
	}
}

func TestFindFieldByNameRecursive_FieldNotFound(t *testing.T) {
	// Test realistic scenario: field doesn't exist in BaseUser or AdminUser
	type BaseUser struct {
		CreatedAt time.Time
		Model
		ID int `db:"id" load:"primary"`
	}
	type AdminUser struct {
		BaseUser
		AdminLevel int
	}

	model := &AdminUser{
		BaseUser: BaseUser{
			ID:        123,
			CreatedAt: time.Now(),
		},
		AdminLevel: 5,
	}

	// Search for field that doesn't exist anywhere
	_, err := getFieldValue(model, "NonExistentField")
	if err == nil {
		t.Error("Expected error for non-existent field")
	}
	if !errors.Is(err, ErrFieldNotFound) {
		t.Errorf("Expected ErrFieldNotFound, got %v", err)
	}
}

func TestFindFieldByTagRecursive_EmbeddedStructNotFound(t *testing.T) {
	// Test case where embedded struct exists but doesn't contain the field
	// This exercises the recursive return-false path
	type EmbeddedA struct {
		FieldA string
	}
	type EmbeddedB struct {
		FieldB int
	}
	type TestModel struct {
		EmbeddedA // Embedded
		EmbeddedB // Embedded
		Direct    bool
	}

	model := &TestModel{
		EmbeddedA: EmbeddedA{FieldA: "a"},
		EmbeddedB: EmbeddedB{FieldB: 42},
		Direct:    true,
	}

	// Search for field that doesn't exist in any embedded struct
	field, found := findFieldByTag(model, "load", "nonexistent")
	if found {
		t.Error("Expected not to find field")
	}
	if field != nil {
		t.Error("Expected nil field")
	}
}
