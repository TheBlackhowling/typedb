package typedb

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestInsert_PostgreSQL_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
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

func TestInsert_MySQL_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "mysql", 5*time.Second)
	ctx := context.Background()

	user := &InsertModel{Name: "John", Email: "john@example.com"}

	mock.ExpectExec("INSERT INTO `users`").
		WithArgs("John", "john@example.com").
		WillReturnResult(sqlmock.NewResult(123, 1))

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

func TestInsert_SkipsNilValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	user := &InsertModel{Name: "John"} // Email is empty string, should be skipped

	rows := sqlmock.NewRows([]string{"id"}).AddRow(123)

	mock.ExpectQuery(`INSERT INTO "users" \("name"\) VALUES \(\$1\) RETURNING "id"`).
		WithArgs("John").
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

func TestInsert_NoTableName_Error(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	model := &NoTableNameModel{ID: 1}

	err = Insert(ctx, typedbDB, model)
	if err == nil {
		t.Fatal("Expected error for missing TableName() method")
	}

	if !strings.Contains(err.Error(), "TableName") {
		t.Errorf("Expected error about TableName, got: %v", err)
	}
}

func TestInsert_NoPrimaryKey_Error(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	model := &NoPrimaryKeyModel{ID: 1}

	err = Insert(ctx, typedbDB, model)
	if err == nil {
		t.Fatal("Expected error for missing load:\"primary\" tag")
	}

	if !strings.Contains(err.Error(), "load:\"primary\"") {
		t.Errorf("Expected error about primary key, got: %v", err)
	}
}

func TestInsert_JoinedModel_Error(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	typedbDB := NewDB(db, "postgres", 5*time.Second)
	ctx := context.Background()

	model := &JoinedModel{UserID: 1, Bio: "test"}

	err = Insert(ctx, typedbDB, model)
	if err == nil {
		t.Fatal("Expected error for joined model")
	}

	if !strings.Contains(err.Error(), "joined") && !strings.Contains(err.Error(), "dot notation") {
		t.Errorf("Expected error about joined model, got: %v", err)
	}
}
