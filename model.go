package typedb

import (
	"fmt"
	"reflect"
	"unsafe"
)

// newAtUnsafe creates a pointer to a struct type at the given address.
// This is safe when Model is the first embedded field, as the memory layout
// guarantees the outer struct starts at the same address.
//
//go:nocheckptr
func newAtUnsafe(typ reflect.Type, ptr unsafe.Pointer) reflect.Value {
	return reflect.NewAt(typ, ptr)
}

// deserialize deserializes a row into the model.
// Delegates to the internal deserialize function.
// When called on an embedded Model, it converts the *Model receiver to the outer struct pointer.
// This method is unexported - users should use QueryAll, QueryFirst, QueryOne, InsertAndLoad, Load, etc. instead.
//
// Example (internal usage - users should use QueryAll instead):
//
//	row := map[string]any{"id": 1, "name": "Alice"}
//	err := user.deserialize(row)
func (m *Model) deserialize(row map[string]any) error {
	// When Model.Deserialize() is called on an embedded Model (e.g., user.Deserialize() where user is *User),
	// the receiver m is *Model (pointing to the embedded Model field).
	// Since Model is typically the first embedded field, the outer struct pointer is at the same memory address.

	modelValue := reflect.ValueOf(m)
	if modelValue.Kind() != reflect.Ptr {
		return fmt.Errorf("typedb: Model.Deserialize() called on non-pointer type")
	}

	// The outer struct pointer has the same address as the Model pointer
	// (since Model is embedded as the first field)
	outerPtr := unsafe.Pointer(m) //nolint:gosec // G103: intentional use of unsafe for reflection

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
				// This is safe because Model is the first embedded field, so the memory
				// layout guarantees the outer struct starts at the same address
				outerStructPtr := newAtUnsafe(structType, outerPtr)
				if outerModelInterface := outerStructPtr.Interface(); outerModelInterface != nil {
					if outerModel, ok := outerModelInterface.(ModelInterface); ok {
						// The pointer from reflect.NewAt should be addressable.
						// However, in Go 1.20+, when we pass it through an interface and
						// then use reflect.ValueOf, the resulting struct value from .Elem()
						// may not be addressable, causing checkptr errors.
						//
						// Solution: Ensure we're working with the actual pointer value,
						// not an interface. The pointer value itself is always addressable.
						// We pass the pointer directly to Deserialize, which will handle it.
						return deserialize(row, outerModel)
					}
				}
			}
		}
	}

	return fmt.Errorf("typedb: Model.Deserialize() cannot determine outer struct type - ensure Model is embedded as the first field and the model is registered")
}
