package typedb

import (
	"fmt"
	"reflect"
	"sync"
)

// ModelOptions configures behavior for a registered model.
type ModelOptions struct {
	// PartialUpdate enables tracking of original model state after deserialization.
	// When enabled, Update() will only update fields that have changed since the last deserialization.
	// This requires keeping a copy of the deserialized object, which uses additional memory.
	PartialUpdate bool
}

var (
	registeredModels []reflect.Type
	modelOptions     map[reflect.Type]ModelOptions // Maps model type to its options
	registerMutex    sync.RWMutex
	validationOnce   sync.Once
	validationError  error
)

func init() {
	modelOptions = make(map[reflect.Type]ModelOptions)
}

// RegisterModel registers a model type for validation.
// Requires a pointer type (e.g., RegisterModel[*User]()).
// Models should call this function in their init() functions.
// Panics if validation fails, as init() functions cannot return errors.
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
		panic(fmt.Errorf("typedb: RegisterModel requires a pointer type (e.g., RegisterModel[*User]())"))
	}
	t = t.Elem()

	// Validate the model BEFORE adding to registry to prevent invalid models from being registered
	// Use the zero value instance created at the start of the function
	if err := ValidateModel(model); err != nil {
		// Log the error for visibility before panicking
		defaultLogger.Error("Model registration validation failed", "model", t.Name(), "error", err)
		panic(fmt.Errorf("typedb: validation failed for model %s during registration: %w", t.Name(), err))
	}

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

// RegisterModelWithOptions registers a model type with options for validation and behavior configuration.
// Requires a pointer type (e.g., RegisterModelWithOptions[*User](ModelOptions{PartialUpdate: true})).
// Models should call this function in their init() functions when they need custom behavior.
// Panics if validation fails, as init() functions cannot return errors.
//
// Example:
//
//	type User struct {
//	    typedb.Model
//	    ID int `db:"id" load:"primary"`
//	}
//
//	func init() {
//	    typedb.RegisterModelWithOptions[*User](typedb.ModelOptions{PartialUpdate: true})
//	}
func RegisterModelWithOptions[T ModelInterface](opts ModelOptions) {
	var model T
	t := reflect.TypeOf(model)
	if t.Kind() != reflect.Ptr {
		panic(fmt.Errorf("typedb: RegisterModelWithOptions requires a pointer type (e.g., RegisterModelWithOptions[*User](...))"))
	}
	t = t.Elem()

	// Validate the model BEFORE adding to registry to prevent invalid models from being registered
	// Use the zero value instance created at the start of the function
	if err := ValidateModel(model); err != nil {
		// Log the error for visibility before panicking
		defaultLogger.Error("Model registration validation failed", "model", t.Name(), "error", err)
		panic(fmt.Errorf("typedb: validation failed for model %s during registration: %w", t.Name(), err))
	}

	registerMutex.Lock()
	defer registerMutex.Unlock()

	// Register the model type if not already registered
	alreadyRegistered := false
	for _, registered := range registeredModels {
		if registered == t {
			alreadyRegistered = true
			break
		}
	}
	if !alreadyRegistered {
		registeredModels = append(registeredModels, t)
	}

	// Store options for this model type
	modelOptions[t] = opts
}

// GetModelOptions returns the options for a registered model type.
// Returns zero-value ModelOptions if the model is not registered or has no options.
func GetModelOptions(modelType reflect.Type) ModelOptions {
	registerMutex.RLock()
	defer registerMutex.RUnlock()

	if opts, ok := modelOptions[modelType]; ok {
		return opts
	}
	return ModelOptions{}
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

// ClearRegisteredModels clears all registered models.
// This is intended for testing purposes only and should not be used in production code.
// Models should be registered in init() functions and remain registered for the lifetime of the program.
func ClearRegisteredModels() {
	registerMutex.Lock()
	defer registerMutex.Unlock()

	registeredModels = nil
	modelOptions = make(map[reflect.Type]ModelOptions)
}
