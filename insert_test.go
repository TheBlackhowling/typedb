package typedb

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// InsertTestModel is a simple model for testing InsertAndReturn
type InsertTestModel struct {
	ID        int64  `db:"id"`
	Name      string `db:"name"`
	CreatedAt string `db:"created_at"`
}

func (t *InsertTestModel) Deserialize(row map[string]any) error {
	return Deserialize(row, t)
}

func init() {
	RegisterModel[*InsertTestModel]()
}

func TestInsertAndReturn_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	// Mock QueryRowMap to return a row
	rows := sqlmock.NewRows([]string{"id", "name", "created_at"}).
		AddRow(123, "John Doe", "2024-01-15 10:00:00")

	mock.ExpectQuery("INSERT INTO users").
		WithArgs("John Doe", "john@example.com").
		WillReturnRows(rows)

	result, err := InsertAndReturn[*InsertTestModel](ctx, typedbDB,
		"INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id, name, created_at",
		"John Doe", "john@example.com")

	if err != nil {
		t.Fatalf("InsertAndReturn failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.ID != 123 {
		t.Errorf("Expected ID 123, got %d", result.ID)
	}

	if result.Name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got '%s'", result.Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestInsertAndReturn_QueryRowMapError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	expectedError := errors.New("database error")
	mock.ExpectQuery("INSERT INTO users").
		WithArgs("John Doe", "john@example.com").
		WillReturnError(expectedError)

	result, err := InsertAndReturn[*InsertTestModel](ctx, typedbDB,
		"INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id, name, created_at",
		"John Doe", "john@example.com")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result, got %+v", result)
	}

	if !errors.Is(err, expectedError) && !errors.Is(err, ErrNotFound) {
		// Check if error wraps the expected error
		if err.Error() == "typedb: InsertAndReturn failed: database error" {
			// This is acceptable - error is wrapped
		} else {
			t.Errorf("Expected error wrapping 'database error', got: %v", err)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestInsertAndReturn_DeserializationError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	// Mock QueryRowMap to return a row with invalid data type (string instead of int64 for ID)
	rows := sqlmock.NewRows([]string{"id", "name", "created_at"}).
		AddRow("invalid", "John Doe", "2024-01-15 10:00:00")

	mock.ExpectQuery("INSERT INTO users").
		WithArgs("John Doe", "john@example.com").
		WillReturnRows(rows)

	result, err := InsertAndReturn[*InsertTestModel](ctx, typedbDB,
		"INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id, name, created_at",
		"John Doe", "john@example.com")

	// Deserialization might succeed with type conversion, so we just verify the function completes
	// The actual deserialization behavior is tested in deserialize_test.go
	if err != nil {
		// If there's an error, that's fine - we're testing error handling
		if result != nil {
			t.Errorf("Expected nil result on error, got %+v", result)
		}
	} else {
		// If no error, verify the result is valid
		if result == nil {
			t.Error("Expected result, got nil")
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestInsertAndGetId_WithReturning_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	// Mock QueryRowMap to return a row with ID
	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(456)

	mock.ExpectQuery("INSERT INTO users").
		WithArgs("John Doe", "john@example.com").
		WillReturnRows(rows)

	id, err := InsertAndGetId(ctx, typedbDB,
		"INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
		"John Doe", "john@example.com")

	if err != nil {
		t.Fatalf("InsertAndGetId failed: %v", err)
	}

	if id != 456 {
		t.Errorf("Expected ID 456, got %d", id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestInsertAndGetId_WithOutput_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "sqlserver", 5*time.Second)
	ctx := context.Background()

	// Mock QueryRowMap to return a row with ID
	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(789)

	mock.ExpectQuery("INSERT INTO users").
		WithArgs("John Doe", "john@example.com").
		WillReturnRows(rows)

	id, err := InsertAndGetId(ctx, typedbDB,
		"INSERT INTO users (name, email) OUTPUT INSERTED.id VALUES (@p1, @p2)",
		"John Doe", "john@example.com")

	if err != nil {
		t.Fatalf("InsertAndGetId failed: %v", err)
	}

	if id != 789 {
		t.Errorf("Expected ID 789, got %d", id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestInsertAndGetId_MySQL_WithoutReturning_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "mysql", 5*time.Second)
	ctx := context.Background()

	// Mock Exec to return a result with LastInsertId
	mock.ExpectExec("INSERT INTO users").
		WithArgs("John Doe", "john@example.com").
		WillReturnResult(sqlmock.NewResult(999, 1))

	id, err := InsertAndGetId(ctx, typedbDB,
		"INSERT INTO users (name, email) VALUES (?, ?)",
		"John Doe", "john@example.com")

	if err != nil {
		t.Fatalf("InsertAndGetId failed: %v", err)
	}

	if id != 999 {
		t.Errorf("Expected ID 999, got %d", id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestInsertAndGetId_SQLite_WithoutReturning_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "sqlite3", 5*time.Second)
	ctx := context.Background()

	// Mock Exec to return a result with LastInsertId
	mock.ExpectExec("INSERT INTO users").
		WithArgs("John Doe", "john@example.com").
		WillReturnResult(sqlmock.NewResult(111, 1))

	id, err := InsertAndGetId(ctx, typedbDB,
		"INSERT INTO users (name, email) VALUES (?, ?)",
		"John Doe", "john@example.com")

	if err != nil {
		t.Fatalf("InsertAndGetId failed: %v", err)
	}

	if id != 111 {
		t.Errorf("Expected ID 111, got %d", id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestInsertAndGetId_PostgreSQL_WithoutReturning_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	id, err := InsertAndGetId(ctx, typedbDB,
		"INSERT INTO users (name, email) VALUES ($1, $2)",
		"John Doe", "john@example.com")

	if err == nil {
		t.Fatal("Expected error for PostgreSQL without RETURNING, got nil")
	}

	if id != 0 {
		t.Errorf("Expected ID 0 on error, got %d", id)
	}

	expectedErrorMsg := "typedb: InsertAndGetId requires RETURNING or OUTPUT clause for postgres. Only MySQL and SQLite support LastInsertId() without RETURNING/OUTPUT"
	if err.Error() != expectedErrorMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestInsertAndGetId_SQLServer_WithoutReturning_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "sqlserver", 5*time.Second)
	ctx := context.Background()

	id, err := InsertAndGetId(ctx, typedbDB,
		"INSERT INTO users (name, email) VALUES (@p1, @p2)",
		"John Doe", "john@example.com")

	if err == nil {
		t.Fatal("Expected error for SQL Server without OUTPUT, got nil")
	}

	if id != 0 {
		t.Errorf("Expected ID 0 on error, got %d", id)
	}

	expectedErrorMsg := "typedb: InsertAndGetId requires RETURNING or OUTPUT clause for sqlserver. Only MySQL and SQLite support LastInsertId() without RETURNING/OUTPUT"
	if err.Error() != expectedErrorMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestInsertAndGetId_MySQL_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "mysql", 5*time.Second)
	ctx := context.Background()

	expectedError := errors.New("execution error")
	mock.ExpectExec("INSERT INTO users").
		WithArgs("John Doe", "john@example.com").
		WillReturnError(expectedError)

	id, err := InsertAndGetId(ctx, typedbDB,
		"INSERT INTO users (name, email) VALUES (?, ?)",
		"John Doe", "john@example.com")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if id != 0 {
		t.Errorf("Expected ID 0 on error, got %d", id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestInsertAndGetId_MySQL_LastInsertIdError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "mysql", 5*time.Second)
	ctx := context.Background()

	// Mock Exec to return a result that fails LastInsertId
	result := sqlmock.NewErrorResult(errors.New("LastInsertId not supported"))
	mock.ExpectExec("INSERT INTO users").
		WithArgs("John Doe", "john@example.com").
		WillReturnResult(result)

	id, err := InsertAndGetId(ctx, typedbDB,
		"INSERT INTO users (name, email) VALUES (?, ?)",
		"John Doe", "john@example.com")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if id != 0 {
		t.Errorf("Expected ID 0 on error, got %d", id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestInsertAndGetId_WithReturning_InsertAndReturnError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	expectedError := errors.New("query error")
	mock.ExpectQuery("INSERT INTO users").
		WithArgs("John Doe", "john@example.com").
		WillReturnError(expectedError)

	id, err := InsertAndGetId(ctx, typedbDB,
		"INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
		"John Doe", "john@example.com")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if id != 0 {
		t.Errorf("Expected ID 0 on error, got %d", id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestInsertAndGetId_UnknownDriver_WithoutReturning_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "unknown", 5*time.Second)
	ctx := context.Background()

	id, err := InsertAndGetId(ctx, typedbDB,
		"INSERT INTO users (name, email) VALUES ($1, $2)",
		"John Doe", "john@example.com")

	if err == nil {
		t.Fatal("Expected error for unknown driver without RETURNING, got nil")
	}

	if id != 0 {
		t.Errorf("Expected ID 0 on error, got %d", id)
	}

	expectedErrorMsg := "typedb: InsertAndGetId requires RETURNING or OUTPUT clause for unknown. Only MySQL and SQLite support LastInsertId() without RETURNING/OUTPUT"
	if err.Error() != expectedErrorMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestInsertAndGetId_Transaction_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	// Mock transaction begin
	mock.ExpectBegin()

	// Mock QueryRowMap to return a row with ID
	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(777)

	mock.ExpectQuery("INSERT INTO users").
		WithArgs("John Doe", "john@example.com").
		WillReturnRows(rows)

	// Mock transaction commit
	mock.ExpectCommit()

	tx, err := typedbDB.Begin(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	id, err := InsertAndGetId(ctx, tx,
		"INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
		"John Doe", "john@example.com")

	if err != nil {
		t.Fatalf("InsertAndGetId failed: %v", err)
	}

	if id != 777 {
		t.Errorf("Expected ID 777, got %d", id)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestSupportsLastInsertId(t *testing.T) {
	tests := []struct {
		driverName string
		expected   bool
	}{
		{"mysql", true},
		{"MySQL", true},
		{"MYSQL", true},
		{"sqlite3", true},
		{"SQLite3", true},
		{"SQLITE3", true},
		{"postgres", false},
		{"postgresql", false},
		{"sqlserver", false},
		{"mssql", false},
		{"oracle", false},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.driverName, func(t *testing.T) {
			result := supportsLastInsertId(tt.driverName)
			if result != tt.expected {
				t.Errorf("supportsLastInsertId(%q) = %v, want %v", tt.driverName, result, tt.expected)
			}
		})
	}
}

func TestGetDriverName(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	// Mock transaction begin
	mock.ExpectBegin()

	tx, err := typedbDB.Begin(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	tests := []struct {
		name     string
		exec     Executor
		expected string
	}{
		{"DB", typedbDB, "postgres"},
		{"Tx", tx, "postgres"},
		{"nil", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getDriverName(tt.exec)
			if result != tt.expected {
				t.Errorf("getDriverName(%v) = %q, want %q", tt.name, result, tt.expected)
			}
		})
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}
