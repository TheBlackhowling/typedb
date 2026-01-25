package typedb

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

// Update updates a model in the database by automatically building the UPDATE query.
// The model must:
//   - Implement TableName() method that returns the table name
//   - Have a field with load:"primary" tag (for WHERE clause)
//   - Have the primary key field set (non-zero value)
//   - Not have dot notation in db tags (simple model, not joined)
//
// Nil/zero value fields are excluded from the UPDATE.
// Only fields that are set (non-zero/non-nil) will be updated.
//
// Fields with dbUpdate:"auto-timestamp" tag are automatically populated with database timestamp functions
// (e.g., CURRENT_TIMESTAMP, NOW(), GETDATE()) and do not need to be set in the model.
//
// Partial Update:
// When a model is registered with RegisterModelWithOptions(ModelOptions{PartialUpdate: true}),
// Update() will only update fields that have changed since the last deserialization (Load, Query, etc.).
// This requires keeping a copy of the deserialized object, which uses additional memory.
// The original copy is automatically saved after deserialization and refreshed after successful updates.
//
// Example (standard update):
//
//	type User struct {
//	    Model
//	    ID        int    `db:"id" load:"primary"`
//	    Name      string `db:"name"`
//	    Email     string `db:"email"`
//	    UpdatedAt string `db:"updated_at" dbUpdate:"auto-timestamp"`
//	}
//
//	func (u *User) TableName() string {
//	    return "users"
//	}
//
//	user := &User{ID: 123, Name: "John Updated"}
//	err := typedb.Update(ctx, db, user)
//	// Generates: UPDATE users SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2
//
// Example (partial update):
//
//	func init() {
//	    typedb.RegisterModelWithOptions[*User](typedb.ModelOptions{PartialUpdate: true})
//	}
//
//	// Load user
//	user := &User{ID: 123}
//	typedb.Load(ctx, db, user) // Original copy saved automatically
//
//	// Modify only name
//	user.Name = "New Name"
//	typedb.Update(ctx, db, user) // Only updates name field, not email
// validateUpdateModel validates model for update operation
func validateUpdateModel[T ModelInterface](model T) (string, *reflect.StructField, string, error) {
	tableName, err := getTableName(model)
	if err != nil {
		return "", nil, "", fmt.Errorf("typedb: Update validation failed: %w", err)
	}

	if hasDotNotation(model) {
		return "", nil, "", fmt.Errorf("typedb: Update cannot be used with joined models (detected dot notation in db tags)")
	}

	primaryField, found := findFieldByTag(model, "load", "primary")
	if !found {
		return "", nil, "", fmt.Errorf("typedb: Update requires a field with load:\"primary\" tag")
	}

	primaryKeyColumn := primaryField.Tag.Get("db")
	if primaryKeyColumn == "" || primaryKeyColumn == "-" {
		return "", nil, "", fmt.Errorf("typedb: primary key field %s must have a db tag", primaryField.Name)
	}

	// Extract column name (handle dot notation - use last part)
	if strings.Contains(primaryKeyColumn, ".") {
		parts := strings.Split(primaryKeyColumn, ".")
		primaryKeyColumn = parts[len(parts)-1]
	}

	return tableName, primaryField, primaryKeyColumn, nil
}

// buildUpdateQuery builds the UPDATE query with SET clause
func buildUpdateQuery(driverName string, tableName string, primaryKeyColumn string,
	columns []string, values []any, autoUpdateColumns []string) (string, []any) {
	quotedTableName := quoteIdentifier(driverName, tableName)
	quotedPrimaryKeyColumn := quoteIdentifier(driverName, primaryKeyColumn)

	var setClauses []string
	placeholderIndex := 1

	// Add regular fields with placeholders
	for _, col := range columns {
		quotedCol := quoteIdentifier(driverName, col)
		placeholder := generatePlaceholder(driverName, placeholderIndex)
		setClauses = append(setClauses, fmt.Sprintf("%s = %s", quotedCol, placeholder))
		placeholderIndex++
	}

	// Add auto-update timestamp fields with database functions
	for _, col := range autoUpdateColumns {
		quotedCol := quoteIdentifier(driverName, col)
		timestampFunc := getTimestampFunction(driverName)
		setClauses = append(setClauses, fmt.Sprintf("%s = %s", quotedCol, timestampFunc))
	}

	wherePlaceholder := generatePlaceholder(driverName, len(values)+1)
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s = %s",
		quotedTableName,
		strings.Join(setClauses, ", "),
		quotedPrimaryKeyColumn,
		wherePlaceholder)

	allValues := append(values, nil) // Placeholder for primary key value
	return query, allValues
}

func Update[T ModelInterface](ctx context.Context, exec Executor, model T) error {
	// Validation
	tableName, primaryField, primaryKeyColumn, err := validateUpdateModel(model)
	if err != nil {
		return err
	}

	// Get primary key value
	primaryKeyValue, err := getFieldValue(model, primaryField.Name)
	if err != nil {
		return fmt.Errorf("typedb: Update failed to get primary key value: %w", err)
	}

	if isZeroOrNil(primaryKeyValue) {
		return fmt.Errorf("typedb: Update requires primary key field %s to be set (non-zero value)", primaryField.Name)
	}

	// Get changed fields if partial update enabled
	driverName := getDriverName(exec)
	structType := reflect.TypeOf(model).Elem()
	opts := GetModelOptions(structType)
	var changedFields map[string]bool
	if opts.PartialUpdate {
		changedFields, err = getChangedFields(model, primaryField.Name)
		if err != nil {
			return fmt.Errorf("typedb: Update failed to get changed fields: %w", err)
		}
	}

	// Serialize fields
	columns, values, autoUpdateColumns, maskIndices, err := serializeModelFieldsForUpdate(model, primaryField.Name, driverName, changedFields)
	if err != nil {
		return fmt.Errorf("typedb: Update failed to serialize model: %w", err)
	}

	if len(columns) == 0 && len(autoUpdateColumns) == 0 {
		return fmt.Errorf("typedb: Update requires at least one non-nil field to update")
	}

	// Store mask indices
	if len(maskIndices) > 0 {
		ctx = WithMaskIndices(ctx, maskIndices)
	}

	// Build query
	query, allValues := buildUpdateQuery(driverName, tableName, primaryKeyColumn, columns, values, autoUpdateColumns)
	allValues[len(allValues)-1] = primaryKeyValue.Interface()

	// Execute
	_, err = exec.Exec(ctx, query, allValues...)
	if err != nil {
		return fmt.Errorf("typedb: Update failed: %w", err)
	}

	// If partial update is enabled, refresh the original copy after successful update
	if opts.PartialUpdate {
		if err := saveOriginalCopyIfEnabled(model); err != nil {
			// Log error but don't fail the update - the update succeeded
			// The original copy will be refreshed on the next deserialization
		}
	}

	return nil
}

// getTimestampFunction returns the database-specific function for getting the current timestamp.
func getTimestampFunction(driverName string) string {
	driverName = strings.ToLower(driverName)
	switch driverName {
	case "postgres", "sqlite3":
		return "CURRENT_TIMESTAMP"
	case "mysql":
		return "NOW()"
	case "sqlserver", "mssql":
		return "GETDATE()"
	case "oracle":
		return "CURRENT_TIMESTAMP"
	default:
		// Default to SQL standard
		return "CURRENT_TIMESTAMP"
	}
}

// serializeModelFieldsForUpdate collects non-nil/non-zero fields from a model for UPDATE operations.
// Excludes primary key field, fields with db:"-" tag, and fields with dbUpdate:"false" tag.
// Fields with db:"-" are excluded from all database operations (INSERT, UPDATE, SELECT).
// Fields with dbUpdate:"false" are excluded from UPDATE but can still be used in INSERT and SELECT.
// Fields with dbUpdate:"auto-timestamp" are automatically populated with database timestamp functions.
// If changedFields is provided (partial update enabled), only includes fields that have changed.
// Returns: column names, field values for serialization, auto-update column names, and mask indices (for fields with nolog:"true" tag).
func serializeModelFieldsForUpdate(model ModelInterface, primaryKeyFieldName string, driverName string, changedFields map[string]bool) ([]string, []any, []string, []int, error) {
	modelValue := reflect.ValueOf(model)
	if modelValue.Kind() != reflect.Ptr || modelValue.IsNil() {
		return nil, nil, nil, nil, fmt.Errorf("typedb: model must be a non-nil pointer")
	}

	modelValue = modelValue.Elem()
	if modelValue.Kind() != reflect.Struct {
		return nil, nil, nil, nil, fmt.Errorf("typedb: model must be a pointer to struct")
	}

	var columns []string
	var values []any
	var autoUpdateColumns []string
	var maskIndices []int

	iterateStructFields(modelValue.Type(), modelValue, primaryKeyFieldName, func(field reflect.StructField, fieldValue reflect.Value, columnName string) bool {
		// Check for dbUpdate tag
		dbUpdateTag := field.Tag.Get("dbUpdate")

		// Skip fields with dbUpdate:"false" tag
		if dbUpdateTag == "false" {
			return true
		}

		// Handle fields with dbUpdate:"auto-timestamp" tag - use database function
		if dbUpdateTag == "auto-timestamp" {
			// Include auto-timestamp fields if partial update is disabled, or if field has changed
			if changedFields == nil || changedFields[columnName] {
				autoUpdateColumns = append(autoUpdateColumns, columnName)
			}
			return true
		}

		// If partial update is enabled, only include changed fields
		if changedFields != nil {
			if !changedFields[columnName] {
				return true
			}
		}

		// Skip nil/zero values for regular fields (always exclude zero values)
		if isZeroOrNil(fieldValue) {
			return true
		}

		// Track if this field should be masked in logs
		shouldMask := field.Tag.Get("nolog") == "true"
		if shouldMask {
			maskIndices = append(maskIndices, len(values))
		}

		columns = append(columns, columnName)
		values = append(values, fieldValue.Interface())
		return true
	})

	return columns, values, autoUpdateColumns, maskIndices, nil
}

// extractOriginalCopy extracts the original copy from Model.originalCopy field using unsafe
func extractOriginalCopy(structValue reflect.Value) (interface{}, error) {
	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Type().Field(i)
		if field.Anonymous && field.Type == reflect.TypeOf(Model{}) {
			// Use unsafe to access unexported field
			modelFieldValue := structValue.Field(i)
			modelFieldPtr := unsafe.Pointer(modelFieldValue.UnsafeAddr())
			originalCopyFieldType := field.Type.Field(0) // Model.originalCopy field
			originalCopyFieldPtr := unsafe.Add(modelFieldPtr, originalCopyFieldType.Offset)
			originalCopy := *(*interface{})(originalCopyFieldPtr)
			if originalCopy != nil {
				return originalCopy, nil
			}
		}
	}
	return nil, nil // No original copy found
}

// compareFieldMaps compares two field maps and returns a map of changed column names
func compareFieldMaps(currentFields, originalFields map[string]reflect.Value) map[string]bool {
	changedFields := make(map[string]bool)

	// Check fields in current model
	for columnName, currentFieldValue := range currentFields {
		originalFieldValue, exists := originalFields[columnName]
		if !exists {
			// Field exists in current but not in original - consider it changed
			changedFields[columnName] = true
			continue
		}

		// Compare field values
		currentVal := currentFieldValue.Interface()
		originalVal := originalFieldValue.Interface()

		// Use DeepEqual for comparison
		if !reflect.DeepEqual(currentVal, originalVal) {
			changedFields[columnName] = true
		}
	}

	// Check for fields in original but not in current (shouldn't happen, but be safe)
	for columnName := range originalFields {
		if _, exists := currentFields[columnName]; !exists {
			changedFields[columnName] = true
		}
	}

	return changedFields
}

// getChangedFields compares the current model state with its original copy and returns
// a map of column names that have changed. Returns nil if partial update is not enabled
// or if no original copy exists.
func getChangedFields(model ModelInterface, primaryKeyFieldName string) (map[string]bool, error) {
	modelValue := reflect.ValueOf(model)
	if modelValue.Kind() != reflect.Ptr || modelValue.IsNil() {
		return nil, fmt.Errorf("model must be a non-nil pointer")
	}

	structValue := modelValue.Elem()
	if structValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("model must be a pointer to struct")
	}

	// Extract original copy
	originalCopy, err := extractOriginalCopy(structValue)
	if err != nil {
		return nil, err
	}

	if originalCopy == nil {
		// No original copy exists - treat all fields as changed (fallback to normal update)
		return nil, nil
	}

	// Validate original copy
	originalValue := reflect.ValueOf(originalCopy)
	if originalValue.Kind() != reflect.Ptr || originalValue.IsNil() {
		return nil, fmt.Errorf("original copy must be a non-nil pointer")
	}

	originalStructValue := originalValue.Elem()
	if originalStructValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("original copy must be a pointer to struct")
	}

	// Build field maps for comparison
	currentFields := buildFieldMapForComparison(structValue, primaryKeyFieldName)
	originalFields := buildFieldMapForComparison(originalStructValue, primaryKeyFieldName)

	// Compare and return changed fields
	return compareFieldMaps(currentFields, originalFields), nil
}

// buildFieldMapForComparison builds a map of column names to field values for comparison.
// Excludes primary key and fields with db:"-" tag.
func buildFieldMapForComparison(structValue reflect.Value, primaryKeyFieldName string) map[string]reflect.Value {
	fieldMap := make(map[string]reflect.Value)

	iterateStructFields(structValue.Type(), structValue, primaryKeyFieldName, func(field reflect.StructField, fieldValue reflect.Value, columnName string) bool {
		fieldMap[columnName] = fieldValue
		return true
	})

	return fieldMap
}
