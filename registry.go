package typedb

import (
	"reflect"
	"sync"
)

var (
	registeredModels []reflect.Type
	registerMutex    sync.RWMutex
	validationOnce   sync.Once
	validationError  error
)

// RegisterModel registers a model type for validation.
// Requires a pointer type (e.g., RegisterModel[*User]()).
// Models should call this function in their init() functions.
//
// Example:
//
//	type User struct {
//	    typedb.Model
//	    ID int `db:"id" load:"primary"`
//	}
//
//	func init() {
//	    typedb.RegisterModel[*User]()
//	}
func RegisterModel[T ModelInterface]() {
	var model T
	t := reflect.TypeOf(model)
	if t.Kind() != reflect.Ptr {
		panic("typedb: RegisterModel requires a pointer type (e.g., RegisterModel[*User]())")
	}
	t = t.Elem()

	registerMutex.Lock()
	defer registerMutex.Unlock()

	// Avoid duplicates
	for _, registered := range registeredModels {
		if registered == t {
			return
		}
	}

	registeredModels = append(registeredModels, t)
}

// GetRegisteredModels returns all registered model types.
// Used by the validation system to validate all models.
// Returns a copy of the registered models slice to prevent external modification.
func GetRegisteredModels() []reflect.Type {
	registerMutex.RLock()
	defer registerMutex.RUnlock()

	// Return a copy to prevent external modification
	result := make([]reflect.Type, len(registeredModels))
	copy(result, registeredModels)
	return result
}
