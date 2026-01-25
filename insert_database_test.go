package typedb

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestInsert_SQLite_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "sqlite3", 5*time.Second)
	ctx := context.Background()

	user := &InsertModel{Name: "John", Email: "john@example.com"}

	rows := sqlmock.NewRows([]string{"id"}).AddRow(123)

	mock.ExpectQuery(`INSERT INTO "users" \("name", "email"\) VALUES \(\?, \?\) RETURNING "id"`).
		WithArgs("John", "john@example.com").
		WillReturnRows(rows)

	err = Insert(ctx, typedbDB, user)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	if user.ID != 123 {
		t.Errorf("Expected ID 123, got %d", user.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestInsert_SQLServer_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "sqlserver", 5*time.Second)
	ctx := context.Background()

	user := &InsertModel{Name: "John", Email: "john@example.com"}

	rows := sqlmock.NewRows([]string{"id"}).AddRow(123)

	// SQL Server OUTPUT clause comes BEFORE VALUES, not after
	mock.ExpectQuery(`INSERT INTO \[users\] \(\[name\], \[email\]\) OUTPUT INSERTED\.\[id\] VALUES \(@p1, @p2\)`).
		WithArgs("John", "john@example.com").
		WillReturnRows(rows)

	err = Insert(ctx, typedbDB, user)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	if user.ID != 123 {
		t.Errorf("Expected ID 123, got %d", user.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestInsert_Oracle_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "oracle", 5*time.Second)
	ctx := context.Background()

	user := &InsertModel{Name: "John", Email: "john@example.com"}

	// Oracle uses Exec with RETURNING ... INTO :3 with sql.Out parameter
	// The third argument is sql.Out{Dest: &id} for the RETURNING INTO clause
	// sqlmock doesn't populate sql.Out automatically, so we wrap the DB
	// to intercept Exec calls and manually populate the sql.Out value
	result := sqlmock.NewResult(0, 1)
	mock.ExpectExec(`INSERT INTO "USERS" \("NAME", "EMAIL"\) VALUES \(:1, :2\) RETURNING "ID" INTO :3`).
		WithArgs("John", "john@example.com", sqlmock.AnyArg()).
		WillReturnResult(result)
	
	// Wrap the DB to intercept Exec calls and populate sql.Out
	wrappedDB := &oracleTestWrapper{DB: typedbDB, testID: 123}
	
	err = Insert(ctx, wrappedDB, user)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	if user.ID != 123 {
		t.Errorf("Expected ID 123, got %d", user.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestInsert_UnknownDriver_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "unknown", 5*time.Second)
	ctx := context.Background()

	user := &InsertModel{Name: "John", Email: "john@example.com"}

	rows := sqlmock.NewRows([]string{"id"}).AddRow(123)

	mock.ExpectQuery(`INSERT INTO "users" \("name", "email"\) VALUES \(\$1, \$2\) RETURNING "id"`).
		WithArgs("John", "john@example.com").
		WillReturnRows(rows)

	err = Insert(ctx, typedbDB, user)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	if user.ID != 123 {
		t.Errorf("Expected ID 123, got %d", user.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestInsert_MySQL_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "mysql", 5*time.Second)
	ctx := context.Background()

	user := &InsertModel{Name: "John", Email: "john@example.com"}

	mock.ExpectExec("INSERT INTO `users`").
		WithArgs("John", "john@example.com").
		WillReturnError(fmt.Errorf("database error"))

	err = Insert(ctx, typedbDB, user)
	if err == nil {
		t.Fatal("Expected error from Exec")
	}

	if !strings.Contains(err.Error(), "Insert failed") {
		t.Errorf("Expected error about Insert failed, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestInsert_MySQL_LastInsertIdError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "mysql", 5*time.Second)
	ctx := context.Background()

	user := &InsertModel{Name: "John", Email: "john@example.com"}

	result := sqlmock.NewErrorResult(fmt.Errorf("LastInsertId error"))
	mock.ExpectExec("INSERT INTO `users`").
		WithArgs("John", "john@example.com").
		WillReturnResult(result)

	err = Insert(ctx, typedbDB, user)
	if err == nil {
		t.Fatal("Expected error from LastInsertId")
	}

	if !strings.Contains(err.Error(), "last insert ID") {
		t.Errorf("Expected error about last insert ID, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestInsert_QueryRowMapError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	user := &InsertModel{Name: "John", Email: "john@example.com"}

	mock.ExpectQuery(`INSERT INTO "users"`).
		WithArgs("John", "john@example.com").
		WillReturnError(fmt.Errorf("query error"))

	err = Insert(ctx, typedbDB, user)
	if err == nil {
		t.Fatal("Expected error from QueryRowMap")
	}

	if !strings.Contains(err.Error(), "Insert failed") {
		t.Errorf("Expected error about Insert failed, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestInsert_PrimaryKeyNotFound_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	user := &InsertModel{Name: "John", Email: "john@example.com"}

	// Return row without "id" column
	rows := sqlmock.NewRows([]string{"wrong_column"}).AddRow(123)

	mock.ExpectQuery(`INSERT INTO "users"`).
		WithArgs("John", "john@example.com").
		WillReturnRows(rows)

	err = Insert(ctx, typedbDB, user)
	if err == nil {
		t.Fatal("Expected error for missing primary key column")
	}

	if !strings.Contains(err.Error(), "RETURNING clause did not return primary key") {
		t.Errorf("Expected error about RETURNING clause, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestInsert_PrimaryKeyUppercase_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	user := &InsertModel{Name: "John", Email: "john@example.com"}

	// Return row with uppercase "ID" instead of lowercase "id"
	rows := sqlmock.NewRows([]string{"ID"}).AddRow(123)

	mock.ExpectQuery(`INSERT INTO "users"`).
		WithArgs("John", "john@example.com").
		WillReturnRows(rows)

	err = Insert(ctx, typedbDB, user)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	if user.ID != 123 {
		t.Errorf("Expected ID 123, got %d", user.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestInsert_Oracle_InsertAndReturnError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "oracle", 5*time.Second)
	ctx := context.Background()

	user := &InsertModel{Name: "John", Email: "john@example.com"}

	// Oracle uses Exec with RETURNING ... INTO :3 with sql.Out parameter
	mock.ExpectExec(`INSERT INTO "USERS" \("NAME", "EMAIL"\) VALUES \(:1, :2\) RETURNING "ID" INTO :3`).
		WithArgs("John", "john@example.com", sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("Insert error"))

	err = Insert(ctx, typedbDB, user)
	if err == nil {
		t.Fatal("Expected error from Insert")
	}

	if !strings.Contains(err.Error(), "Insert failed") {
		t.Errorf("Expected error about Insert failed, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

// InsertModelWithDbInsertFalse is a model with a field that should be skipped during Insert
type InsertModelWithDbInsertFalse struct {
	Model
	ID        int64  `db:"id" load:"primary"`
	Name      string `db:"name"`
	Email     string `db:"email"`
	CreatedAt string `db:"created_at" dbInsert:"false"` // Should be skipped in Insert but can be read/updated
}

func (m *InsertModelWithDbInsertFalse) TableName() string {
	return "users"
}

func (m *InsertModelWithDbInsertFalse) QueryByID() string {
	return "SELECT id, name, email, updated_at FROM users WHERE id = $1"
}

func init() {
	RegisterModel[*InsertModelWithDbInsertFalse]()
}

// InsertModelWithDbUpdateFalse is a model with a field that should be skipped during Update
type InsertModelWithDbUpdateFalse struct {
	Model
	ID        int64  `db:"id" load:"primary"`
	Name      string `db:"name"`
	Email     string `db:"email"`
	UpdatedAt string `db:"updated_at" dbUpdate:"false"` // Should be skipped in Update but can be read/inserted
}

func (m *InsertModelWithDbUpdateFalse) TableName() string {
	return "users"
}

func (m *InsertModelWithDbUpdateFalse) QueryByID() string {
	return "SELECT id, name, email, updated_at FROM users WHERE id = $1"
}

func init() {
	RegisterModel[*InsertModelWithDbUpdateFalse]()
}

// InsertModelWithBothTags is a model with a field that should be skipped in both Insert and Update
type InsertModelWithBothTags struct {
	Model
	ID        int64  `db:"id" load:"primary"`
	Name      string `db:"name"`
	Email     string `db:"email"`
	CreatedAt string `db:"created_at" dbInsert:"false" dbUpdate:"false"` // Should be skipped in Insert and Update but can be read
}

func (m *InsertModelWithBothTags) TableName() string {
	return "users"
}

func (m *InsertModelWithBothTags) QueryByID() string {
	return "SELECT id, name, email, created_at FROM users WHERE id = $1"
}

func init() {
	RegisterModel[*InsertModelWithBothTags]()
}

func TestInsert_DbInsertFalse_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	user := &InsertModelWithDbInsertFalse{
		ID:        0, // Will be set by Insert
		Name:      "John",
		Email:     "john@example.com",
		CreatedAt: "2024-01-01", // Should be skipped due to dbInsert:"false" tag
	}

	mock.ExpectQuery(`INSERT INTO "users" \("name", "email"\) VALUES \(\$1, \$2\) RETURNING "id"`).
		WithArgs("John", "john@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(123))

	err = Insert(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Insert() error = %v, want nil", err)
	}

	if user.ID != 123 {
		t.Errorf("Insert() ID = %v, want 123", user.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestInsert_DbUpdateFalse_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	// UpdatedAt should be included in INSERT even though it has dbUpdate:"false"
	user := &InsertModelWithDbUpdateFalse{
		ID:        0, // Will be set by Insert
		Name:      "John",
		Email:     "john@example.com",
		UpdatedAt: "2024-01-01", // Should be included in Insert (dbUpdate:"false" doesn't affect Insert)
	}

	mock.ExpectQuery(`INSERT INTO "users" \("name", "email", "updated_at"\) VALUES \(\$1, \$2, \$3\) RETURNING "id"`).
		WithArgs("John", "john@example.com", "2024-01-01").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(123))

	err = Insert(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Insert() error = %v, want nil", err)
	}

	if user.ID != 123 {
		t.Errorf("Insert() ID = %v, want 123", user.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestInsert_BothTagsFalse_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	// CreatedAt should be excluded from INSERT due to dbInsert:"false" (dbUpdate:"false" doesn't affect Insert)
	user := &InsertModelWithBothTags{
		ID:        0, // Will be set by Insert
		Name:      "John",
		Email:     "john@example.com",
		CreatedAt: "2024-01-01", // Should be skipped due to dbInsert:"false" tag
	}

	mock.ExpectQuery(`INSERT INTO "users" \("name", "email"\) VALUES \(\$1, \$2\) RETURNING "id"`).
		WithArgs("John", "john@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(123))

	err = Insert(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Insert() error = %v, want nil", err)
	}

	if user.ID != 123 {
		t.Errorf("Insert() ID = %v, want 123", user.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}
