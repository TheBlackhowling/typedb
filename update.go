package typedb

import (
	"context"
	"fmt"
	"reflect"
	"strings"
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
// Example:
//
//	type User struct {
//	    Model
//	    ID    int    `db:"id" load:"primary"`
//	    Name  string `db:"name"`
//	    Email string `db:"email"`
//	}
//
//	func (u *User) TableName() string {
//	    return "users"
//	}
//
//	user := &User{ID: 123, Name: "John Updated"}
//	err := typedb.Update(ctx, db, user)
//	// Generates: UPDATE users SET name = $1 WHERE id = $2
func Update[T ModelInterface](ctx context.Context, exec Executor, model T) error {
	// Validate model has TableName() method
	tableName, err := getTableName(model)
	if err != nil {
		return fmt.Errorf("typedb: Update validation failed: %w", err)
	}

	// Validate model doesn't have dot notation (not a joined model)
	if hasDotNotation(model) {
		return fmt.Errorf("typedb: Update cannot be used with joined models (detected dot notation in db tags)")
	}

	// Find primary key field
	primaryField, found := FindFieldByTag(model, "load", "primary")
	if !found {
		return fmt.Errorf("typedb: Update requires a field with load:\"primary\" tag")
	}

	// Get primary key column name from db tag
	primaryKeyColumn := primaryField.Tag.Get("db")
	if primaryKeyColumn == "" || primaryKeyColumn == "-" {
		return fmt.Errorf("typedb: primary key field %s must have a db tag", primaryField.Name)
	}

	// Extract column name (handle dot notation - use last part)
	if strings.Contains(primaryKeyColumn, ".") {
		parts := strings.Split(primaryKeyColumn, ".")
		primaryKeyColumn = parts[len(parts)-1]
	}

	// Get primary key value from model (for WHERE clause)
	primaryKeyValue, err := GetFieldValue(model, primaryField.Name)
	if err != nil {
		return fmt.Errorf("typedb: Update failed to get primary key value: %w", err)
	}

	// Validate primary key is set (non-zero)
	if isZeroOrNil(primaryKeyValue) {
		return fmt.Errorf("typedb: Update requires primary key field %s to be set (non-zero value)", primaryField.Name)
	}

	// Collect fields and values (excluding primary key, nil/zero values, and fields with noupdate tag)
	columns, values, err := serializeModelFieldsForUpdate(model, primaryField.Name)
	if err != nil {
		return fmt.Errorf("typedb: Update failed to serialize model: %w", err)
	}

	if len(columns) == 0 {
		return fmt.Errorf("typedb: Update requires at least one non-nil field to update")
	}

	// Get driver name for database-specific SQL generation
	driverName := getDriverName(exec)

	// Build UPDATE query
	quotedTableName := quoteIdentifier(driverName, tableName)
	quotedPrimaryKeyColumn := quoteIdentifier(driverName, primaryKeyColumn)

	// Build SET clause
	var setClauses []string
	for i, col := range columns {
		quotedCol := quoteIdentifier(driverName, col)
		placeholder := generatePlaceholder(driverName, i+1)
		setClauses = append(setClauses, fmt.Sprintf("%s = %s", quotedCol, placeholder))
	}

	// Add primary key value to args for WHERE clause
	wherePlaceholder := generatePlaceholder(driverName, len(values)+1)
	allValues := append(values, primaryKeyValue.Interface())

	// Build UPDATE query
	updateQuery := fmt.Sprintf("UPDATE %s SET %s WHERE %s = %s",
		quotedTableName,
		strings.Join(setClauses, ", "),
		quotedPrimaryKeyColumn,
		wherePlaceholder)

	// Execute UPDATE
	_, err = exec.Exec(ctx, updateQuery, allValues...)
	if err != nil {
		return fmt.Errorf("typedb: Update failed: %w", err)
	}

	return nil
}

// serializeModelFieldsForUpdate collects non-nil/non-zero fields from a model for UPDATE operations.
// Excludes primary key field, fields with db:"-" tag, and fields with dbUpdate:"false" tag.
// Fields with db:"-" are excluded from all database operations (INSERT, UPDATE, SELECT).
// Fields with dbUpdate:"false" are excluded from UPDATE but can still be used in INSERT and SELECT.
// Returns: column names and field values for serialization.
func serializeModelFieldsForUpdate(model ModelInterface, primaryKeyFieldName string) ([]string, []any, error) {
	modelValue := reflect.ValueOf(model)
	if modelValue.Kind() != reflect.Ptr || modelValue.IsNil() {
		return nil, nil, fmt.Errorf("typedb: model must be a non-nil pointer")
	}

	modelValue = modelValue.Elem()
	if modelValue.Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("typedb: model must be a pointer to struct")
	}

	var columns []string
	var values []any

	modelType := modelValue.Type()
	var processFields func(reflect.Type, reflect.Value)
	processFields = func(t reflect.Type, v reflect.Value) {
		if t.Kind() != reflect.Struct {
			return
		}

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}

			fieldValue := v.Field(i)

			// Handle embedded structs
			if field.Anonymous {
				embeddedType := field.Type
				if embeddedType.Kind() == reflect.Ptr {
					if fieldValue.IsNil() {
						continue
					}
					embeddedType = embeddedType.Elem()
					fieldValue = fieldValue.Elem()
				}
				if embeddedType.Kind() == reflect.Struct {
					processFields(embeddedType, fieldValue)
					continue
				}
			}

			// Get db tag
			dbTag := field.Tag.Get("db")
			if dbTag == "" || dbTag == "-" {
				continue
			}

			// Skip if this is the primary key field
			if field.Name == primaryKeyFieldName {
				continue
			}

			// Skip fields with dbUpdate:"false" tag
			if field.Tag.Get("dbUpdate") == "false" {
				continue
			}

			// Skip nil/zero values
			if isZeroOrNil(fieldValue) {
				continue
			}

			// Extract column name (handle dot notation - use last part)
			columnName := dbTag
			if strings.Contains(dbTag, ".") {
				parts := strings.Split(dbTag, ".")
				columnName = parts[len(parts)-1]
			}

			columns = append(columns, columnName)
			values = append(values, fieldValue.Interface())
		}
	}

	processFields(modelType, modelValue)

	return columns, values, nil
}
