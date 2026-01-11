package typedb

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// ValidationError represents a validation error for a specific model.
type ValidationError struct {
	ModelName string
	Errors    []string
}

// Error implements the error interface.
func (ve *ValidationError) Error() string {
	if len(ve.Errors) == 0 {
		return fmt.Sprintf("typedb: validation failed for model %s", ve.ModelName)
	}
	return fmt.Sprintf("typedb: validation failed for model %s:\n  %s", ve.ModelName, strings.Join(ve.Errors, "\n  "))
}

// ValidationErrors represents multiple validation errors, grouped by model.
type ValidationErrors struct {
	Errors []*ValidationError
}

// Error implements the error interface.
func (ve *ValidationErrors) Error() string {
	if len(ve.Errors) == 0 {
		return "typedb: validation failed"
	}
	var parts []string
	for _, err := range ve.Errors {
		parts = append(parts, err.Error())
	}
	return strings.Join(parts, "\n\n")
}

// ValidateModel validates a single model instance.
// Checks that:
// - Only one field has load:"primary" tag
// - Fields with load:"primary" must have QueryBy{Field}() method
// - Fields with load:"unique" must have QueryBy{Field}() method
// - Fields with load:"composite:name" must have QueryBy{Field1}{Field2}...() method (fields sorted alphabetically)
// - Query methods return string
func ValidateModel[T ModelInterface](model T) error {
	t := reflect.TypeOf(model)
	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("typedb: ValidateModel requires a pointer type")
	}
	t = t.Elem()

	var errors []string

	// Find all fields with load tags
	var primaryFields []*reflect.StructField
	uniqueFields := make(map[string]*reflect.StructField)
	compositeGroups := make(map[string][]*reflect.StructField)

	collectLoadFields(t, &primaryFields, uniqueFields, compositeGroups)

	// Validate primary key
	if len(primaryFields) > 1 {
		fieldNames := make([]string, len(primaryFields))
		for i, f := range primaryFields {
			fieldNames[i] = f.Name
		}
		errors = append(errors, fmt.Sprintf("multiple primary key fields found: %s (only one allowed)", strings.Join(fieldNames, ", ")))
	} else if len(primaryFields) == 1 {
		field := primaryFields[0]
		methodName := "QueryBy" + field.Name
		if err := validateQueryMethod(model, methodName); err != nil {
			errors = append(errors, fmt.Sprintf("primary key field %s: %v", field.Name, err))
		}
	}

	// Validate unique fields
	for fieldName := range uniqueFields {
		methodName := "QueryBy" + fieldName
		if err := validateQueryMethod(model, methodName); err != nil {
			errors = append(errors, fmt.Sprintf("unique field %s: %v", fieldName, err))
		}
	}

	// Validate composite keys
	for compositeName, fields := range compositeGroups {
		if len(fields) < 2 {
			fieldNames := make([]string, len(fields))
			for i, f := range fields {
				fieldNames[i] = f.Name
			}
			errors = append(errors, fmt.Sprintf("composite key %q has only %d field(s): %s (at least 2 required)", compositeName, len(fields), strings.Join(fieldNames, ", ")))
			continue
		}

		// Sort field names alphabetically
		fieldNames := make([]string, len(fields))
		for i, f := range fields {
			fieldNames[i] = f.Name
		}
		sort.Strings(fieldNames)

		// Build method name: QueryBy{Field1}{Field2}...
		methodName := "QueryBy" + strings.Join(fieldNames, "")
		if err := validateQueryMethod(model, methodName); err != nil {
			errors = append(errors, fmt.Sprintf("composite key %q: %v", compositeName, err))
		}
	}

	if len(errors) > 0 {
		return &ValidationError{
			ModelName: t.Name(),
			Errors:    errors,
		}
	}

	return nil
}

// collectLoadFields recursively collects all fields with load tags.
func collectLoadFields(t reflect.Type, primaryFields *[]*reflect.StructField, uniqueFields map[string]*reflect.StructField, compositeGroups map[string][]*reflect.StructField) {
	if t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		loadTag := field.Tag.Get("load")
		if loadTag == "" {
			// Check embedded structs
			if field.Anonymous {
				embeddedType := field.Type
				if embeddedType.Kind() == reflect.Ptr {
					embeddedType = embeddedType.Elem()
				}
				collectLoadFields(embeddedType, primaryFields, uniqueFields, compositeGroups)
			}
			continue
		}

		// Parse load tag
		tagParts := splitTag(loadTag)
		for _, part := range tagParts {
			part = strings.TrimSpace(part)

			if part == "primary" {
				*primaryFields = append(*primaryFields, &field)
			} else if part == "unique" {
				uniqueFields[field.Name] = &field
			} else if strings.HasPrefix(part, "composite:") {
				compositeName := strings.TrimPrefix(part, "composite:")
				if compositeName == "" {
					continue
				}
				compositeGroups[compositeName] = append(compositeGroups[compositeName], &field)
			}
		}
	}
}

// validateQueryMethod validates that a query method exists and has the correct signature.
// Expected signature: func() string
func validateQueryMethod(model any, methodName string) error {
	method, found := FindMethod(model, methodName)
	if !found {
		return fmt.Errorf("QueryBy%s() method not found", strings.TrimPrefix(methodName, "QueryBy"))
	}

	// Check method signature: should be func() string
	methodType := method.Type

	// Should have no input parameters (except receiver)
	if methodType.NumIn() != 1 {
		return fmt.Errorf("QueryBy%s() method should have no parameters (has %d)", strings.TrimPrefix(methodName, "QueryBy"), methodType.NumIn()-1)
	}

	// Should return exactly one value (string)
	if methodType.NumOut() != 1 {
		return fmt.Errorf("QueryBy%s() method should return exactly one value (has %d)", strings.TrimPrefix(methodName, "QueryBy"), methodType.NumOut())
	}

	// Return type should be string
	returnType := methodType.Out(0)
	if returnType.Kind() != reflect.String {
		return fmt.Errorf("QueryBy%s() method should return string (returns %v)", strings.TrimPrefix(methodName, "QueryBy"), returnType)
	}

	return nil
}

// ValidateAllRegistered validates all registered models.
// Returns ValidationErrors containing all validation failures grouped by model.
func ValidateAllRegistered() error {
	models := GetRegisteredModels()
	if len(models) == 0 {
		return nil
	}

	var validationErrors []*ValidationError

	for _, modelType := range models {
		// Create a zero value instance of the model
		modelPtr := reflect.New(modelType).Interface()

		// Cast to ModelInterface
		model, ok := modelPtr.(ModelInterface)
		if !ok {
			validationErrors = append(validationErrors, &ValidationError{
				ModelName: modelType.Name(),
				Errors:    []string{fmt.Sprintf("model does not implement ModelInterface")},
			})
			continue
		}

		// Validate the model
		if err := ValidateModel(model); err != nil {
			if ve, ok := err.(*ValidationError); ok {
				validationErrors = append(validationErrors, ve)
			} else {
				validationErrors = append(validationErrors, &ValidationError{
					ModelName: modelType.Name(),
					Errors:    []string{err.Error()},
				})
			}
		}
	}

	if len(validationErrors) > 0 {
		return &ValidationErrors{
			Errors: validationErrors,
		}
	}

	return nil
}

// MustValidateAllRegistered validates all registered models and panics if validation fails.
// This is called automatically by Open() to ensure all models are valid before use.
func MustValidateAllRegistered() {
	if err := ValidateAllRegistered(); err != nil {
		panic(err)
	}
}
