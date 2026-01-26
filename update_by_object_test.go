package typedb

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// UpdateModelWithSkipTag is a model with a field that should be skipped during Update
type UpdateModelWithSkipTag struct {
	Model
	ID        int64  `db:"id" load:"primary"`
	Name      string `db:"name"`
	Email     string `db:"email"`
	CreatedAt string `db:"created_at" dbUpdate:"false"` // Should be skipped in Update but can be read/inserted
}

// UpdateModelWithDbInsertFalse is a model with a field that should be skipped during Insert
type UpdateModelWithDbInsertFalse struct {
	Model
	ID        int64  `db:"id" load:"primary"`
	Name      string `db:"name"`
	Email     string `db:"email"`
	UpdatedAt string `db:"updated_at" dbInsert:"false"` // Should be skipped in Insert but can be read/updated
}

func (m *UpdateModelWithDbInsertFalse) TableName() string {
	return "users"
}

func (m *UpdateModelWithDbInsertFalse) QueryByID() string {
	return "SELECT id, name, email, updated_at FROM users WHERE id = $1"
}

func init() {
	RegisterModel[*UpdateModelWithDbInsertFalse]()
}

// UpdateModelWithBothTags is a model with a field that should be skipped in both Insert and Update
type UpdateModelWithBothTags struct {
	Model
	ID        int64  `db:"id" load:"primary"`
	Name      string `db:"name"`
	Email     string `db:"email"`
	CreatedAt string `db:"created_at" dbInsert:"false" dbUpdate:"false"` // Should be skipped in Insert and Update but can be read
}

func (m *UpdateModelWithBothTags) TableName() string {
	return "users"
}

func (m *UpdateModelWithBothTags) QueryByID() string {
	return "SELECT id, name, email, created_at FROM users WHERE id = $1"
}

func init() {
	RegisterModel[*UpdateModelWithBothTags]()
}

func (m *UpdateModelWithSkipTag) TableName() string {
	return "users"
}

func (m *UpdateModelWithSkipTag) QueryByID() string {
	return "SELECT id, name, email, created_at FROM users WHERE id = $1"
}

func init() {
	RegisterModel[*UpdateModelWithSkipTag]()
}

// UpdateModelWithSkipTagValue is a model with dbUpdate:"false" tag
type UpdateModelWithSkipTagValue struct {
	Model
	ID        int64  `db:"id" load:"primary"`
	Name      string `db:"name"`
	Email     string `db:"email"`
	UpdatedAt string `db:"updated_at" dbUpdate:"false"` // Should be skipped in Update but can be read/inserted
}

func (m *UpdateModelWithSkipTagValue) TableName() string {
	return "users"
}

func (m *UpdateModelWithSkipTagValue) QueryByID() string {
	return "SELECT id, name, email, updated_at FROM users WHERE id = $1"
}

func init() {
	RegisterModel[*UpdateModelWithSkipTagValue]()
}

// UpdateModelWithAutoTimestamp is a model with a field that should be auto-populated with database timestamp
type UpdateModelWithAutoTimestamp struct {
	Model
	ID        int64  `db:"id" load:"primary"`
	Name      string `db:"name"`
	Email     string `db:"email"`
	UpdatedAt string `db:"updated_at" dbUpdate:"auto-timestamp"` // Should be auto-populated with database timestamp function
}

func (m *UpdateModelWithAutoTimestamp) TableName() string {
	return "users"
}

func (m *UpdateModelWithAutoTimestamp) QueryByID() string {
	return "SELECT id, name, email, updated_at FROM users WHERE id = $1"
}

func init() {
	RegisterModel[*UpdateModelWithAutoTimestamp]()
}

// UpdateModelWithPartialUpdate is a model with partial update enabled
type UpdateModelWithPartialUpdate struct {
	Model
	ID    int64  `db:"id" load:"primary"`
	Name  string `db:"name"`
	Email string `db:"email"`
}

func (m *UpdateModelWithPartialUpdate) TableName() string {
	return "users"
}

func (m *UpdateModelWithPartialUpdate) QueryByID() string {
	return "SELECT id, name, email FROM users WHERE id = $1"
}

func init() {
	RegisterModelWithOptions[*UpdateModelWithPartialUpdate](ModelOptions{PartialUpdate: true})
}

// TestUpdate_PartialUpdate_OnlyChangedFields tests that partial update only updates changed fields
func TestUpdate_PartialUpdate_OnlyChangedFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	// Create a user and simulate deserialization (save original copy)
	user := &UpdateModelWithPartialUpdate{
		ID:    123,
		Name:  "John",
		Email: "john@example.com",
	}

	// Simulate deserialization by saving original copy
	row := map[string]any{
		"id":    int64(123),
		"name":  "John",
		"email": "john@example.com",
	}
	if deserializeErr := deserialize(row, user); deserializeErr != nil {
		t.Fatalf("Failed to deserialize: %v", deserializeErr)
	}

	// Modify only name
	user.Name = "John Updated"
	// Email remains unchanged

	// Expect UPDATE to only include changed fields (name)
	mock.ExpectExec(`UPDATE "users" SET "name" = \$1 WHERE "id" = \$2`).
		WithArgs("John Updated", int64(123)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Update(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

// TestUpdate_PartialUpdate_NoOriginalCopy tests that partial update falls back to normal update when no original copy exists
func TestUpdate_PartialUpdate_NoOriginalCopy(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	// Create a user without deserialization (no original copy)
	user := &UpdateModelWithPartialUpdate{
		ID:    123,
		Name:  "John Updated",
		Email: "john.updated@example.com",
	}

	// Expect UPDATE to include all non-zero fields (fallback behavior)
	mock.ExpectExec(`UPDATE "users" SET "name" = \$1, "email" = \$2 WHERE "id" = \$3`).
		WithArgs("John Updated", "john.updated@example.com", int64(123)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Update(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

// TestUpdate_PartialUpdate_MultipleChangedFields tests that partial update includes all changed fields
func TestUpdate_PartialUpdate_MultipleChangedFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	// Create a user and simulate deserialization
	user := &UpdateModelWithPartialUpdate{
		ID:    123,
		Name:  "John",
		Email: "john@example.com",
	}

	row := map[string]any{
		"id":    int64(123),
		"name":  "John",
		"email": "john@example.com",
	}
	if deserializeErr := deserialize(row, user); deserializeErr != nil {
		t.Fatalf("Failed to deserialize: %v", deserializeErr)
	}

	// Modify both name and email
	user.Name = "John Updated"
	user.Email = "john.updated@example.com"

	// Expect UPDATE to include both changed fields
	mock.ExpectExec(`UPDATE "users" SET "name" = \$1, "email" = \$2 WHERE "id" = \$3`).
		WithArgs("John Updated", "john.updated@example.com", int64(123)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Update(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

// TestUpdate_NonPartialUpdate_AllNonZeroFields tests that non-partial update includes all non-zero fields
// even if they haven't changed (when model doesn't have partial update enabled)
func TestUpdate_NonPartialUpdate_AllNonZeroFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	// InsertModel doesn't have partial update enabled
	// When updating, ALL non-zero fields should be included, regardless of whether they changed
	user := &InsertModel{
		ID:    123,
		Name:  "John Updated",
		Email: "john@example.com", // This field is set but hasn't changed - should still be included
	}

	// Expect UPDATE to include ALL non-zero fields (both name and email)
	mock.ExpectExec(`UPDATE "users" SET "name" = \$1, "email" = \$2 WHERE "id" = \$3`).
		WithArgs("John Updated", "john@example.com", int64(123)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Update(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

// TestUpdate_NonPartialUpdate_OnlySetFields tests that non-partial update only includes fields that are set
func TestUpdate_NonPartialUpdate_OnlySetFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	// InsertModel doesn't have partial update enabled
	// Only set fields should be included (zero values are excluded)
	user := &InsertModel{
		ID:   123,
		Name: "John Updated",
		// Email is not set (zero value) - should be excluded
	}

	// Expect UPDATE to only include name (email is zero value, so excluded)
	mock.ExpectExec(`UPDATE "users" SET "name" = \$1 WHERE "id" = \$2`).
		WithArgs("John Updated", int64(123)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Update(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

// TestUpdate_PartialUpdate_UnchangedFieldsExcluded tests that partial update excludes unchanged fields
// This is the key difference from non-partial update
func TestUpdate_PartialUpdate_UnchangedFieldsExcluded(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	// Create a user and simulate deserialization (save original copy)
	user := &UpdateModelWithPartialUpdate{
		ID:    123,
		Name:  "John",
		Email: "john@example.com",
	}

	// Simulate deserialization by saving original copy
	row := map[string]any{
		"id":    int64(123),
		"name":  "John",
		"email": "john@example.com",
	}
	if deserializeErr := deserialize(row, user); deserializeErr != nil {
		t.Fatalf("Failed to deserialize: %v", deserializeErr)
	}

	// Modify only name, keep email unchanged
	user.Name = "John Updated"
	// Email remains unchanged - should NOT be included in UPDATE with partial update

	// Expect UPDATE to only include changed fields (name), NOT unchanged fields (email)
	mock.ExpectExec(`UPDATE "users" SET "name" = \$1 WHERE "id" = \$2`).
		WithArgs("John Updated", int64(123)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Update(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

// Update tests

func TestUpdate_PostgreSQL_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	user := &InsertModel{ID: 123, Name: "John Updated", Email: "john.updated@example.com"}

	mock.ExpectExec(`UPDATE "users" SET "name" = \$1, "email" = \$2 WHERE "id" = \$3`).
		WithArgs("John Updated", "john.updated@example.com", int64(123)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Update(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestUpdate_MySQL_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "mysql", 5*time.Second)
	ctx := context.Background()

	user := &InsertModel{ID: 123, Name: "John Updated"}

	mock.ExpectExec("UPDATE `users` SET `name` = \\? WHERE `id` = \\?").
		WithArgs("John Updated", int64(123)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Update(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestUpdate_SQLite_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "sqlite3", 5*time.Second)
	ctx := context.Background()

	user := &InsertModel{ID: 123, Email: "updated@example.com"}

	mock.ExpectExec(`UPDATE "users" SET "email" = \? WHERE "id" = \?`).
		WithArgs("updated@example.com", int64(123)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Update(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestUpdate_SQLServer_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "sqlserver", 5*time.Second)
	ctx := context.Background()

	user := &InsertModel{ID: 123, Name: "John Updated", Email: "john.updated@example.com"}

	mock.ExpectExec(`UPDATE \[users\] SET \[name\] = @p1, \[email\] = @p2 WHERE \[id\] = @p3`).
		WithArgs("John Updated", "john.updated@example.com", int64(123)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Update(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestUpdate_Oracle_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "oracle", 5*time.Second)
	ctx := context.Background()

	user := &InsertModel{ID: 123, Name: "John Updated"}

	mock.ExpectExec(`UPDATE "USERS" SET "NAME" = :1 WHERE "ID" = :2`).
		WithArgs("John Updated", int64(123)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Update(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestUpdate_NoTableName_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	user := &NoTableNameModel{ID: 123}

	err = Update(ctx, typedbDB, user)
	if err == nil {
		t.Error("Update() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "TableName") {
		t.Errorf("Update() error = %v, want error containing 'TableName'", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestUpdate_NoPrimaryKey_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	user := &NoPrimaryKeyModel{ID: 123}

	err = Update(ctx, typedbDB, user)
	if err == nil {
		t.Error("Update() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "primary") {
		t.Errorf("Update() error = %v, want error containing 'primary'", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestUpdate_PrimaryKeyNotSet_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	user := &InsertModel{Name: "John"} // ID is zero

	err = Update(ctx, typedbDB, user)
	if err == nil {
		t.Error("Update() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "primary key") || !strings.Contains(err.Error(), "set") {
		t.Errorf("Update() error = %v, want error about primary key not being set", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestUpdate_DotNotation_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	user := &JoinedModel{UserID: 123, Bio: "Updated bio"}

	err = Update(ctx, typedbDB, user)
	if err == nil {
		t.Error("Update() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "dot notation") {
		t.Errorf("Update() error = %v, want error containing 'dot notation'", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestUpdate_AllZeroFields_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	user := &InsertModel{ID: 123} // Only primary key set, no other fields

	err = Update(ctx, typedbDB, user)
	if err == nil {
		t.Error("Update() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "at least one non-nil field") {
		t.Errorf("Update() error = %v, want error about needing fields to update", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestUpdate_ZeroFieldsExcluded(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	user := &InsertModel{ID: 123, Name: "John Updated"} // Email is empty, should be excluded

	mock.ExpectExec(`UPDATE "users" SET "name" = \$1 WHERE "id" = \$2`).
		WithArgs("John Updated", int64(123)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Update(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestUpdate_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	user := &InsertModel{ID: 123, Name: "John Updated"}

	mock.ExpectExec(`UPDATE "users" SET "name" = \$1 WHERE "id" = \$2`).
		WithArgs("John Updated", int64(123)).
		WillReturnError(fmt.Errorf("database error"))

	err = Update(ctx, typedbDB, user)
	if err == nil {
		t.Error("Update() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "Update failed") {
		t.Errorf("Update() error = %v, want error containing 'Update failed'", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestUpdate_SkipTagDash_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	user := &UpdateModelWithSkipTag{
		ID:        123,
		Name:      "John Updated",
		Email:     "john.updated@example.com",
		CreatedAt: "2024-01-01", // Should be skipped due to dbUpdate:"false" tag
	}

	mock.ExpectExec(`UPDATE "users" SET "name" = \$1, "email" = \$2 WHERE "id" = \$3`).
		WithArgs("John Updated", "john.updated@example.com", int64(123)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Update(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestUpdate_DbUpdateFalseTagValue_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	user := &UpdateModelWithSkipTagValue{
		ID:        123,
		Name:      "John Updated",
		Email:     "john.updated@example.com",
		UpdatedAt: "2024-01-01", // Should be skipped due to dbUpdate:"false" tag
	}

	mock.ExpectExec(`UPDATE "users" SET "name" = \$1, "email" = \$2 WHERE "id" = \$3`).
		WithArgs("John Updated", "john.updated@example.com", int64(123)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Update(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestUpdate_SkipTagDashWithNonZeroValue_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	// Even though CreatedAt has a non-zero value, it should be skipped
	user := &UpdateModelWithSkipTag{
		ID:        123,
		Name:      "John Updated",
		CreatedAt: "2024-01-01 10:00:00", // Non-zero but should be skipped
	}

	mock.ExpectExec(`UPDATE "users" SET "name" = \$1 WHERE "id" = \$2`).
		WithArgs("John Updated", int64(123)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Update(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestUpdate_DbInsertFalse_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	// UpdatedAt should be included in UPDATE even though it has dbInsert:"false"
	user := &UpdateModelWithDbInsertFalse{
		ID:        123,
		Name:      "John Updated",
		Email:     "john.updated@example.com",
		UpdatedAt: "2024-01-01", // Should be included in Update (dbInsert:"false" doesn't affect Update)
	}

	mock.ExpectExec(`UPDATE "users" SET "name" = \$1, "email" = \$2, "updated_at" = \$3 WHERE "id" = \$4`).
		WithArgs("John Updated", "john.updated@example.com", "2024-01-01", int64(123)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Update(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestUpdate_DbUpdateFalse_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	// CreatedAt should be excluded from UPDATE due to dbUpdate:"false"
	user := &UpdateModelWithSkipTag{
		ID:        123,
		Name:      "John Updated",
		Email:     "john.updated@example.com",
		CreatedAt: "2024-01-01", // Should be skipped due to dbUpdate:"false" tag
	}

	mock.ExpectExec(`UPDATE "users" SET "name" = \$1, "email" = \$2 WHERE "id" = \$3`).
		WithArgs("John Updated", "john.updated@example.com", int64(123)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Update(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

func TestUpdate_BothTagsFalse_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	// CreatedAt should be excluded from UPDATE due to dbUpdate:"false" (dbInsert:"false" doesn't affect Update)
	user := &UpdateModelWithBothTags{
		ID:        123,
		Name:      "John Updated",
		Email:     "john.updated@example.com",
		CreatedAt: "2024-01-01", // Should be skipped due to dbUpdate:"false" tag
	}

	mock.ExpectExec(`UPDATE "users" SET "name" = \$1, "email" = \$2 WHERE "id" = \$3`).
		WithArgs("John Updated", "john.updated@example.com", int64(123)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Update(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

// TestUpdate_AutoTimestamp_PostgreSQL tests auto-updated timestamp with PostgreSQL
func TestUpdate_AutoTimestamp_PostgreSQL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	// UpdatedAt should be auto-populated with CURRENT_TIMESTAMP, not included in args
	user := &UpdateModelWithAutoTimestamp{
		ID:    123,
		Name:  "John Updated",
		Email: "john.updated@example.com",
		// UpdatedAt is not set - should be auto-populated
	}

	mock.ExpectExec(`UPDATE "users" SET "name" = \$1, "email" = \$2, "updated_at" = CURRENT_TIMESTAMP WHERE "id" = \$3`).
		WithArgs("John Updated", "john.updated@example.com", int64(123)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Update(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

// TestUpdate_AutoTimestamp_MySQL tests auto-updated timestamp with MySQL
func TestUpdate_AutoTimestamp_MySQL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "mysql", 5*time.Second)
	ctx := context.Background()

	user := &UpdateModelWithAutoTimestamp{
		ID:   123,
		Name: "John Updated",
		// UpdatedAt is not set - should be auto-populated
	}

	mock.ExpectExec("UPDATE `users` SET `name` = \\?, `updated_at` = NOW\\(\\) WHERE `id` = \\?").
		WithArgs("John Updated", int64(123)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Update(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

// TestUpdate_AutoTimestamp_SQLite tests auto-updated timestamp with SQLite
func TestUpdate_AutoTimestamp_SQLite(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "sqlite3", 5*time.Second)
	ctx := context.Background()

	user := &UpdateModelWithAutoTimestamp{
		ID:    123,
		Email: "updated@example.com",
		// UpdatedAt is not set - should be auto-populated
	}

	mock.ExpectExec(`UPDATE "users" SET "email" = \?, "updated_at" = CURRENT_TIMESTAMP WHERE "id" = \?`).
		WithArgs("updated@example.com", int64(123)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Update(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

// TestUpdate_AutoTimestamp_SQLServer tests auto-updated timestamp with SQL Server
func TestUpdate_AutoTimestamp_SQLServer(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "sqlserver", 5*time.Second)
	ctx := context.Background()

	user := &UpdateModelWithAutoTimestamp{
		ID:    123,
		Name:  "John Updated",
		Email: "john.updated@example.com",
		// UpdatedAt is not set - should be auto-populated
	}

	mock.ExpectExec(`UPDATE \[users\] SET \[name\] = @p1, \[email\] = @p2, \[updated_at\] = GETDATE\(\) WHERE \[id\] = @p3`).
		WithArgs("John Updated", "john.updated@example.com", int64(123)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Update(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

// TestUpdate_AutoTimestamp_Oracle tests auto-updated timestamp with Oracle
func TestUpdate_AutoTimestamp_Oracle(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "oracle", 5*time.Second)
	ctx := context.Background()

	user := &UpdateModelWithAutoTimestamp{
		ID:   123,
		Name: "John Updated",
		// UpdatedAt is not set - should be auto-populated
	}

	mock.ExpectExec(`UPDATE "USERS" SET "NAME" = :1, "UPDATED_AT" = CURRENT_TIMESTAMP WHERE "ID" = :2`).
		WithArgs("John Updated", int64(123)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Update(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

// TestUpdate_AutoTimestamp_OnlyAutoField tests Update with only auto-timestamp field (no regular fields)
func TestUpdate_AutoTimestamp_OnlyAutoField(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	// Only UpdatedAt field with dbUpdate:"auto-timestamp", no other fields set
	user := &UpdateModelWithAutoTimestamp{
		ID: 123,
		// Name and Email are zero values, UpdatedAt is auto
	}

	mock.ExpectExec(`UPDATE "users" SET "updated_at" = CURRENT_TIMESTAMP WHERE "id" = \$1`).
		WithArgs(int64(123)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Update(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}

// TestUpdate_AutoTimestamp_WithRegularFields tests Update with both auto-timestamp and regular fields
func TestUpdate_AutoTimestamp_WithRegularFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	user := &UpdateModelWithAutoTimestamp{
		ID:    123,
		Name:  "John Updated",
		Email: "john.updated@example.com",
		// UpdatedAt is auto-populated
	}

	mock.ExpectExec(`UPDATE "users" SET "name" = \$1, "email" = \$2, "updated_at" = CURRENT_TIMESTAMP WHERE "id" = \$3`).
		WithArgs("John Updated", "john.updated@example.com", int64(123)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = Update(ctx, typedbDB, user)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet mock expectations: %v", err)
	}
}
