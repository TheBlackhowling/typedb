package typedb

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	// ErrFieldNotFound is returned when a field cannot be found
	ErrFieldNotFound = errors.New("typedb: field not found")
	// ErrMethodNotFound is returned when a method cannot be found
	ErrMethodNotFound = errors.New("typedb: method not found")
)

// GetModelType returns the reflect.Type of a model.
// Handles both pointer and value types, returning the struct type.
func GetModelType(model any) reflect.Type {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// FindFieldByTag finds a struct field by its tag value.
// Searches through embedded structs (like Model base).
// Returns the field and true if found, nil and false otherwise.
//
// Example:
//
//	type User struct {
//	    typedb.Model
//	    ID int `db:"id" load:"primary"`
//	}
//
//	field, found := FindFieldByTag(User{}, "load", "primary")
//	// field.Name == "ID", found == true
func FindFieldByTag(model any, tagKey, tagValue string) (*reflect.StructField, bool) {
	t := GetModelType(model)
	return findFieldByTagRecursive(t, tagKey, tagValue)
}

// findFieldByTagRecursive recursively searches for a field with the given tag.
// Handles embedded structs by traversing all fields.
func findFieldByTagRecursive(t reflect.Type, tagKey, tagValue string) (*reflect.StructField, bool) {
	if t.Kind() != reflect.Struct {
		return nil, false
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Check if this field has the tag value
		tag := field.Tag.Get(tagKey)
		if tag == tagValue || containsTagValue(tag, tagValue) {
			return &field, true
		}

		// If field is embedded (anonymous), search recursively
		if field.Anonymous {
			embeddedType := field.Type
			if embeddedType.Kind() == reflect.Ptr {
				embeddedType = embeddedType.Elem()
			}
			if foundField, found := findFieldByTagRecursive(embeddedType, tagKey, tagValue); found {
				return foundField, true
			}
		}
	}

	return nil, false
}

// containsTagValue checks if a tag string contains the given value.
// Handles comma-separated tag values like "primary,composite:name".
// Performs exact matching on each comma-separated part.
func containsTagValue(tag, value string) bool {
	if tag == "" || value == "" {
		return false
	}

	for _, part := range splitTag(tag) {
		if part == value {
			return true
		}
	}

	return false
}

// splitTag splits a tag string by commas.
// Go struct tag values returned by Tag.Get() are already unquoted.
func splitTag(tag string) []string {
	if tag == "" {
		return nil
	}
	parts := strings.Split(tag, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// GetFieldValue gets the value of a field by name from a model.
// Handles both pointer and value types.
// Returns an error if the field is not found or cannot be accessed.
func GetFieldValue(model any, fieldName string) (reflect.Value, error) {
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}, fmt.Errorf("typedb: cannot get field value from nil pointer")
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("typedb: model must be a struct or pointer to struct")
	}

	field, found := findFieldByNameRecursive(v.Type(), fieldName)
	if !found {
		return reflect.Value{}, fmt.Errorf("%w: %s", ErrFieldNotFound, fieldName)
	}

	return v.FieldByIndex(field.Index), nil
}

// SetFieldValue sets the value of a field by name in a model.
// Handles both pointer and value types.
// Returns an error if the field is not found, cannot be accessed, or value type is incompatible.
func SetFieldValue(model any, fieldName string, value any) error {
	v := reflect.ValueOf(model)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("typedb: model must be a pointer to set field value")
	}

	if v.IsNil() {
		return fmt.Errorf("typedb: cannot set field value on nil pointer")
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("typedb: model must be a pointer to struct")
	}

	field, found := findFieldByNameRecursive(v.Type(), fieldName)
	if !found {
		return fmt.Errorf("%w: %s", ErrFieldNotFound, fieldName)
	}

	fieldValue := v.FieldByIndex(field.Index)
	if !fieldValue.CanSet() {
		return fmt.Errorf("typedb: field %s cannot be set", fieldName)
	}

	valueV := reflect.ValueOf(value)
	if !valueV.Type().AssignableTo(fieldValue.Type()) {
		return fmt.Errorf("typedb: cannot assign %v to field %s of type %v", valueV.Type(), fieldName, fieldValue.Type())
	}

	fieldValue.Set(valueV)
	return nil
}

// findFieldByNameRecursive recursively searches for a field by name.
// Handles embedded structs by traversing all fields.
func findFieldByNameRecursive(t reflect.Type, fieldName string) (*reflect.StructField, bool) {
	if t.Kind() != reflect.Struct {
		return nil, false
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Direct match
		if field.Name == fieldName {
			return &field, true
		}

		// If field is embedded (anonymous), search recursively
		if field.Anonymous {
			embeddedType := field.Type
			if embeddedType.Kind() == reflect.Ptr {
				embeddedType = embeddedType.Elem()
			}
			if foundField, found := findFieldByNameRecursive(embeddedType, fieldName); found {
				// Create a new field with adjusted index
				adjustedField := *foundField
				adjustedField.Index = append(append([]int(nil), field.Index...), foundField.Index...)
				return &adjustedField, true
			}
		}
	}

	return nil, false
}

// FindMethod finds a method by name on a model.
// Handles both pointer and value types.
// Returns the method and true if found, nil and false otherwise.
func FindMethod(model any, methodName string) (*reflect.Method, bool) {
	t := reflect.TypeOf(model)
	method, found := t.MethodByName(methodName)
	if !found {
		return nil, false
	}
	return &method, true
}

// CallMethod calls a method by name on a model with the given arguments.
// Handles both pointer and value types.
// Returns the method results and an error if the method is not found or call fails.
func CallMethod(model any, methodName string, args ...any) ([]reflect.Value, error) {
	v := reflect.ValueOf(model)
	method := v.MethodByName(methodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("%w: %s", ErrMethodNotFound, methodName)
	}

	// Convert args to reflect.Value slice
	argValues := make([]reflect.Value, len(args))
	for i, arg := range args {
		argValues[i] = reflect.ValueOf(arg)
	}

	// Call the method
	results := method.Call(argValues)
	return results, nil
}
