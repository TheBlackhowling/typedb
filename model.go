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
	modelValue := reflect.ValueOf(m)
	if modelValue.Kind() != reflect.Ptr {
		return fmt.Errorf("typedb: Model.Deserialize() called on non-pointer type")
	}

	outerPtr := unsafe.Pointer(m) // #nosec G103 // intentional use of unsafe for reflection

	registeredModels := GetRegisteredModels()
	for _, structType := range registeredModels {
		if structType.Kind() != reflect.Struct {
			continue
		}

		if structType.NumField() > 0 {
			firstField := structType.Field(0)
			if firstField.Anonymous && firstField.Type == reflect.TypeOf(Model{}) {
				outerStructPtr := newAtUnsafe(structType, outerPtr)
				if outerModelInterface := outerStructPtr.Interface(); outerModelInterface != nil {
					if outerModel, ok := outerModelInterface.(ModelInterface); ok {
						return deserialize(row, outerModel)
					}
				}
			}
		}
	}

	return fmt.Errorf("typedb: Model.Deserialize() cannot determine outer struct type - ensure Model is embedded as the first field and the model is registered")
}
