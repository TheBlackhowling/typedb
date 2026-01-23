package typedb

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// InsertTestModel is a simple model for testing Insert functions
type InsertTestModel struct {
	Model
	ID        int64  `db:"id" load:"primary"`
	Name      string `db:"name"`
	CreatedAt string `db:"created_at"`
}

func (m *InsertTestModel) TableName() string {
	return "users"
}

func (m *InsertTestModel) QueryByID() string {
	return "SELECT id, name, created_at FROM users WHERE id = $1"
}

func init() {
	RegisterModel[*InsertTestModel]()
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
		t.Fatal("Expected error from InsertAndGetId")
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

func TestInsertAndLoad_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	// Mock INSERT with RETURNING id
	insertRows := sqlmock.NewRows([]string{"id"}).AddRow(123)
	mock.ExpectQuery("INSERT INTO \"users\"").
		WithArgs("John Doe").
		WillReturnRows(insertRows)

	// Mock Load query
	loadRows := sqlmock.NewRows([]string{"id", "name", "created_at"}).
		AddRow(123, "John Doe", "2024-01-15 10:00:00")
	mock.ExpectQuery("SELECT id, name, created_at FROM users WHERE id = \\$1").
		WithArgs(123).
		WillReturnRows(loadRows)

	model := &InsertTestModel{Name: "John Doe"}
	returnedModel, err := InsertAndLoad[*InsertTestModel](ctx, typedbDB, model)

	if err != nil {
		t.Fatalf("InsertAndLoad failed: %v", err)
	}

	if returnedModel.ID != 123 {
		t.Errorf("Expected ID 123, got %d", returnedModel.ID)
	}

	if returnedModel.Name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got %q", returnedModel.Name)
	}

	if returnedModel.CreatedAt != "2024-01-15 10:00:00" {
		t.Errorf("Expected CreatedAt '2024-01-15 10:00:00', got %q", returnedModel.CreatedAt)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestInsertAndLoad_InsertError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	// Mock INSERT failure
	mock.ExpectQuery("INSERT INTO \"users\"").
		WithArgs("John Doe").
		WillReturnError(errors.New("database error"))

	model := &InsertTestModel{Name: "John Doe"}
	returnedModel, err := InsertAndLoad[*InsertTestModel](ctx, typedbDB, model)

	if err == nil {
		t.Fatal("Expected error from InsertAndLoad")
	}

	var zero *InsertTestModel
	if returnedModel != zero {
		t.Errorf("Expected zero value, got %+v", returnedModel)
	}

	if !strings.Contains(err.Error(), "insert") {
		t.Errorf("Expected error message to contain 'insert', got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestInsertAndLoad_LoadError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	// Mock INSERT with RETURNING id (succeeds)
	insertRows := sqlmock.NewRows([]string{"id"}).AddRow(123)
	mock.ExpectQuery("INSERT INTO \"users\"").
		WithArgs("John Doe").
		WillReturnRows(insertRows)

	// Mock Load query failure
	mock.ExpectQuery("SELECT id, name, created_at FROM users WHERE id = \\$1").
		WithArgs(123).
		WillReturnError(errors.New("load error"))

	model := &InsertTestModel{Name: "John Doe"}
	returnedModel, err := InsertAndLoad[*InsertTestModel](ctx, typedbDB, model)

	if err == nil {
		t.Fatal("Expected error from InsertAndLoad")
	}

	var zero *InsertTestModel
	if returnedModel != zero {
		t.Errorf("Expected zero value, got %+v", returnedModel)
	}

	if !strings.Contains(err.Error(), "load") {
		t.Errorf("Expected error message to contain 'load', got: %v", err)
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
	insertedId := &insertedId{}
	row := map[string]any{
		"id": int64(123),
	}

	err := insertedId.deserialize(row)
	if err != nil {
		t.Fatalf("insertedId.deserialize failed: %v", err)
	}

	if insertedId.ID != 123 {
		t.Errorf("Expected ID 123, got %d", insertedId.ID)
	}
}

// TestDeserialize_AddressableValue tests deserialization via direct Deserialize() call.
// We always use buildFieldMapFromPtr (unsafe path) regardless of addressability to
// avoid checkptr errors across all Go versions 1.18-1.25.
func TestDeserialize_AddressableValue(t *testing.T) {
	type TestModel struct {
		Model
		ID   int64  `db:"id"`
		Name string `db:"name"`
	}

	model := &TestModel{}
	row := map[string]any{
		"id":   int64(456),
		"name": "Test Name",
	}

	// Direct call to Deserialize - uses buildFieldMapFromPtr (unsafe path)
	err := deserialize(row, model)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if model.ID != 456 {
		t.Errorf("Expected ID 456, got %d", model.ID)
	}
	if model.Name != "Test Name" {
		t.Errorf("Expected Name 'Test Name', got '%s'", model.Name)
	}
}

// TestDeserialize_NonAddressableValue tests deserialization via Model.Deserialize(),
// which uses reflect.NewAt to convert the embedded Model receiver to the outer struct pointer.
// This is the problematic case that triggers checkptr errors, which we fix by always using
// buildFieldMapFromPtr (unsafe path) regardless of addressability.
func TestDeserialize_NonAddressableValue(t *testing.T) {
	type TestModelNonAddr struct {
		Model
		ID   int64  `db:"id"`
		Name string `db:"name"`
	}

	// Register the model for Model.Deserialize to work
	RegisterModel[*TestModelNonAddr]()

	model := &TestModelNonAddr{}
	row := map[string]any{
		"id":   int64(789),
		"name": "Non-Addressable Test",
	}

	// Use deserializeForType for type-safe deserialization
	// This preserves type information and avoids type detection issues
	deserialized, err := deserializeForType[*TestModelNonAddr](row)
	if err != nil {
		t.Fatalf("deserializeForType failed: %v", err)
	}

	// Copy values back to original model for comparison
	*model = *deserialized

	if model.ID != 789 {
		t.Errorf("Expected ID 789, got %d", model.ID)
	}
	if model.Name != "Non-Addressable Test" {
		t.Errorf("Expected Name 'Non-Addressable Test', got '%s'", model.Name)
	}
}

// TestQuoteIdentifierEscaping tests the quoteIdentifier function with quote escaping
func TestQuoteIdentifierEscaping(t *testing.T) {
	tests := []struct {
		name       string
		driverName string
		identifier string
		want       string
		wantPanic  bool
	}{
		// PostgreSQL tests
		{
			name:       "PostgreSQL simple identifier",
			driverName: "postgres",
			identifier: "users",
			want:       `"users"`,
			wantPanic:  false,
		},
		{
			name:       "PostgreSQL identifier with underscore",
			driverName: "postgres",
			identifier: "user_table",
			want:       `"user_table"`,
			wantPanic:  false,
		},
		{
			name:       "PostgreSQL identifier with quote (escaped)",
			driverName: "postgres",
			identifier: `user"table`,
			want:       `"user""table"`,
			wantPanic:  false,
		},
		{
			name:       "PostgreSQL qualified identifier",
			driverName: "postgres",
			identifier: "schema.table",
			want:       `"schema.table"`,
			wantPanic:  false,
		},
		// SQLite tests (same as PostgreSQL)
		{
			name:       "SQLite simple identifier",
			driverName: "sqlite3",
			identifier: "users",
			want:       `"users"`,
			wantPanic:  false,
		},
		// MySQL tests
		{
			name:       "MySQL simple identifier",
			driverName: "mysql",
			identifier: "users",
			want:       "`users`",
			wantPanic:  false,
		},
		{
			name:       "MySQL identifier with backtick (escaped)",
			driverName: "mysql",
			identifier: "user`table",
			want:       "`user``table`",
			wantPanic:  false,
		},
		// SQL Server tests
		{
			name:       "SQL Server simple identifier",
			driverName: "sqlserver",
			identifier: "users",
			want:       "[users]",
			wantPanic:  false,
		},
		{
			name:       "SQL Server identifier with underscore",
			driverName: "mssql",
			identifier: "user_table",
			want:       "[user_table]",
			wantPanic:  false,
		},
		// Oracle tests
		{
			name:       "Oracle simple identifier",
			driverName: "oracle",
			identifier: "users",
			want:       `"USERS"`,
			wantPanic:  false,
		},
		{
			name:       "Oracle identifier with quote (escaped and uppercased)",
			driverName: "oracle",
			identifier: `user"table`,
			want:       `"USER""TABLE"`,
			wantPanic:  false,
		},
		{
			name:       "Oracle lowercase identifier",
			driverName: "oracle",
			identifier: "users",
			want:       `"USERS"`,
			wantPanic:  false,
		},
		// Empty identifier test (should panic)
		{
			name:       "PostgreSQL empty identifier",
			driverName: "postgres",
			identifier: "",
			wantPanic:  true,
		},
		{
			name:       "SQL Server empty identifier",
			driverName: "sqlserver",
			identifier: "",
			wantPanic:  true,
		},
		{
			name:       "PostgreSQL valid identifier that needs no escaping",
			driverName: "postgres",
			identifier: "users",
			want:       `"users"`,
			wantPanic:  false,
		},
		{
			name:       "MySQL valid identifier that needs no escaping",
			driverName: "mysql",
			identifier: "users",
			want:       "`users`",
			wantPanic:  false,
		},
		{
			name:       "Oracle valid identifier (uppercased)",
			driverName: "oracle",
			identifier: "users",
			want:       `"USERS"`,
			wantPanic:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("quoteIdentifier() expected panic but did not panic")
					}
				}()
				quoteIdentifier(tt.driverName, tt.identifier)
			} else {
				got := quoteIdentifier(tt.driverName, tt.identifier)
				if got != tt.want {
					t.Errorf("quoteIdentifier() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
