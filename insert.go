package typedb

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
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
// insertAndGetIdOracle handles Oracle-specific insert and get ID logic
func insertAndGetIdOracle(ctx context.Context, exec Executor, insertQuery string, args []any) (int64, error) {
	queryUpper := strings.ToUpper(insertQuery)
	returningIdx := strings.Index(queryUpper, "RETURNING")
	if returningIdx == -1 {
		return 0, fmt.Errorf("typedb: InsertAndGetId Oracle query must contain RETURNING clause")
	}

	returningPart := insertQuery[returningIdx+9:]
	returningPart = strings.TrimLeft(returningPart, " \t\n\r")

	intoIdx := strings.Index(strings.ToUpper(returningPart), " INTO ")
	if intoIdx != -1 {
		// Already has INTO clause
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

	// Need to add INTO clause
	returningFields := returningPart
	if spaceIdx := strings.Index(returningFields, " "); spaceIdx != -1 {
		returningFields = returningFields[:spaceIdx]
	}

	queryBeforeReturning := insertQuery[:returningIdx+9]
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

// convertIdToInt64 converts various ID types to int64
func convertIdToInt64(idValue any) (int64, error) {
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
		return int64(v), nil
	default:
		return 0, fmt.Errorf("typedb: InsertAndGetId returned non-integer ID type: %T", idValue)
	}
}

// extractIdFromRow extracts ID value from QueryRowMap result
func extractIdFromRow(row map[string]any) (any, error) {
	idValue, ok := row["id"]
	if !ok {
		idValue, ok = row["ID"]
		if !ok {
			return nil, fmt.Errorf("typedb: InsertAndGetId RETURNING/OUTPUT clause did not return 'id' column")
		}
	}
	return idValue, nil
}

func InsertAndGetId(ctx context.Context, exec Executor, insertQuery string, args ...any) (int64, error) {
	queryUpper := strings.ToUpper(insertQuery)
	hasReturning := strings.Contains(queryUpper, "RETURNING") || strings.Contains(queryUpper, "OUTPUT")

	if !hasReturning {
		// MySQL/SQLite path
		driverName := getDriverName(exec)
		if !supportsLastInsertId(driverName) {
			return 0, fmt.Errorf("typedb: InsertAndGetId requires RETURNING or OUTPUT clause for %s. Only MySQL and SQLite support LastInsertId() without RETURNING/OUTPUT", driverName)
		}

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

	// Oracle path
	driverName := getDriverName(exec)
	driverNameLower := strings.ToLower(driverName)
	if driverNameLower == "oracle" {
		return insertAndGetIdOracle(ctx, exec, insertQuery, args)
	}

	// Standard RETURNING/OUTPUT path
	row, err := exec.QueryRowMap(ctx, insertQuery, args...)
	if err != nil {
		return 0, fmt.Errorf("typedb: InsertAndGetId failed: %w", err)
	}

	idValue, err := extractIdFromRow(row)
	if err != nil {
		return 0, err
	}

	return convertIdToInt64(idValue)
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

// validateIdentifier validates that an identifier contains only allowed characters.
// Identifiers can contain alphanumeric characters, underscores, dots (for qualified names),
// and quote characters (which will be escaped). Dangerous characters like semicolons,
// SQL keywords, etc. are rejected to prevent SQL injection.
// Returns an error if the identifier is invalid.
func validateIdentifier(identifier string) error {
	if identifier == "" {
		return fmt.Errorf("typedb: identifier cannot be empty")
	}

	// Allow alphanumeric, underscore, dot, and quote characters
	// Reject dangerous characters: semicolon, dash, parentheses, etc.
	// This regex matches: letters, digits, underscore, dot, and quote characters
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_."` + "`" + `]+$`)
	if !validPattern.MatchString(identifier) {
		return fmt.Errorf("typedb: invalid identifier '%s': identifiers can only contain alphanumeric characters, underscores, dots, and quote characters", identifier)
	}

	// Additional check: reject identifiers that contain SQL injection patterns
	// Note: We don't reject SQL keywords as they might be legitimate identifier names
	// The regex above already rejects semicolons, spaces, and other dangerous characters
	dangerousPatterns := []string{
		";",
		"--",
		"/*",
		"*/",
	}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(identifier, pattern) {
			return fmt.Errorf("typedb: invalid identifier '%s': contains potentially dangerous SQL pattern", identifier)
		}
	}

	return nil
}

// quoteIdentifier quotes an identifier based on driver name.
// Validates the identifier and escapes quote characters to prevent SQL injection.
// Panics if identifier is invalid (since identifiers come from struct tags at compile time).
func quoteIdentifier(driverName, identifier string) string {
	// Validate identifier first
	if err := validateIdentifier(identifier); err != nil {
		// Panic is acceptable here since identifiers come from struct tags (compile-time constants)
		// If this panics, it indicates a programming error, not a runtime security issue
		panic(err.Error())
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
// Returns: column names, field values, and mask indices (for fields with nolog:"true" tag).
func serializeModelFields(model ModelInterface, primaryKeyFieldName string) ([]string, []any, []int, error) {
	modelValue := reflect.ValueOf(model)
	if modelValue.Kind() != reflect.Ptr || modelValue.IsNil() {
		return nil, nil, nil, fmt.Errorf("typedb: model must be a non-nil pointer")
	}

	modelValue = modelValue.Elem()
	if modelValue.Kind() != reflect.Struct {
		return nil, nil, nil, fmt.Errorf("typedb: model must be a pointer to struct")
	}

	var columns []string
	var values []any
	var maskIndices []int

	iterateStructFields(modelValue.Type(), modelValue, primaryKeyFieldName, func(field reflect.StructField, fieldValue reflect.Value, columnName string) bool {
		// Skip fields with dbInsert:"false" tag
		if field.Tag.Get("dbInsert") == "false" {
			return true
		}

		// Skip nil/zero values
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

	return columns, values, maskIndices, nil
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
// buildInsertQueryParts builds the common INSERT query parts (quoted table name, quoted columns, placeholders)
func buildInsertQueryParts(driverName, tableName string, columns []string, values []any) (string, []string, []string) {
	quotedTableName := quoteIdentifier(driverName, tableName)
	quotedColumns := make([]string, len(columns))
	for i, col := range columns {
		quotedColumns[i] = quoteIdentifier(driverName, col)
	}

	placeholders := make([]string, len(values))
	for i := 1; i <= len(values); i++ {
		placeholders[i-1] = generatePlaceholder(driverName, i)
	}

	return quotedTableName, quotedColumns, placeholders
}

// insertMySQL handles MySQL-specific insert logic (uses LastInsertId instead of RETURNING)
func insertMySQL[T ModelInterface](ctx context.Context, exec Executor, model T,
	quotedTableName string, quotedColumns []string, placeholders []string, values []any,
	primaryField *reflect.StructField) error {
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

	return setFieldValue(model, primaryField.Name, id)
}

// insertOracle handles Oracle-specific insert logic (uses RETURNING ... INTO with sql.Out)
func insertOracle[T ModelInterface](ctx context.Context, exec Executor, model T,
	driverName string, quotedTableName string, quotedColumns []string, placeholders []string,
	values []any, primaryKeyColumn string, primaryField *reflect.StructField) error {
	quotedPK := quoteIdentifier(driverName, primaryKeyColumn)
	returningPlaceholder := fmt.Sprintf(":%d", len(values)+1)
	insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING %s INTO %s",
		quotedTableName,
		strings.Join(quotedColumns, ", "),
		strings.Join(placeholders, ", "),
		quotedPK,
		returningPlaceholder)

	var id int64
	outParam := sql.Out{Dest: &id}
	args := make([]any, len(values)+1)
	copy(args, values)
	args[len(values)] = outParam

	result, err := exec.Exec(ctx, insertQuery, args...)
	if err != nil {
		return fmt.Errorf("typedb: Insert failed: %w", err)
	}

	if result == nil {
		return fmt.Errorf("typedb: Insert returned nil result")
	}

	return setFieldValue(model, primaryField.Name, id)
}

// insertWithReturning handles standard RETURNING/OUTPUT path (PostgreSQL, SQLite, SQL Server)
func insertWithReturning[T ModelInterface](ctx context.Context, exec Executor, model T,
	driverName string, quotedTableName string, quotedColumns []string, placeholders []string,
	values []any, returningClause string, primaryKeyColumn string, primaryField *reflect.StructField) error {
	driverNameLower := strings.ToLower(driverName)
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

	idValue, ok := row[primaryKeyColumn]
	if !ok {
		idValue, ok = row[strings.ToUpper(primaryKeyColumn)]
		if !ok {
			return fmt.Errorf("typedb: Insert RETURNING clause did not return primary key column %s", primaryKeyColumn)
		}
	}

	return setFieldValue(model, primaryField.Name, idValue)
}

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
	columns, values, maskIndices, err := serializeModelFields(model, primaryField.Name)
	if err != nil {
		return fmt.Errorf("typedb: Insert failed to serialize model: %w", err)
	}

	if len(columns) == 0 {
		return fmt.Errorf("typedb: Insert requires at least one non-nil field to insert")
	}

	// Store mask indices in context for logging
	if len(maskIndices) > 0 {
		ctx = WithMaskIndices(ctx, maskIndices)
	}

	// Get driver name for database-specific SQL generation
	driverName := getDriverName(exec)

	// Build common query parts
	quotedTableName, quotedColumns, placeholders := buildInsertQueryParts(driverName, tableName, columns, values)

	// Route to database-specific handler
	driverNameLower := strings.ToLower(driverName)
	switch driverNameLower {
	case "mysql":
		return insertMySQL(ctx, exec, model, quotedTableName, quotedColumns, placeholders, values, primaryField)
	case "oracle":
		return insertOracle(ctx, exec, model, driverName, quotedTableName, quotedColumns, placeholders, values, primaryKeyColumn, primaryField)
	default:
		returningClause := buildReturningClause(driverName, primaryKeyColumn)
		return insertWithReturning(ctx, exec, model, driverName, quotedTableName, quotedColumns, placeholders, values, returningClause, primaryKeyColumn, primaryField)
	}
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
