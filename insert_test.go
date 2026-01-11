package typedb

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// InsertTestModel is a simple model for testing InsertAndReturn
type InsertTestModel struct {
	Model
	ID        int64  `db:"id"`
	Name      string `db:"name"`
	CreatedAt string `db:"created_at"`
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

	if result.ID != 123 {
		t.Errorf("Expected ID 123, got %d", result.ID)
	}

	if result.Name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got %q", result.Name)
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

	mock.ExpectQuery("INSERT INTO users").
		WithArgs("John Doe", "john@example.com").
		WillReturnError(errors.New("database error"))

	result, err := InsertAndReturn[*InsertTestModel](ctx, typedbDB,
		"INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id, name, created_at",
		"John Doe", "john@example.com")

	if err == nil {
		t.Fatal("Expected error from InsertAndReturn")
	}

	var zero *InsertTestModel
	if result != zero {
		t.Errorf("Expected zero value, got %+v", result)
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

	// Return invalid data type (string for int64 ID)
	rows := sqlmock.NewRows([]string{"id", "name", "created_at"}).
		AddRow("invalid", "John Doe", "2024-01-15 10:00:00")

	mock.ExpectQuery("INSERT INTO users").
		WithArgs("John Doe", "john@example.com").
		WillReturnRows(rows)

	result, err := InsertAndReturn[*InsertTestModel](ctx, typedbDB,
		"INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id, name, created_at",
		"John Doe", "john@example.com")

	if err == nil {
		t.Fatal("Expected deserialization error")
	}

	var zero *InsertTestModel
	if result != zero {
		t.Errorf("Expected zero value, got %+v", result)
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

	rows := sqlmock.NewRows([]string{"id"}).AddRow(123)

	mock.ExpectQuery("INSERT INTO users").
		WithArgs("John Doe", "john@example.com").
		WillReturnRows(rows)

	id, err := InsertAndGetId(ctx, typedbDB,
		"INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
		"John Doe", "john@example.com")

	if err != nil {
		t.Fatalf("InsertAndGetId failed: %v", err)
	}

	if id != 123 {
		t.Errorf("Expected ID 123, got %d", id)
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

	rows := sqlmock.NewRows([]string{"id"}).AddRow(123)

	mock.ExpectQuery("INSERT INTO users").
		WithArgs("John Doe", "john@example.com").
		WillReturnRows(rows)

	id, err := InsertAndGetId(ctx, typedbDB,
		"INSERT INTO users (name, email) OUTPUT INSERTED.id VALUES (@p1, @p2)",
		"John Doe", "john@example.com")

	if err != nil {
		t.Fatalf("InsertAndGetId failed: %v", err)
	}

	if id != 123 {
		t.Errorf("Expected ID 123, got %d", id)
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

	mock.ExpectExec("INSERT INTO users").
		WithArgs("John Doe", "john@example.com").
		WillReturnResult(sqlmock.NewResult(123, 1))

	id, err := InsertAndGetId(ctx, typedbDB,
		"INSERT INTO users (name, email) VALUES (?, ?)",
		"John Doe", "john@example.com")

	if err != nil {
		t.Fatalf("InsertAndGetId failed: %v", err)
	}

	if id != 123 {
		t.Errorf("Expected ID 123, got %d", id)
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

	mock.ExpectExec("INSERT INTO users").
		WithArgs("John Doe", "john@example.com").
		WillReturnResult(sqlmock.NewResult(123, 1))

	id, err := InsertAndGetId(ctx, typedbDB,
		"INSERT INTO users (name, email) VALUES (?, ?)",
		"John Doe", "john@example.com")

	if err != nil {
		t.Fatalf("InsertAndGetId failed: %v", err)
	}

	if id != 123 {
		t.Errorf("Expected ID 123, got %d", id)
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
		t.Fatal("Expected error for missing RETURNING clause")
	}

	expectedErrorMsg := "typedb: InsertAndGetId requires RETURNING or OUTPUT clause for postgres"
	if !strings.Contains(err.Error(), expectedErrorMsg) {
		t.Errorf("Expected error message containing %q, got: %v", expectedErrorMsg, err)
	}

	if id != 0 {
		t.Errorf("Expected ID 0 on error, got %d", id)
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
		t.Fatal("Expected error for missing OUTPUT clause")
	}

	expectedErrorMsg := "typedb: InsertAndGetId requires RETURNING or OUTPUT clause for sqlserver"
	if !strings.Contains(err.Error(), expectedErrorMsg) {
		t.Errorf("Expected error message containing %q, got: %v", expectedErrorMsg, err)
	}

	if id != 0 {
		t.Errorf("Expected ID 0 on error, got %d", id)
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

	mock.ExpectExec("INSERT INTO users").
		WithArgs("John Doe", "john@example.com").
		WillReturnError(errors.New("database error"))

	id, err := InsertAndGetId(ctx, typedbDB,
		"INSERT INTO users (name, email) VALUES (?, ?)",
		"John Doe", "john@example.com")

	if err == nil {
		t.Fatal("Expected error from Exec")
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

	result := sqlmock.NewErrorResult(errors.New("LastInsertId error"))
	mock.ExpectExec("INSERT INTO users").
		WithArgs("John Doe", "john@example.com").
		WillReturnResult(result)

	id, err := InsertAndGetId(ctx, typedbDB,
		"INSERT INTO users (name, email) VALUES (?, ?)",
		"John Doe", "john@example.com")

	if err == nil {
		t.Fatal("Expected error from LastInsertId")
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

	mock.ExpectQuery("INSERT INTO users").
		WithArgs("John Doe", "john@example.com").
		WillReturnError(errors.New("database error"))

	id, err := InsertAndGetId(ctx, typedbDB,
		"INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
		"John Doe", "john@example.com")

	if err == nil {
		t.Fatal("Expected error from InsertAndReturn")
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
		t.Fatal("Expected error for unknown driver without RETURNING")
	}

	expectedErrorMsg := "typedb: InsertAndGetId requires RETURNING or OUTPUT clause for unknown"
	if !strings.Contains(err.Error(), expectedErrorMsg) {
		t.Errorf("Expected error message containing %q, got: %v", expectedErrorMsg, err)
	}

	if id != 0 {
		t.Errorf("Expected ID 0 on error, got %d", id)
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

	mock.ExpectBegin()
	tx, err := typedbDB.Begin(ctx, nil)
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	rows := sqlmock.NewRows([]string{"id"}).AddRow(123)

	mock.ExpectQuery("INSERT INTO users").
		WithArgs("John Doe", "john@example.com").
		WillReturnRows(rows)

	id, err := InsertAndGetId(ctx, tx,
		"INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
		"John Doe", "john@example.com")

	if err != nil {
		t.Fatalf("InsertAndGetId failed: %v", err)
	}

	if id != 123 {
		t.Errorf("Expected ID 123, got %d", id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
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

	mock.ExpectBegin()
	tx, err := typedbDB.Begin(ctx, nil)
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
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

func TestInsertedId_Deserialize(t *testing.T) {
	insertedId := &InsertedId{}
	row := map[string]any{
		"id": int64(123),
	}

	err := insertedId.Deserialize(row)
	if err != nil {
		t.Fatalf("InsertedId.Deserialize failed: %v", err)
	}

	if insertedId.ID != 123 {
		t.Errorf("Expected ID 123, got %d", insertedId.ID)
	}
}
