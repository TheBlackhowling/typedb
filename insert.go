package typedb

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

// supportsLastInsertId checks if the driver supports LastInsertId().
// Only MySQL and SQLite support LastInsertId().
// PostgreSQL and SQL Server require RETURNING/OUTPUT clauses.
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
	case interface{ GetDriverName() string }:
		// Handle wrappers that expose driver name (e.g., oracleTestWrapper)
		return e.GetDriverName()
	default:
		return ""
	}
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

	// Handle Oracle specially - it requires RETURNING ... INTO syntax with sql.Out
	driverName := getDriverName(exec)
	driverNameLower := strings.ToLower(driverName)
	if driverNameLower == "oracle" {
		// Oracle requires RETURNING ... INTO :bindvar syntax
		// Find RETURNING clause (case-insensitive) in original query
		originalReturningIdx := -1
		for i := 0; i <= len(insertQuery)-8; i++ {
			if len(insertQuery) >= i+9 && strings.EqualFold(insertQuery[i:i+9], "RETURNING ") {
				originalReturningIdx = i
				break
			} else if len(insertQuery) >= i+8 && strings.EqualFold(insertQuery[i:i+8], "RETURNING") {
				// Check if next char is space or end of string
				if i+8 >= len(insertQuery) || insertQuery[i+8] == ' ' {
					originalReturningIdx = i
					break
				}
			}
		}
		if originalReturningIdx == -1 {
			return 0, fmt.Errorf("typedb: InsertAndGetId Oracle query must contain RETURNING clause")
		}

		// Extract the RETURNING part (everything after RETURNING)
		returningPart := insertQuery[originalReturningIdx+8:] // Skip "RETURNING" (8 chars)
		// Skip any whitespace after RETURNING
		returningPart = strings.TrimLeft(returningPart, " \t\n\r")
		
		// Find where RETURNING clause ends (before INTO, or end of query)
		intoIdx := strings.Index(strings.ToUpper(returningPart), " INTO ")
		if intoIdx != -1 {
			// Already has INTO clause, use as-is but need to handle sql.Out
			var id int64
			outParam := sql.Out{Dest: &id}
			argsWithOut := make([]any, len(args)+1)
			copy(argsWithOut, args)
			argsWithOut[len(args)] = outParam

			_, err := exec.Exec(ctx, insertQuery, argsWithOut...)
			if err != nil {
				return 0, fmt.Errorf("typedb: InsertAndGetId failed: %w", err)
			}
			return id, nil
		}

		// Need to add INTO clause - extract what's being returned (usually just "id")
		returningFields := strings.TrimSpace(returningPart)
		// Remove any trailing parts (like FROM, WHERE, etc. shouldn't be there in INSERT RETURNING)
		if spaceIdx := strings.Index(returningFields, " "); spaceIdx != -1 {
			returningFields = returningFields[:spaceIdx]
		}

		// Build new query with INTO clause
		queryBeforeReturning := insertQuery[:originalReturningIdx+8] // Everything up to and including "RETURNING"
		returningPlaceholder := fmt.Sprintf(":%d", len(args)+1)
		newQuery := queryBeforeReturning + " " + returningFields + " INTO " + returningPlaceholder

		var id int64
		outParam := sql.Out{Dest: &id}
		argsWithOut := make([]any, len(args)+1)
		copy(argsWithOut, args)
		argsWithOut[len(args)] = outParam

		_, err := exec.Exec(ctx, newQuery, argsWithOut...)
		if err != nil {
			return 0, fmt.Errorf("typedb: InsertAndGetId failed: %w", err)
		}
		return id, nil
	}

	// For databases with RETURNING/OUTPUT, use QueryRowMap to get the ID directly
	row, err := exec.QueryRowMap(ctx, insertQuery, args...)
	if err != nil {
		return 0, fmt.Errorf("typedb: InsertAndGetId failed: %w", err)
	}
	
	// Extract ID from the returned row
	idValue, ok := row["id"]
	if !ok {
		// Try uppercase (SQL Server sometimes)
		idValue, ok = row["ID"]
		if !ok {
			return 0, fmt.Errorf("typedb: InsertAndGetId RETURNING/OUTPUT clause did not return 'id' column")
		}
	}
	
	// Convert to int64
	switch v := idValue.(type) {
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int:
		return int64(v), nil
	case float64:
		// Handle JSON number deserialization
		return int64(v), nil
	default:
		return 0, fmt.Errorf("typedb: InsertAndGetId returned non-integer ID type: %T", idValue)
	}
}

// getTableName gets the table name from a model using TableName() method.
// Returns error if TableName() method doesn't exist or returns empty string.
func getTableName(model ModelInterface) (string, error) {
	// Try to call TableName() method
	_, found := findMethod(model, "TableName")
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
	modelType := getModelType(model)
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
// Escapes quote characters to prevent SQL injection.
// Identifiers come from struct tags (compile-time constants), so no runtime validation is needed.
func quoteIdentifier(driverName, identifier string) string {
	if identifier == "" {
		panic("typedb: identifier cannot be empty")
	}
	
	driverName = strings.ToLower(driverName)
	switch driverName {
	case "postgres", "sqlite3":
		// Escape double quotes by doubling them
		escaped := strings.ReplaceAll(identifier, `"`, `""`)
		return `"` + escaped + `"`
	case "mysql":
		// Escape backticks by doubling them
		escaped := strings.ReplaceAll(identifier, "`", "``")
		return "`" + escaped + "`"
	case "sqlserver", "mssql":
		// Square brackets don't need escaping, but validate no closing bracket
		if strings.Contains(identifier, "]") {
			panic(fmt.Sprintf("typedb: SQL Server identifier cannot contain ']': %s", identifier))
		}
		return "[" + identifier + "]"
	case "oracle":
		// Escape double quotes by doubling them
		escaped := strings.ReplaceAll(identifier, `"`, `""`)
		// Oracle defaults to uppercase, but we'll preserve what's provided
		return `"` + strings.ToUpper(escaped) + `"`
	default:
		// Default to PostgreSQL style
		escaped := strings.ReplaceAll(identifier, `"`, `""`)
		return `"` + escaped + `"`
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

// fieldVisitor is a callback function that processes each field during struct iteration.
// Parameters:
//   - field: the struct field metadata
//   - fieldValue: the field's reflect.Value
//   - columnName: the extracted column name (handles dot notation)
// Returns: true if iteration should continue, false to stop
type fieldVisitor func(field reflect.StructField, fieldValue reflect.Value, columnName string) bool

// iterateStructFields iterates over struct fields, handling embedded structs and extracting db tags.
// It calls the visitor function for each valid field (exported, has db tag, not primary key).
// The visitor receives the field metadata, field value, and extracted column name.
func iterateStructFields(structType reflect.Type, structValue reflect.Value, primaryKeyFieldName string, visitor fieldVisitor) {
	if structType.Kind() != reflect.Struct {
		return
	}

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

			// Extract column name (handle dot notation - use last part)
			columnName := dbTag
			if strings.Contains(dbTag, ".") {
				parts := strings.Split(dbTag, ".")
				columnName = parts[len(parts)-1]
			}

			// Call visitor - if it returns false, stop iteration
			if !visitor(field, fieldValue, columnName) {
				return
			}
		}
	}

	processFields(structType, structValue)
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

	iterateStructFields(modelValue.Type(), modelValue, primaryKeyFieldName, func(field reflect.StructField, fieldValue reflect.Value, columnName string) bool {
		// Skip fields with dbInsert:"false" tag
		if field.Tag.Get("dbInsert") == "false" {
			return true
		}

		// Skip nil/zero values
		if isZeroOrNil(fieldValue) {
			return true
		}

		columns = append(columns, columnName)
		values = append(values, fieldValue.Interface())
		return true
	})

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
	primaryField, found := findFieldByTag(model, "load", "primary")
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
		return setFieldValue(model, primaryField.Name, id)
	}

	// Handle Oracle (go-ora driver requires special RETURNING INTO syntax)
	// Use RETURNING ... INTO with sql.Out for bind variable
	if driverNameLower == "oracle" {
		quotedPK := quoteIdentifier(driverName, primaryKeyColumn)
		// Oracle requires RETURNING ... INTO :N where N is the next positional placeholder
		// Use sql.Out to capture the returned ID
		returningPlaceholder := fmt.Sprintf(":%d", len(values)+1)
		insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING %s INTO %s",
			quotedTableName,
			strings.Join(quotedColumns, ", "),
			strings.Join(placeholders, ", "),
			quotedPK,
			returningPlaceholder)
		
		// Create sql.Out parameter for the RETURNING INTO bind variable
		var id int64
		outParam := sql.Out{Dest: &id}
		
		// Append sql.Out to values for the RETURNING INTO clause
		args := make([]any, len(values)+1)
		copy(args, values)
		args[len(values)] = outParam
		
		result, err := exec.Exec(ctx, insertQuery, args...)
		if err != nil {
			return fmt.Errorf("typedb: Insert failed: %w", err)
		}

		// Verify the operation succeeded
		if result == nil {
			return fmt.Errorf("typedb: Insert returned nil result")
		}

		// The ID is now in the id variable from sql.Out
		// Set primary key on model
		return setFieldValue(model, primaryField.Name, id)
	}

	// For databases with RETURNING support (PostgreSQL, SQLite, SQL Server)
	// SQL Server OUTPUT clause comes BEFORE VALUES, not after
	var insertQuery string
	if driverNameLower == "sqlserver" || driverNameLower == "mssql" {
		insertQuery = fmt.Sprintf("INSERT INTO %s (%s)%s VALUES (%s)",
			quotedTableName,
			strings.Join(quotedColumns, ", "),
			returningClause,
			strings.Join(placeholders, ", "))
	} else {
		insertQuery = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)%s",
			quotedTableName,
			strings.Join(quotedColumns, ", "),
			strings.Join(placeholders, ", "),
			returningClause)
	}

	row, err := exec.QueryRowMap(ctx, insertQuery, values...)
	if err != nil {
		return fmt.Errorf("typedb: Insert failed: %w", err)
	}

	// Extract primary key value from returned row
	idValue, ok := row[primaryKeyColumn]
	if !ok {
		// Try uppercase (SQL Server sometimes)
		idValue, ok = row[strings.ToUpper(primaryKeyColumn)]
		if !ok {
			return fmt.Errorf("typedb: Insert RETURNING clause did not return primary key column %s", primaryKeyColumn)
		}
	}

	// Set primary key on model
	return setFieldValue(model, primaryField.Name, idValue)
}

// InsertAndLoad inserts a model and then loads the full object from the database.
// This is a convenience function that combines Insert() and Load().
// The model must:
//   - Implement TableName() method that returns the table name
//   - Have a field with load:"primary" tag (for ID retrieval and loading)
//   - Have a QueryBy{PrimaryField}() method for loading
//   - Not have dot notation in db tags (simple model, not joined)
//
// Nil/zero value fields are excluded from the INSERT.
// After insertion, the model is loaded from the database with all fields populated.
//
// Example:
//
//	type User struct {
//	    Model
//	    ID        int    `db:"id" load:"primary"`
//	    Name      string `db:"name"`
//	    Email     string `db:"email"`
//	    CreatedAt string `db:"created_at"`
//	}
//
//	func (u *User) TableName() string {
//	    return "users"
//	}
//
//	func (u *User) QueryByID() string {
//	    return "SELECT id, name, email, created_at FROM users WHERE id = $1"
//	}
//
//	user := &User{Name: "John", Email: "john@example.com"}
//	returnedUser, err := typedb.InsertAndLoad(ctx, db, user)
//	// returnedUser.ID, returnedUser.CreatedAt, etc. are all populated
func InsertAndLoad[T ModelInterface](ctx context.Context, exec Executor, model T) (T, error) {
	var zero T
	
	// First, insert the model (this sets the ID)
	err := Insert(ctx, exec, model)
	if err != nil {
		return zero, fmt.Errorf("typedb: InsertAndLoad failed during insert: %w", err)
	}
	
	// Then load the full object
	err = Load(ctx, exec, model)
	if err != nil {
		return zero, fmt.Errorf("typedb: InsertAndLoad failed during load: %w", err)
	}
	
	return model, nil
}
