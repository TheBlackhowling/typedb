package typedb

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

// supportsLastInsertId checks if the driver supports LastInsertId().
// Only MySQL and SQLite support LastInsertId().
// PostgreSQL, SQL Server, and Oracle require RETURNING/OUTPUT clauses.
func supportsLastInsertId(driverName string) bool {
	driverName = strings.ToLower(driverName)
	return driverName == "mysql" || driverName == "sqlite3"
}

// getDriverName extracts the driver name from an Executor.
// Returns empty string if driver name cannot be determined.
func getDriverName(exec Executor) string {
	switch e := exec.(type) {
	case *DB:
		return e.driverName
	case *Tx:
		return e.driverName
	default:
		return ""
	}
}

// InsertAndReturn executes an INSERT statement with a RETURNING/OUTPUT clause
// and deserializes the returned row into a model instance.
// The insertQuery must include a RETURNING clause (PostgreSQL/SQLite),
// OUTPUT clause (SQL Server), or RETURNING ... INTO clause (Oracle).
//
// Note: MySQL does not support RETURNING/OUTPUT clauses. For MySQL, use LAST_INSERT_ID()
// manually or execute a separate SELECT query after INSERT.
//
// Example:
//
//	user := &User{Name: "John", Email: "john@example.com"}
//	returned, err := typedb.InsertAndReturn[*User](ctx, db,
//		"INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id, created_at, updated_at",
//		user.Name, user.Email)
//	// returned.ID, returned.CreatedAt, returned.UpdatedAt are now populated
func InsertAndReturn[T ModelInterface](ctx context.Context, exec Executor, insertQuery string, args ...any) (T, error) {
	var zero T

	// Execute the INSERT with RETURNING/OUTPUT and get the returned row
	row, err := exec.QueryRowMap(ctx, insertQuery, args...)
	if err != nil {
		return zero, fmt.Errorf("typedb: InsertAndReturn failed: %w", err)
	}

	// Deserialize the returned row into a new model instance
	model, err := DeserializeForType[T](row)
	if err != nil {
		return zero, fmt.Errorf("typedb: InsertAndReturn deserialization failed: %w", err)
	}

	return model, nil
}

// InsertedId is an internal model used by InsertAndGetId to retrieve just the ID.
// Note: This uses int64 to support all integer ID types (SMALLINT/int16, INTEGER/int32, BIGINT/int64).
// The deserialization layer automatically converts smaller integer types to int64.
// For type-safe ID retrieval with specific types, use InsertAndReturn with your own model.
type InsertedId struct {
	Model
	ID int64 `db:"id"`
}

func init() {
	// Register InsertedId so Model.Deserialize can find the outer struct type
	// when called directly on an InsertedId instance.
	RegisterModel[*InsertedId]()
}

// InsertAndGetId executes an INSERT statement and returns the inserted ID as int64.
// This is a convenience helper that works with all supported databases.
//
// For databases with RETURNING/OUTPUT support (PostgreSQL, SQLite, SQL Server, Oracle),
// the insertQuery should include a RETURNING id clause, OUTPUT INSERTED.id clause,
// or RETURNING id INTO :id clause.
//
// For MySQL (which doesn't support RETURNING/OUTPUT), this function uses sql.Result.LastInsertId(),
// which is safe because it uses the same connection as the INSERT operation, avoiding race conditions.
//
// Note: This function returns int64, which works for all integer ID types:
// - SMALLINT (int16) - automatically converted to int64
// - INTEGER (int32) - automatically converted to int64
// - BIGINT (int64) - returned as-is
//
// For type-safe ID retrieval with specific types (e.g., int16, int32), use InsertAndReturn
// with your own model that has the correct ID field type.
//
// Example (PostgreSQL/SQLite/SQL Server/Oracle):
//
//	id, err := typedb.InsertAndGetId(ctx, db,
//		"INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
//		"John", "john@example.com")
//
// Example (MySQL):
//
//	id, err := typedb.InsertAndGetId(ctx, db,
//		"INSERT INTO users (name, email) VALUES (?, ?)",
//		"John", "john@example.com")
//
// For type-safe retrieval with int32 ID:
//
//	type UserID struct {
//		ID int32 `db:"id"`
//	}
//	result, err := typedb.InsertAndReturn[*UserID](ctx, db,
//		"INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
//		"John", "john@example.com")
func InsertAndGetId(ctx context.Context, exec Executor, insertQuery string, args ...any) (int64, error) {
	// Check if query has RETURNING/OUTPUT clause
	queryUpper := strings.ToUpper(insertQuery)
	hasReturning := strings.Contains(queryUpper, "RETURNING") || strings.Contains(queryUpper, "OUTPUT")

	if !hasReturning {
		// Check if driver supports LastInsertId() (MySQL and SQLite only)
		driverName := getDriverName(exec)
		if !supportsLastInsertId(driverName) {
			return 0, fmt.Errorf("typedb: InsertAndGetId requires RETURNING or OUTPUT clause for %s. Only MySQL and SQLite support LastInsertId() without RETURNING/OUTPUT", driverName)
		}

		// MySQL/SQLite case: use sql.Result.LastInsertId() which is safe because it uses
		// the same connection as the Exec() call, avoiding race conditions.
		result, err := exec.Exec(ctx, insertQuery, args...)
		if err != nil {
			return 0, fmt.Errorf("typedb: InsertAndGetId INSERT failed: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("typedb: InsertAndGetId LastInsertId failed: %w", err)
		}

		return id, nil
	}

	// Use InsertAndReturn with the internal InsertedId model for RETURNING/OUTPUT databases
	result, err := InsertAndReturn[*InsertedId](ctx, exec, insertQuery, args...)
	if err != nil {
		return 0, err
	}

	return result.ID, nil
}

// getTableName gets the table name from a model using TableName() method.
// Returns error if TableName() method doesn't exist or returns empty string.
func getTableName(model ModelInterface) (string, error) {
	// Try to call TableName() method
	_, found := FindMethod(model, "TableName")
	if !found {
		return "", fmt.Errorf("typedb: model must implement TableName() method")
	}

	// Call the method
	results := reflect.ValueOf(model).MethodByName("TableName").Call(nil)
	if len(results) != 1 {
		return "", fmt.Errorf("typedb: TableName() method must return exactly one value")
	}

	tableName := results[0].String()
	if tableName == "" {
		return "", fmt.Errorf("typedb: TableName() method returned empty string")
	}

	return tableName, nil
}

// hasDotNotation checks if any db tags contain dot notation (indicating joined model).
func hasDotNotation(model ModelInterface) bool {
	modelType := GetModelType(model)
	return checkDotNotationRecursive(modelType)
}

func checkDotNotationRecursive(t reflect.Type) bool {
	if t.Kind() != reflect.Struct {
		return false
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		// Check db tag for dot notation
		dbTag := field.Tag.Get("db")
		if dbTag != "" && dbTag != "-" && strings.Contains(dbTag, ".") {
			return true
		}

		// Check embedded structs
		if field.Anonymous {
			embeddedType := field.Type
			if embeddedType.Kind() == reflect.Ptr {
				embeddedType = embeddedType.Elem()
			}
			if checkDotNotationRecursive(embeddedType) {
				return true
			}
		}
	}

	return false
}

// generatePlaceholder generates a parameter placeholder based on driver name and position.
func generatePlaceholder(driverName string, position int) string {
	driverName = strings.ToLower(driverName)
	switch driverName {
	case "postgres":
		return fmt.Sprintf("$%d", position)
	case "mysql":
		return "?"
	case "sqlite3":
		return "?"
	case "sqlserver", "mssql":
		return fmt.Sprintf("@p%d", position)
	case "oracle":
		return fmt.Sprintf(":%d", position)
	default:
		// Default to PostgreSQL style
		return fmt.Sprintf("$%d", position)
	}
}

// quoteIdentifier quotes an identifier based on driver name.
func quoteIdentifier(driverName, identifier string) string {
	driverName = strings.ToLower(driverName)
	switch driverName {
	case "postgres", "sqlite3":
		return `"` + identifier + `"`
	case "mysql":
		return "`" + identifier + "`"
	case "sqlserver", "mssql":
		return "[" + identifier + "]"
	case "oracle":
		// Oracle defaults to uppercase, but we'll preserve what's provided
		return `"` + strings.ToUpper(identifier) + `"`
	default:
		// Default to PostgreSQL style
		return `"` + identifier + `"`
	}
}

// buildReturningClause builds a RETURNING/OUTPUT clause based on driver name and primary key column.
// For Oracle, returns standard RETURNING clause (driver handles bind variables via QueryRowMap).
func buildReturningClause(driverName, primaryKeyColumn string) string {
	driverName = strings.ToLower(driverName)
	quotedPK := quoteIdentifier(driverName, primaryKeyColumn)

	switch driverName {
	case "postgres", "sqlite3":
		return " RETURNING " + quotedPK
	case "sqlserver", "mssql":
		return " OUTPUT INSERTED." + quotedPK
	case "oracle":
		// Oracle uses RETURNING ... INTO :bindvar, but QueryRowMap handles bind variables automatically
		// So we can use standard RETURNING syntax
		return " RETURNING " + quotedPK
	case "mysql":
		// MySQL doesn't support RETURNING
		return ""
	default:
		// Default to PostgreSQL style
		return " RETURNING " + quotedPK
	}
}

// serializeModelFields collects non-nil/non-zero fields from a model and returns columns and values.
// Excludes primary key field, fields with db:"-" tag, and fields with dbInsert:"false" tag.
// Fields with db:"-" are excluded from all database operations (INSERT, UPDATE, SELECT).
// Fields with dbInsert:"false" are excluded from INSERT but can still be used in UPDATE and SELECT.
// Returns: column names and field values for serialization.
func serializeModelFields(model ModelInterface, primaryKeyFieldName string) ([]string, []any, error) {
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

			// Skip if this is the primary key field (we'll get it from RETURNING)
			if field.Name == primaryKeyFieldName {
				continue
			}

			// Skip fields with dbInsert:"false" tag
			if field.Tag.Get("dbInsert") == "false" {
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

// isZeroOrNil checks if a reflect.Value is zero or nil.
func isZeroOrNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return v.IsNil()
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	default:
		return false
	}
}

// Insert inserts a model into the database by automatically building the INSERT query.
// The model must:
//   - Implement TableName() method that returns the table name
//   - Have a field with load:"primary" tag (for ID retrieval)
//   - Not have dot notation in db tags (simple model, not joined)
//
// Nil/zero value fields are excluded from the INSERT.
// The primary key field is set on the model after insertion.
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
//	user := &User{Name: "John", Email: "john@example.com"}
//	err := typedb.Insert(ctx, db, user)
//	// user.ID is now set with the inserted ID
func Insert[T ModelInterface](ctx context.Context, exec Executor, model T) error {
	// Validate model has TableName() method
	tableName, err := getTableName(model)
	if err != nil {
		return fmt.Errorf("typedb: Insert validation failed: %w", err)
	}

	// Validate model doesn't have dot notation (not a joined model)
	if hasDotNotation(model) {
		return fmt.Errorf("typedb: Insert cannot be used with joined models (detected dot notation in db tags)")
	}

	// Find primary key field
	primaryField, found := FindFieldByTag(model, "load", "primary")
	if !found {
		return fmt.Errorf("typedb: Insert requires a field with load:\"primary\" tag")
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

	// Collect fields and values (excluding primary key and nil/zero values)
	columns, values, err := serializeModelFields(model, primaryField.Name)
	if err != nil {
		return fmt.Errorf("typedb: Insert failed to serialize model: %w", err)
	}

	if len(columns) == 0 {
		return fmt.Errorf("typedb: Insert requires at least one non-nil field to insert")
	}

	// Get driver name for database-specific SQL generation
	driverName := getDriverName(exec)

	// Build INSERT query
	quotedTableName := quoteIdentifier(driverName, tableName)
	var quotedColumns []string
	for _, col := range columns {
		quotedColumns = append(quotedColumns, quoteIdentifier(driverName, col))
	}

	// Build placeholders
	var placeholders []string
	for i := 1; i <= len(values); i++ {
		placeholders = append(placeholders, generatePlaceholder(driverName, i))
	}

	// Build RETURNING clause (database-specific)
	returningClause := buildReturningClause(driverName, primaryKeyColumn)
	driverNameLower := strings.ToLower(driverName)

	// Handle MySQL (no RETURNING, use LastInsertId)
	if driverNameLower == "mysql" {
		insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			quotedTableName,
			strings.Join(quotedColumns, ", "),
			strings.Join(placeholders, ", "))

		result, err := exec.Exec(ctx, insertQuery, values...)
		if err != nil {
			return fmt.Errorf("typedb: Insert failed: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("typedb: Insert failed to get last insert ID: %w", err)
		}

		// Set primary key on model
		return SetFieldValue(model, primaryField.Name, id)
	}

	// Handle Oracle special case (RETURNING INTO requires bind variables)
	// Oracle drivers typically handle RETURNING INTO automatically via QueryRowMap
	// Use InsertAndReturn to handle Oracle's RETURNING INTO syntax
	if driverNameLower == "oracle" {
		insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)%s",
			quotedTableName,
			strings.Join(quotedColumns, ", "),
			strings.Join(placeholders, ", "),
			returningClause)

		// Use InsertAndReturn to get the full model back (handles Oracle bind variables)
		returnedModel, err := InsertAndReturn[T](ctx, exec, insertQuery, values...)
		if err != nil {
			return fmt.Errorf("typedb: Insert failed: %w", err)
		}

		// Get the primary key value from returned model and set it on original model
		returnedValue := reflect.ValueOf(returnedModel)
		if returnedValue.Kind() == reflect.Ptr {
			returnedValue = returnedValue.Elem()
		}
		returnedPKValue := returnedValue.FieldByName(primaryField.Name)
		if !returnedPKValue.IsValid() {
			return fmt.Errorf("typedb: Insert failed to get primary key from returned model")
		}

		// Set primary key on original model
		return SetFieldValue(model, primaryField.Name, returnedPKValue.Interface())
	}

	// For databases with RETURNING support (PostgreSQL, SQLite, SQL Server)
	insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)%s",
		quotedTableName,
		strings.Join(quotedColumns, ", "),
		strings.Join(placeholders, ", "),
		returningClause)

	row, err := exec.QueryRowMap(ctx, insertQuery, values...)
	if err != nil {
		return fmt.Errorf("typedb: Insert failed: %w", err)
	}

	// Extract primary key value from returned row
	idValue, ok := row[primaryKeyColumn]
	if !ok {
		// Try uppercase (Oracle, SQL Server sometimes)
		idValue, ok = row[strings.ToUpper(primaryKeyColumn)]
		if !ok {
			return fmt.Errorf("typedb: Insert RETURNING clause did not return primary key column %s", primaryKeyColumn)
		}
	}

	// Set primary key on model
	return SetFieldValue(model, primaryField.Name, idValue)
}
