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

func (u *TestUser) Deserialize(row map[string]any) error {
	// Stub implementation for testing registry
	return nil
}

type TestPost struct {
	Model
	ID     int `db:"id" load:"primary"`
	UserID int `db:"user_id"`
}

func (p *TestPost) Deserialize(row map[string]any) error {
	// Stub implementation for testing registry
	return nil
}

type TestModel struct {
	Model
	ID int `db:"id" load:"primary"`
}

func (m *TestModel) Deserialize(row map[string]any) error {
	// Stub implementation for testing registry
	return nil
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
