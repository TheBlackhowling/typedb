package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/TheBlackHowling/typedb"
)

// ValidRegistrationModel is a valid model for testing registration
type ValidRegistrationModel struct {
	typedb.Model
	ID   int    `db:"id" load:"primary"`
	Name string `db:"name"`
}

func (m *ValidRegistrationModel) TableName() string {
	return "test_registration"
}

func (m *ValidRegistrationModel) QueryByID() string {
	return "SELECT id, name FROM test_registration WHERE id = ?"
}

// InvalidRegistrationModelMissingQueryBy is an invalid model missing QueryByID method
type InvalidRegistrationModelMissingQueryBy struct {
	typedb.Model
	ID   int    `db:"id" load:"primary"`
	Name string `db:"name"`
}

func (m *InvalidRegistrationModelMissingQueryBy) TableName() string {
	return "test_registration"
}

// InvalidRegistrationModelMissingUniqueQueryBy is an invalid model missing QueryByEmail method
type InvalidRegistrationModelMissingUniqueQueryBy struct {
	typedb.Model
	ID    int    `db:"id" load:"primary"`
	Email string `db:"email" load:"unique"`
}

func (m *InvalidRegistrationModelMissingUniqueQueryBy) TableName() string {
	return "test_registration"
}

func (m *InvalidRegistrationModelMissingUniqueQueryBy) QueryByID() string {
	return "SELECT id, email FROM test_registration WHERE id = ?"
}

// InvalidRegistrationModelMissingCompositeQueryBy is an invalid model missing composite QueryBy method
type InvalidRegistrationModelMissingCompositeQueryBy struct {
	typedb.Model
	UserID int `db:"user_id" load:"composite:testcomposite"`
	PostID int `db:"post_id" load:"composite:testcomposite"`
}

func (m *InvalidRegistrationModelMissingCompositeQueryBy) TableName() string {
	return "test_registration"
}

func TestSQLite_RegistrationValidation_ValidModel(t *testing.T) {
	// Reset registry for isolated test
	typedb.ResetValidation()

	// This should not panic - valid model with QueryByID method
	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()

		typedb.RegisterModel[*ValidRegistrationModel]()
	}()

	if panicked {
		t.Error("RegisterModel should not panic for valid model with QueryByID method")
	}
}

func TestSQLite_RegistrationValidation_MissingPrimaryQueryBy(t *testing.T) {
	// Reset registry for isolated test
	typedb.ResetValidation()

	// This should panic - missing QueryByID method
	panicked := false
	var panicMsg string
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
				panicMsg = fmt.Sprintf("%v", r)
			}
		}()

		typedb.RegisterModel[*InvalidRegistrationModelMissingQueryBy]()
	}()

	if !panicked {
		t.Error("RegisterModel should panic for model missing QueryByID method")
	}

	if !strings.Contains(panicMsg, "validation failed") {
		t.Errorf("Expected panic message to contain 'validation failed', got: %s", panicMsg)
	}

	if !strings.Contains(panicMsg, "QueryByID") {
		t.Errorf("Expected panic message to mention QueryByID, got: %s", panicMsg)
	}
}

func TestSQLite_RegistrationValidation_MissingUniqueQueryBy(t *testing.T) {
	// Reset registry for isolated test
	typedb.ResetValidation()

	// This should panic - missing QueryByEmail method for unique field
	panicked := false
	var panicMsg string
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
				panicMsg = fmt.Sprintf("%v", r)
			}
		}()

		typedb.RegisterModel[*InvalidRegistrationModelMissingUniqueQueryBy]()
	}()

	if !panicked {
		t.Error("RegisterModel should panic for model missing QueryByEmail method for unique field")
	}

	if !strings.Contains(panicMsg, "validation failed") {
		t.Errorf("Expected panic message to contain 'validation failed', got: %s", panicMsg)
	}

	if !strings.Contains(panicMsg, "QueryByEmail") {
		t.Errorf("Expected panic message to mention QueryByEmail, got: %s", panicMsg)
	}
}

func TestSQLite_RegistrationValidation_MissingCompositeQueryBy(t *testing.T) {
	// Reset registry for isolated test
	typedb.ResetValidation()

	// This should panic - missing QueryByPostIDUserID method for composite key
	panicked := false
	var panicMsg string
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
				panicMsg = fmt.Sprintf("%v", r)
			}
		}()

		typedb.RegisterModel[*InvalidRegistrationModelMissingCompositeQueryBy]()
	}()

	if !panicked {
		t.Error("RegisterModel should panic for model missing QueryByPostIDUserID method for composite key")
	}

	if !strings.Contains(panicMsg, "validation failed") {
		t.Errorf("Expected panic message to contain 'validation failed', got: %s", panicMsg)
	}

	if !strings.Contains(panicMsg, "QueryByPostIDUserID") {
		t.Errorf("Expected panic message to mention QueryByPostIDUserID, got: %s", panicMsg)
	}
}

func TestSQLite_RegistrationValidation_RegisterModelWithOptions_ValidModel(t *testing.T) {
	// Reset registry for isolated test
	typedb.ResetValidation()

	// This should not panic - valid model with QueryByID method
	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()

		typedb.RegisterModelWithOptions[*ValidRegistrationModel](typedb.ModelOptions{PartialUpdate: true})
	}()

	if panicked {
		t.Error("RegisterModelWithOptions should not panic for valid model with QueryByID method")
	}
}

func TestSQLite_RegistrationValidation_RegisterModelWithOptions_InvalidModel(t *testing.T) {
	// Reset registry for isolated test
	typedb.ResetValidation()

	// This should panic - missing QueryByID method
	panicked := false
	var panicMsg string
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
				panicMsg = fmt.Sprintf("%v", r)
			}
		}()

		typedb.RegisterModelWithOptions[*InvalidRegistrationModelMissingQueryBy](typedb.ModelOptions{PartialUpdate: true})
	}()

	if !panicked {
		t.Error("RegisterModelWithOptions should panic for model missing QueryByID method")
	}

	if !strings.Contains(panicMsg, "validation failed") {
		t.Errorf("Expected panic message to contain 'validation failed', got: %s", panicMsg)
	}
}

func TestSQLite_RegistrationValidation_ValidModelCanBeUsed(t *testing.T) {
	db := setupTestDB(t)
	defer closeDB(t, db)
	defer os.Remove(getTestDSN())

	ctx := context.Background()

	// Verify that a valid registered model can be used
	// Note: This test verifies that registration-time validation doesn't break normal usage
	user := &User{ID: 1}
	if err := typedb.Load(ctx, db, user); err != nil {
		t.Fatalf("Load failed for valid registered model: %v", err)
	}

	if user.Name == "" {
		t.Error("User name should be loaded")
	}
}
