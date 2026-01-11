package typedb

import (
	"fmt"
	"reflect"
)

// Deserialize deserializes a row into the model.
// Delegates to the standard Deserialize function.
//
// Example:
//
//	row := map[string]any{"id": 1, "name": "Alice"}
//	err := user.Deserialize(row)
func (m *Model) Deserialize(row map[string]any) error {
	// Get the concrete model type
	modelValue := reflect.ValueOf(m)
	if modelValue.Kind() != reflect.Ptr {
		return fmt.Errorf("typedb: Model.Deserialize() called on non-pointer type")
	}

	// Use the standard Deserialize function
	return Deserialize(row, m)
}
