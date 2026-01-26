package typedb

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// Test models for Insert function additional tests
// PrimaryKeyNoDbTagModel has primary key without db tag
type PrimaryKeyNoDbTagModel struct {
	Model
	ID   int    `load:"primary"` // Missing db tag
	Name string `db:"name"`
}

func (m *PrimaryKeyNoDbTagModel) TableName() string {
	return "users"
}

// PrimaryKeyDashTagModel has primary key with db:"-" tag
type PrimaryKeyDashTagModel struct {
	Model
	ID   int    `db:"-" load:"primary"` // db:"-" tag
	Name string `db:"name"`
}

func (m *PrimaryKeyDashTagModel) TableName() string {
	return "users"
}

// PrimaryKeyDotNotationModel has primary key with dot notation
type PrimaryKeyDotNotationModel struct {
	Model
	UserID int    `db:"users.id" load:"primary"` // Dot notation
	Name   string `db:"name"`
}

func (m *PrimaryKeyDotNotationModel) TableName() string {
	return "users"
}

// AllZeroFieldsModel has all fields zero/nil
type AllZeroFieldsModel struct {
	Model
	ID    int64  `db:"id" load:"primary"`
	Name  string `db:"name"`
	Email string `db:"email"`
}

func (m *AllZeroFieldsModel) TableName() string {
	return "users"
}

func TestInsert_PrimaryKeyNoDbTag_Error(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	model := &PrimaryKeyNoDbTagModel{ID: 1, Name: "John"}

	err = Insert(ctx, typedbDB, model)
	if err == nil {
		t.Fatal("Expected error for primary key without db tag")
	}

	if !strings.Contains(err.Error(), "db tag") {
		t.Errorf("Expected error about db tag, got: %v", err)
	}
}

func TestInsert_PrimaryKeyDashTag_Error(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	model := &PrimaryKeyDashTagModel{ID: 1, Name: "John"}

	err = Insert(ctx, typedbDB, model)
	if err == nil {
		t.Fatal("Expected error for primary key with db:\"-\" tag")
	}

	if !strings.Contains(err.Error(), "db tag") {
		t.Errorf("Expected error about db tag, got: %v", err)
	}
}

func TestInsert_AllZeroFields_Error(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	model := &AllZeroFieldsModel{} // All fields are zero

	err = Insert(ctx, typedbDB, model)
	if err == nil {
		t.Fatal("Expected error for all zero fields")
	}

	if !strings.Contains(err.Error(), "at least one non-nil field") {
		t.Errorf("Expected error about non-nil field, got: %v", err)
	}
}
