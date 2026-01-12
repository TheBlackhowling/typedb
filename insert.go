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

	// Check if this is Oracle - Oracle's go-ora driver has issues with RETURNING via QueryContext
	driverName := getDriverName(exec)
	driverNameLower := strings.ToLower(driverName)
	
	if driverNameLower == "oracle" {
		// For Oracle, parse the query to extract INSERT and RETURNING parts
		// Then execute INSERT separately and query back the returned columns
		queryUpper := strings.ToUpper(insertQuery)
		returningIdx := strings.Index(queryUpper, "RETURNING")
		if returningIdx == -1 {
			return zero, fmt.Errorf("typedb: InsertAndReturn requires RETURNING clause for Oracle")
		}
		
		// Extract INSERT part (without RETURNING)
		insertPart := insertQuery[:returningIdx]
		returningPart := insertQuery[returningIdx+9:] // Skip "RETURNING "
		returningPart = strings.TrimSpace(returningPart)
		
		// Execute INSERT without RETURNING
		_, err := exec.Exec(ctx, insertPart, args...)
		if err != nil {
			return zero, fmt.Errorf("typedb: InsertAndReturn failed to execute INSERT: %w", err)
		}
		
		// Parse table name from INSERT query to build SELECT
		// INSERT INTO table_name ... -> SELECT ... FROM table_name WHERE ...
		// For Oracle, we'll use MAX(id) approach since we can't easily build WHERE clause
		// But we need to return the columns specified in RETURNING
		// Actually, we can query back using MAX(id) and then SELECT those columns
		
		// Extract table name from INSERT query
		insertUpper := strings.ToUpper(insertPart)
		tableStart := strings.Index(insertUpper, "INTO ")
		if tableStart == -1 {
			return zero, fmt.Errorf("typedb: InsertAndReturn failed to parse table name from query")
		}
		tableStart += 5 // Skip "INTO "
		tableEnd := strings.Index(insertUpper[tableStart:], " ")
		if tableEnd == -1 {
			tableEnd = len(insertUpper) - tableStart
		}
		tableName := insertPart[tableStart : tableStart+tableEnd]
		tableName = strings.TrimSpace(tableName)
		
		// Build SELECT query using RETURNING columns and MAX(id) to get the last inserted row
		// Parse RETURNING columns (handle "RETURNING col1, col2, ...")
		returningCols := strings.Split(returningPart, ",")
		for i := range returningCols {
			returningCols[i] = strings.TrimSpace(returningCols[i])
		}
		
		// Find ID column (usually first or named 'id')
		var idCol string
		for _, col := range returningCols {
			colUpper := strings.ToUpper(strings.TrimSpace(col))
			if colUpper == "ID" || strings.HasSuffix(colUpper, ".ID") {
				idCol = strings.TrimSpace(col)
				break
			}
		}
		if idCol == "" && len(returningCols) > 0 {
			// Use first column as ID
			idCol = returningCols[0]
		}
		
		// Query MAX(id) then SELECT the row
		maxIDQuery := fmt.Sprintf("SELECT MAX(%s) as max_id FROM %s", idCol, tableName)
		idRow, err := exec.QueryRowMap(ctx, maxIDQuery)
		if err != nil {
			return zero, fmt.Errorf("typedb: InsertAndReturn failed to get ID: %w", err)
		}
		maxID, ok := idRow["max_id"]
		if !ok {
			maxID, ok = idRow["MAX_ID"]
		}
		if !ok || maxID == nil {
			return zero, fmt.Errorf("typedb: InsertAndReturn failed to get inserted ID")
		}
		
		// Build SELECT query for the returned columns
		selectCols := strings.Join(returningCols, ", ")
		selectQuery := fmt.Sprintf("SELECT %s FROM %s WHERE %s = :1", selectCols, tableName, idCol)
		row, err := exec.QueryRowMap(ctx, selectQuery, maxID)
		if err != nil {
			return zero, fmt.Errorf("typedb: InsertAndReturn failed to query returned row: %w", err)
		}
		
		// Deserialize the returned row into a new model instance
		model, err := deserializeForType[T](row)
		if err != nil {
			return zero, fmt.Errorf("typedb: InsertAndReturn deserialization failed: %w", err)
		}
		
		return model, nil
	}

	// Execute the INSERT with RETURNING/OUTPUT and get the returned row
	row, err := exec.QueryRowMap(ctx, insertQuery, args...)
	if err != nil {
		return zero, fmt.Errorf("typedb: InsertAndReturn failed: %w", err)
	}

	// Deserialize the returned row into a new model instance
	model, err := deserializeForType[T](row)
	if err != nil {
		return zero, fmt.Errorf("typedb: InsertAndReturn deserialization failed: %w", err)
	}

	return model, nil
}

// insertedId is an internal model used by InsertAndGetId to retrieve just the ID.
// Note: This uses int64 to support all integer ID types (SMALLINT/int16, INTEGER/int32, BIGINT/int64).
// The deserialization layer automatically converts smaller integer types to int64.
// For type-safe ID retrieval with specific types, use InsertAndReturn with your own model.
type insertedId struct {
	Model
	ID int64 `db:"id"`
}

func init() {
	// Register insertedId so Model.deserialize can find the outer struct type
	// when called directly on an insertedId instance.
	RegisterModel[*insertedId]()
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
	result, err := InsertAndReturn[*insertedId](ctx, exec, insertQuery, args...)
	if err != nil {
		return 0, err
	}

	return result.ID, nil
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
	// Use RETURNING ... INTO :id /*LastInsertId*/ with Exec() and LastInsertId()
	if driverNameLower == "oracle" {
		quotedPK := quoteIdentifier(driverName, primaryKeyColumn)
		// Oracle requires RETURNING ... INTO :bindvar with /*LastInsertId*/ comment
		insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING %s INTO :%s /*LastInsertId*/",
			quotedTableName,
			strings.Join(quotedColumns, ", "),
			strings.Join(placeholders, ", "),
			quotedPK,
			strings.ToLower(primaryKeyColumn))
		
		result, err := exec.Exec(ctx, insertQuery, values...)
		if err != nil {
			return fmt.Errorf("typedb: Insert failed: %w", err)
		}

		// Use LastInsertId() which is thread-safe (uses connection-specific state)
		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("typedb: Insert failed to get last insert ID: %w", err)
		}

		// Set primary key on model
		return setFieldValue(model, primaryField.Name, id)
	}

	// For databases with RETURNING support (PostgreSQL, SQLite, SQL Server)
	// SQL Server OUTPUT clause comes after VALUES
	var insertQuery string
	if driverNameLower == "sqlserver" || driverNameLower == "mssql" {
		insertQuery = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)%s",
			quotedTableName,
			strings.Join(quotedColumns, ", "),
			strings.Join(placeholders, ", "),
			returningClause)
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
