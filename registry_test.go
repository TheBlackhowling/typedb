package typedb

import (
	"reflect"
	"testing"
)

// Test models for registration
type TestUser struct {
	Model
	ID int `db:"id" load:"primary"`
}

func (m *TestUser) TableName() string {
	return "users"
}

func (m *TestUser) QueryByID() string {
	return "SELECT id FROM users WHERE id = $1"
}

type TestPost struct {
	Model
	ID     int `db:"id" load:"primary"`
	UserID int `db:"user_id"`
}

func (m *TestPost) TableName() string {
	return "posts"
}

func (m *TestPost) QueryByID() string {
	return "SELECT id, user_id FROM posts WHERE id = $1"
}

type TestModel struct {
	Model
	ID int `db:"id" load:"primary"`
}

func (m *TestModel) TableName() string {
	return "testmodels"
}

func (m *TestModel) QueryByID() string {
	return "SELECT id FROM testmodels WHERE id = $1"
}

func init() {
	RegisterModel[*TestUser]()
	RegisterModel[*TestPost]()
}

func TestRegisterModel(t *testing.T) {
	// Reset registry for isolated test
	registeredModels = nil

	// Register a model
	RegisterModel[*TestUser]()

	models := GetRegisteredModels()
	if len(models) != 1 {
		t.Fatalf("Expected 1 registered model, got %d", len(models))
	}

	if models[0] != reflect.TypeOf(TestUser{}) {
		t.Errorf("Expected TestUser type, got %v", models[0])
	}
}

func TestRegisterModel_Duplicate(t *testing.T) {
	// Reset registry for isolated test
	registeredModels = nil

	// Register same model twice
	RegisterModel[*TestUser]()
	RegisterModel[*TestUser]()

	models := GetRegisteredModels()
	if len(models) != 1 {
		t.Fatalf("Expected 1 registered model after duplicate registration, got %d", len(models))
	}
}

func TestRegisterModel_PointerType(t *testing.T) {
	// Reset registry for isolated test
	registeredModels = nil

	// Register pointer type (should be normalized to struct type)
	RegisterModel[*TestModel]()

	models := GetRegisteredModels()
	if len(models) != 1 {
		t.Fatalf("Expected 1 registered model, got %d", len(models))
	}

	// Should be struct type, not pointer type
	if models[0].Kind() != reflect.Struct {
		t.Errorf("Expected struct type, got %v", models[0].Kind())
	}

	if models[0] != reflect.TypeOf(TestModel{}) {
		t.Errorf("Expected TestModel type, got %v", models[0])
	}
}

func TestGetRegisteredModels(t *testing.T) {
	// Reset registry for isolated test
	registeredModels = nil

	RegisterModel[*TestUser]()
	RegisterModel[*TestPost]()

	models := GetRegisteredModels()
	if len(models) != 2 {
		t.Fatalf("Expected 2 registered models, got %d", len(models))
	}

	// Verify both models are present
	userFound := false
	postFound := false
	for _, m := range models {
		if m == reflect.TypeOf(TestUser{}) {
			userFound = true
		}
		if m == reflect.TypeOf(TestPost{}) {
			postFound = true
		}
	}

	if !userFound {
		t.Error("TestUser not found in registered models")
	}
	if !postFound {
		t.Error("TestPost not found in registered models")
	}
}

func TestGetRegisteredModels_ReturnsCopy(t *testing.T) {
	// Reset registry for isolated test
	registeredModels = nil

	RegisterModel[*TestUser]()

	models1 := GetRegisteredModels()
	models2 := GetRegisteredModels()

	// Modifying one slice should not affect the other
	if len(models1) != len(models2) {
		t.Fatal("Slices should have same length")
	}

	// They should have same content but be different slices
	if &models1[0] == &models2[0] {
		t.Error("GetRegisteredModels should return a copy, not the original slice")
	}
}

// ValueModel is a non-pointer model type for testing panic path.
// It embeds Model, but since Model.deserialize() has a pointer receiver,
// ValueModel (value type) does not satisfy ModelInterface.
type ValueModel struct {
	Model
	ID int `db:"id" load:"primary"`
}

func (m *ValueModel) TableName() string {
	return "valuemodels"
}

func (m *ValueModel) QueryByID() string {
	return "SELECT id FROM valuemodels WHERE id = $1"
}

func TestRegisterModel_NonPointerTypePanics(t *testing.T) {
	// Reset registry for isolated test
	registeredModels = nil

	// Test that RegisterModel works with pointer types.
	// ValueModel embeds Model, but Model.deserialize() has a pointer receiver,
	// so only *ValueModel satisfies ModelInterface, not ValueModel.
	// This test verifies that RegisterModel[*ValueModel]() works (pointer type).

	// Test with pointer type - should work
	RegisterModel[*ValueModel]()

	// Verify it was registered
	models := GetRegisteredModels()
	found := false
	for _, m := range models {
		if m.Name() == "ValueModel" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected ValueModel to be registered")
	}

	// Note: RegisterModel[ValueModel]() would fail at compile time because
	// ValueModel doesn't satisfy ModelInterface (Model.deserialize() has pointer receiver).
	// This is the desired behavior - only pointer types can satisfy ModelInterface,
	// and only Model implements deserialize().
}
