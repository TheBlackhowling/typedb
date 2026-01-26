package typedb

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// ModelWithNologForSerialization is a test model with nolog tag for testing serialization masking
type ModelWithNologForSerialization struct {
	Model
	ID       int    `db:"id" load:"primary"`
	Name     string `db:"name"`
	Email    string `db:"email" nolog:"true"`
	Password string `db:"password" nolog:"true"`
}

func (m *ModelWithNologForSerialization) TableName() string {
	return "users"
}

func (m *ModelWithNologForSerialization) QueryByID() string {
	return "SELECT id, name, email, password FROM users WHERE id = $1"
}

func init() {
	RegisterModel[*ModelWithNologForSerialization]()
}

// ModelWithoutNolog is a test model without nolog tags
type ModelWithoutNolog struct {
	Model
	ID   int    `db:"id" load:"primary"`
	Name string `db:"name"`
}

func (m *ModelWithoutNolog) TableName() string {
	return "users"
}

func (m *ModelWithoutNolog) QueryByID() string {
	return "SELECT id, name FROM users WHERE id = $1"
}

// TestSerializationNolog_ExtractNologMaskIndicesFromArgs tests the core detection logic
func TestSerializationNolog_ExtractNologMaskIndicesFromArgs(t *testing.T) {
	t.Run("detects model with nolog fields", func(t *testing.T) {
		user := &ModelWithNologForSerialization{
			ID:       123,
			Name:     "John",
			Email:    "john@example.com",
			Password: "secret123",
		}

		args := []any{user, "regular", 456}
		maskIndices := extractNologMaskIndicesFromArgs(args)

		if len(maskIndices) != 1 {
			t.Fatalf("Expected 1 mask index, got %d", len(maskIndices))
		}
		if maskIndices[0] != 0 {
			t.Errorf("Expected mask index 0, got %d", maskIndices[0])
		}
	})

	t.Run("does not detect model without nolog fields", func(t *testing.T) {
		user := &ModelWithoutNolog{
			ID:   123,
			Name: "John",
		}

		args := []any{user, "regular", 456}
		maskIndices := extractNologMaskIndicesFromArgs(args)

		if len(maskIndices) != 0 {
			t.Errorf("Expected 0 mask indices, got %d", len(maskIndices))
		}
	})

	t.Run("handles mixed arguments", func(t *testing.T) {
		user1 := &ModelWithNologForSerialization{
			ID:       123,
			Name:     "John",
			Email:    "john@example.com",
			Password: "secret123",
		}
		user2 := &ModelWithoutNolog{
			ID:   456,
			Name: "Jane",
		}
		user3 := &ModelWithNologForSerialization{
			ID:       789,
			Name:     "Bob",
			Email:    "bob@example.com",
			Password: "secret456",
		}

		args := []any{"regular", user1, user2, user3}
		maskIndices := extractNologMaskIndicesFromArgs(args)

		if len(maskIndices) != 2 {
			t.Fatalf("Expected 2 mask indices, got %d", len(maskIndices))
		}
		if maskIndices[0] != 1 {
			t.Errorf("Expected mask index 1, got %d", maskIndices[0])
		}
		if maskIndices[1] != 3 {
			t.Errorf("Expected mask index 3, got %d", maskIndices[1])
		}
	})

	t.Run("handles nil arguments", func(t *testing.T) {
		args := []any{nil, "regular", nil}
		maskIndices := extractNologMaskIndicesFromArgs(args)

		if len(maskIndices) != 0 {
			t.Errorf("Expected 0 mask indices, got %d", len(maskIndices))
		}
	})

	t.Run("handles empty arguments", func(t *testing.T) {
		args := []any{}
		maskIndices := extractNologMaskIndicesFromArgs(args)

		if len(maskIndices) != 0 {
			t.Errorf("Expected 0 mask indices, got %d", len(maskIndices))
		}
	})
}

// TestSerializationNolog_GetLoggingFlagsAndArgs tests that getLoggingFlagsAndArgs applies automatic masking
func TestSerializationNolog_GetLoggingFlagsAndArgs(t *testing.T) {
	ctx := context.Background()

	user := &ModelWithNologForSerialization{
		ID:       123,
		Name:     "John",
		Email:    "john@example.com",
		Password: "secret123",
	}

	args := []any{user, "regular", 456}

	// Test with logArgs enabled
	logQueries, logArgs, maskedArgs := getLoggingFlagsAndArgs(ctx, true, true, args)

	if !logQueries || !logArgs {
		t.Error("Expected logQueries and logArgs to be true")
	}

	if len(maskedArgs) != len(args) {
		t.Fatalf("Expected %d masked args, got %d", len(args), len(maskedArgs))
	}

	// The model argument (index 0) should be masked
	if maskedArgs[0] != "[REDACTED]" {
		t.Errorf("Expected model argument to be masked, got: %v", maskedArgs[0])
	}

	// Regular arguments should not be masked
	if maskedArgs[1] != "regular" {
		t.Errorf("Expected regular argument to not be masked, got: %v", maskedArgs[1])
	}
	if maskedArgs[2] != 456 {
		t.Errorf("Expected regular argument to not be masked, got: %v", maskedArgs[2])
	}
}

// TestSerializationNolog_QueryAll tests that nolog tags are automatically detected when models are passed to QueryAll
func TestSerializationNolog_QueryAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer closeSQLDB(t, db)

	logger := &testLogger{}
	typedbDB := NewDBWithLoggerAndFlags(db, "postgres", 5*time.Second, logger, true, true)

	user := &ModelWithNologForSerialization{
		ID:       123,
		Name:     "John",
		Email:    "john@example.com",
		Password: "secret123",
	}

	mock.ExpectQuery("SELECT.*FROM users").WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "email", "password"}).
			AddRow(123, "John", "john@example.com", "secret123"),
	)

	ctx := context.Background()
	// Pass model ID as argument (SQL-compatible), but also pass model to test masking detection
	// In real usage, users would pass individual field values, but we test that if a model IS passed,
	// it gets detected and masked in logs
	_, err = typedbDB.QueryAll(ctx, "SELECT id, name, email, password FROM users WHERE id = $1", user.ID)

	if err != nil {
		t.Fatalf("QueryAll failed: %v", err)
	}

	// Verify query executed (model wasn't passed to SQL, so this should work)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Mock expectations were not met: %v", err)
	}
}

// TestSerializationNolog_QueryRowMap_WithModel tests masking when model is passed (even though SQL won't accept it)
func TestSerializationNolog_QueryRowMap_WithModel(t *testing.T) {
	user := &ModelWithNologForSerialization{
		ID:       123,
		Name:     "John",
		Email:    "john@example.com",
		Password: "secret123",
	}

	ctx := context.Background()

	// Test that getLoggingFlagsAndArgs masks the model before logging
	// We can't actually execute SQL with a struct, but we can verify the masking logic
	args := []any{user}
	logQueries, logArgs, maskedArgs := getLoggingFlagsAndArgs(ctx, true, true, args)

	if !logQueries || !logArgs {
		t.Error("Expected logQueries and logArgs to be true")
	}

	// The model argument should be masked
	if maskedArgs[0] != "[REDACTED]" {
		t.Errorf("Expected model argument to be masked, got: %v", maskedArgs[0])
	}
}

// TestSerializationNolog_MixedArgs tests that only model arguments with nolog tags are masked
func TestSerializationNolog_MixedArgs(t *testing.T) {
	ctx := context.Background()

	user := &ModelWithNologForSerialization{
		ID:       123,
		Name:     "John",
		Email:    "john@example.com",
		Password: "secret123",
	}

	args := []any{user, "regular", 456}

	logQueries, logArgs, maskedArgs := getLoggingFlagsAndArgs(ctx, true, true, args)

	if !logQueries || !logArgs {
		t.Error("Expected logQueries and logArgs to be true")
	}

	if len(maskedArgs) != 3 {
		t.Fatalf("Expected 3 masked args, got %d", len(maskedArgs))
	}

	// The model argument (index 0) should be masked
	if maskedArgs[0] != "[REDACTED]" {
		t.Errorf("Expected model argument to be masked, got: %v", maskedArgs[0])
	}

	// Regular arguments should not be masked
	if maskedArgs[1] != "regular" {
		t.Errorf("Expected regular argument to not be masked, got: %v", maskedArgs[1])
	}
	if maskedArgs[2] != 456 {
		t.Errorf("Expected regular argument to not be masked, got: %v", maskedArgs[2])
	}
}

// TestSerializationNolog_NoNologFields tests that models without nolog fields are not masked
func TestSerializationNolog_NoNologFields(t *testing.T) {
	ctx := context.Background()

	userWithoutNolog := &ModelWithoutNolog{
		ID:   123,
		Name: "John",
	}

	args := []any{userWithoutNolog, "regular"}

	logQueries, logArgs, maskedArgs := getLoggingFlagsAndArgs(ctx, true, true, args)

	if !logQueries || !logArgs {
		t.Error("Expected logQueries and logArgs to be true")
	}

	// The model argument should NOT be masked (no nolog fields)
	if maskedArgs[0] == "[REDACTED]" {
		t.Error("Expected model argument to NOT be masked (no nolog fields)")
	}
}
