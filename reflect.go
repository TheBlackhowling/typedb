package typedb

import (
	"fmt"
	"reflect"
	"strings"
)

// getModelType returns the reflect.Type of a model.
// Requires a pointer type and returns the underlying struct type.
// Panics if model is not a pointer.
func getModelType(model any) reflect.Type {
	t := reflect.TypeOf(model)
	if t.Kind() != reflect.Ptr {
		panic("typedb: GetModelType requires a pointer type")
	}
	return t.Elem()
}

// findFieldByTag finds a struct field by its tag value.
// Searches through embedded structs (like Model base).
// Requires a pointer type.
// Returns the field and true if found, nil and false otherwise.
// tagKey specifies which struct tag key to search (e.g., "load", "nolog").
func findFieldByTag(model any, tagKey, tagValue string) (*reflect.StructField, bool) {
	t := getModelType(model)
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

// getFieldValue gets the value of a field by name from a model.
// Requires a pointer type.
// Returns an error if the field is not found or cannot be accessed.
func getFieldValue(model any, fieldName string) (reflect.Value, error) {
	v := reflect.ValueOf(model)
	if v.Kind() != reflect.Ptr {
		return reflect.Value{}, fmt.Errorf("typedb: model must be a pointer type")
	}

	if v.IsNil() {
		return reflect.Value{}, fmt.Errorf("typedb: cannot get field value from nil pointer")
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("typedb: model must be a pointer to struct")
	}

	field, found := findFieldByNameRecursive(v.Type(), fieldName)
	if !found {
		return reflect.Value{}, fmt.Errorf("%w: %s", ErrFieldNotFound, fieldName)
	}

	return v.FieldByIndex(field.Index), nil
}

// setFieldValue sets the value of a field by name in a model.
// Requires a pointer type.
// Returns an error if the field is not found, cannot be accessed, or value type is incompatible.
func setFieldValue(model any, fieldName string, value any) error {
	v := reflect.ValueOf(model)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("typedb: model must be a pointer type")
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

	// Handle type conversion for numeric types
	if valueV.Type().AssignableTo(fieldValue.Type()) {
		fieldValue.Set(valueV)
	} else if valueV.Kind() == reflect.Int64 && fieldValue.Kind() == reflect.Int {
		// Convert int64 to int
		fieldValue.SetInt(valueV.Int())
	} else if valueV.Kind() == reflect.Int && fieldValue.Kind() == reflect.Int64 {
		// Convert int to int64
		fieldValue.SetInt(valueV.Int())
	} else if valueV.Kind() == reflect.Float64 && fieldValue.Kind() == reflect.Float32 {
		// Convert float64 to float32
		fieldValue.SetFloat(valueV.Float())
	} else if valueV.Kind() == reflect.Float32 && fieldValue.Kind() == reflect.Float64 {
		// Convert float32 to float64
		fieldValue.SetFloat(valueV.Float())
	} else {
		return fmt.Errorf("typedb: cannot assign %v to field %s of type %v", valueV.Type(), fieldName, fieldValue.Type())
	}

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
				adjustedIndex := make([]int, len(field.Index), len(field.Index)+len(foundField.Index))
				copy(adjustedIndex, field.Index)
				adjustedIndex = append(adjustedIndex, foundField.Index...)
				adjustedField.Index = adjustedIndex
				return &adjustedField, true
			}
		}
	}

	return nil, false
}

// findMethod finds a method by name on a model.
// Requires a pointer type (methods are typically defined on pointer receivers).
// Returns the method and true if found, nil and false otherwise.
func findMethod(model any, methodName string) (*reflect.Method, bool) {
	t := reflect.TypeOf(model)
	method, found := t.MethodByName(methodName)
	if !found {
		return nil, false
	}
	return &method, true
}

// callMethod calls a method by name on a model with the given arguments.
// Requires a pointer type (methods are typically defined on pointer receivers).
// Returns the method results and an error if the method is not found or call fails.
func callMethod(model any, methodName string, args ...any) ([]reflect.Value, error) {
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
