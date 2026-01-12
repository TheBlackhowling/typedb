package typedb

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// Load loads a model by its primary key field.
// The model must have a field with load:"primary" tag.
// The model must have a QueryBy{Field}() method that returns the SQL query string.
// Updates the model in-place with data from the database.
//
// Example:
//
//	user := &User{ID: 123}
//	err := typedb.Load(ctx, db, user)
func Load[T ModelInterface](ctx context.Context, exec Executor, model T) error {
	// Find primary key field
	primaryField, found := findFieldByTag(model, "load", "primary")
	if !found {
		return fmt.Errorf("typedb: no field with load:\"primary\" tag found")
	}

	// Get field value
	fieldValueReflect, err := getFieldValue(model, primaryField.Name)
	if err != nil {
		return fmt.Errorf("typedb: failed to get primary key value: %w", err)
	}

	// Check if field is zero (not set)
	if fieldValueReflect.IsZero() {
		return fmt.Errorf("typedb: primary key field %s is not set", primaryField.Name)
	}

	fieldValue := fieldValueReflect.Interface()

	// Find and call QueryBy{Field}() method
	methodName := "QueryBy" + primaryField.Name
	_, methodFound := findMethod(model, methodName)
	if !methodFound {
		return fmt.Errorf("typedb: QueryBy%s() method not found", primaryField.Name)
	}

	// Call method to get query string
	methodValue := reflect.ValueOf(model).MethodByName(methodName)
	results := methodValue.Call(nil)
	if len(results) != 1 {
		return fmt.Errorf("typedb: QueryBy%s() should return exactly one value (string)", primaryField.Name)
	}
	query := results[0].String()

	// Execute query using QueryOne
	foundModel, err := QueryOne[T](ctx, exec, query, fieldValue)
	if err != nil {
		return err
	}

	// Update model in-place by copying values
	return updateModelInPlace(model, foundModel)
}

// LoadByField loads a model by any field with a load tag.
// The model must have a QueryBy{Field}() method that returns the SQL query string.
// Updates the model in-place with data from the database.
//
// Example:
//
//	user := &User{Email: "test@example.com"}
//	err := typedb.LoadByField(ctx, db, user, "Email")
func LoadByField[T ModelInterface](ctx context.Context, exec Executor, model T, fieldName string) error {
	// Get field value
	fieldValueReflect, err := getFieldValue(model, fieldName)
	if err != nil {
		return fmt.Errorf("typedb: failed to get field %s value: %w", fieldName, err)
	}

	// Check if field is zero (not set)
	if fieldValueReflect.IsZero() {
		return fmt.Errorf("typedb: field %s is not set", fieldName)
	}

	fieldValue := fieldValueReflect.Interface()

	// Find and call QueryBy{Field}() method
	methodName := "QueryBy" + fieldName
	_, methodFound := findMethod(model, methodName)
	if !methodFound {
		return fmt.Errorf("typedb: QueryBy%s() method not found", fieldName)
	}

	// Call method to get query string
	methodValue := reflect.ValueOf(model).MethodByName(methodName)
	results := methodValue.Call(nil)
	if len(results) != 1 {
		return fmt.Errorf("typedb: QueryBy%s() should return exactly one value (string)", fieldName)
	}
	query := results[0].String()

	// Execute query using QueryOne
	foundModel, err := QueryOne[T](ctx, exec, query, fieldValue)
	if err != nil {
		return err
	}

	// Update model in-place by copying values
	return updateModelInPlace(model, foundModel)
}

// LoadByComposite loads a model by a composite key.
// The model must have fields with load:"composite:name" tags.
// The model must have a QueryBy{Field1}{Field2}...() method (fields sorted alphabetically).
// Updates the model in-place with data from the database.
//
// Example:
//
//	userPost := &UserPost{UserID: 1, PostID: 2}
//	err := typedb.LoadByComposite(ctx, db, userPost, "userpost")
func LoadByComposite[T ModelInterface](ctx context.Context, exec Executor, model T, compositeName string) error {
	// Collect all fields with the composite tag
	t := getModelType(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var compositeFields []*reflect.StructField
	collectCompositeFields(t, compositeName, &compositeFields)

	if len(compositeFields) < 2 {
		return fmt.Errorf("typedb: composite key %q must have at least 2 fields", compositeName)
	}

	// Sort field names alphabetically (same as validation)
	fieldNames := make([]string, len(compositeFields))
	for i, f := range compositeFields {
		fieldNames[i] = f.Name
	}
	sort.Strings(fieldNames)

	// Get values for all fields in composite key
	fieldValues := make([]any, len(fieldNames))
	for i, fieldName := range fieldNames {
		valueReflect, err := getFieldValue(model, fieldName)
		if err != nil {
			return fmt.Errorf("typedb: failed to get field %s value: %w", fieldName, err)
		}

		// Check if field is zero (not set)
		if valueReflect.IsZero() {
			return fmt.Errorf("typedb: composite key field %s is not set", fieldName)
		}

		fieldValues[i] = valueReflect.Interface()
	}

	// Build method name: QueryBy{Field1}{Field2}...
	methodName := "QueryBy" + strings.Join(fieldNames, "")

	// Find and call QueryBy{Field1}{Field2}...() method
	_, methodFound := findMethod(model, methodName)
	if !methodFound {
		return fmt.Errorf("typedb: %s() method not found", methodName)
	}

	// Call method to get query string
	methodValue := reflect.ValueOf(model).MethodByName(methodName)
	results := methodValue.Call(nil)
	if len(results) != 1 {
		return fmt.Errorf("typedb: %s() should return exactly one value (string)", methodName)
	}
	query := results[0].String()

	// Execute query using QueryOne with all field values as arguments
	foundModel, err := QueryOne[T](ctx, exec, query, fieldValues...)
	if err != nil {
		return err
	}

	// Update model in-place by copying values
	return updateModelInPlace(model, foundModel)
}

// collectCompositeFields recursively collects fields with the specified composite tag.
func collectCompositeFields(t reflect.Type, compositeName string, fields *[]*reflect.StructField) {
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
		if loadTag != "" {
			// Check if this field has the composite tag
			tagParts := splitTag(loadTag)
			for _, part := range tagParts {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "composite:") {
					tagCompositeName := strings.TrimPrefix(part, "composite:")
					if tagCompositeName == compositeName {
						*fields = append(*fields, &field)
						break
					}
				}
			}
		}

		// Check embedded structs
		if field.Anonymous {
			embeddedType := field.Type
			if embeddedType.Kind() == reflect.Ptr {
				embeddedType = embeddedType.Elem()
			}
			collectCompositeFields(embeddedType, compositeName, fields)
		}
	}
}

// updateModelInPlace updates the destination model with values from the source model.
// Uses reflection to copy all exported field values.
func updateModelInPlace(dest, src any) error {
	destValue := reflect.ValueOf(dest)
	srcValue := reflect.ValueOf(src)

	if destValue.Kind() != reflect.Ptr || srcValue.Kind() != reflect.Ptr {
		return fmt.Errorf("typedb: both models must be pointers")
	}

	destValue = destValue.Elem()
	srcValue = srcValue.Elem()

	if destValue.Kind() != reflect.Struct || srcValue.Kind() != reflect.Struct {
		return fmt.Errorf("typedb: both models must be pointers to structs")
	}

	// Copy all exported fields
	for i := 0; i < destValue.NumField(); i++ {
		destField := destValue.Field(i)
		if !destField.CanSet() {
			continue
		}

		// Find corresponding field in source by name
		destType := destValue.Type()
		fieldName := destType.Field(i).Name
		srcField := srcValue.FieldByName(fieldName)

		if srcField.IsValid() && srcField.CanInterface() {
			if destField.Type() == srcField.Type() {
				destField.Set(srcField)
			} else if destField.Type().AssignableTo(srcField.Type()) {
				destField.Set(srcField)
			} else if srcField.Type().AssignableTo(destField.Type()) {
				destField.Set(srcField.Convert(destField.Type()))
			}
		}
	}

	return nil
}
