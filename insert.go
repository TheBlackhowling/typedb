package typedb

import (
	"context"
	"fmt"
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
	ID int64 `db:"id"`
}

func (i *InsertedId) Deserialize(row map[string]any) error {
	return Deserialize(row, i)
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
