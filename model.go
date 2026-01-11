package typedb

import (
	"fmt"
	"reflect"
	"unsafe"
)

// Deserialize deserializes a row into the model.
// Delegates to the standard Deserialize function.
// When called on an embedded Model, it converts the *Model receiver to the outer struct pointer.
//
// Example:
//
//	row := map[string]any{"id": 1, "name": "Alice"}
//	err := user.Deserialize(row)
func (m *Model) Deserialize(row map[string]any) error {
	// When Model.Deserialize() is called on an embedded Model (e.g., user.Deserialize() where user is *User),
	// the receiver m is *Model (pointing to the embedded Model field).
	// Since Model is typically the first embedded field, the outer struct pointer is at the same memory address.

	modelValue := reflect.ValueOf(m)
	if modelValue.Kind() != reflect.Ptr {
		return fmt.Errorf("typedb: Model.Deserialize() called on non-pointer type")
	}

	// The outer struct pointer has the same address as the Model pointer
	// (since Model is embedded as the first field)
	outerPtr := unsafe.Pointer(m)

	// Try to find the outer struct type from registered models
	registeredModels := GetRegisteredModels()
	for _, structType := range registeredModels {
		// structType is already the struct type (not pointer)
		if structType.Kind() != reflect.Struct {
			continue
		}

		// Check if first field is embedded Model
		if structType.NumField() > 0 {
			firstField := structType.Field(0)
			if firstField.Anonymous && firstField.Type == reflect.TypeOf(Model{}) {
				// This model embeds Model as first field
				// Create a pointer to this type at the same address
				outerStructPtr := reflect.NewAt(structType, outerPtr)
				if outerModel, ok := outerStructPtr.Interface().(ModelInterface); ok {
					return Deserialize(row, outerModel)
				}
			}
		}
	}

	return fmt.Errorf("typedb: Model.Deserialize() cannot determine outer struct type - ensure Model is embedded as the first field and the model is registered")
}
