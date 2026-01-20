package typedb

import (
	"fmt"
	"strings"
	"testing"
)

// Valid test models

type ValidUser struct {
	Model
	ID    int    `db:"id" load:"primary"`
	Email string `db:"email" load:"unique"`
	Name  string `db:"name"`
}

func (u *ValidUser) QueryByID() string {
	return "SELECT id, email, name FROM users WHERE id = $1"
}

func (u *ValidUser) QueryByEmail() string {
	return "SELECT id, email, name FROM users WHERE email = $1"
}

func (u *ValidUser) Deserialize(row map[string]any) error {
	return nil
}

type ValidUserPost struct {
	Model
	UserID int `db:"user_id" load:"composite:userpost"`
	PostID int `db:"post_id" load:"composite:userpost"`
}

func (up *ValidUserPost) QueryByPostIDUserID() string {
	return "SELECT user_id, post_id FROM user_posts WHERE post_id = $1 AND user_id = $2"
}

func (up *ValidUserPost) Deserialize(row map[string]any) error {
	return nil
}

type ValidUserRole struct {
	Model
	UserID int `db:"user_id" load:"primary,composite:userrole"`
	RoleID int `db:"role_id" load:"composite:userrole"`
}

func (ur *ValidUserRole) QueryByUserID() string {
	return "SELECT user_id, role_id FROM user_roles WHERE user_id = $1"
}

func (ur *ValidUserRole) QueryByRoleIDUserID() string {
	return "SELECT user_id, role_id FROM user_roles WHERE role_id = $1 AND user_id = $2"
}

func (ur *ValidUserRole) Deserialize(row map[string]any) error {
	return nil
}

// Invalid test models

type MissingPrimaryMethod struct {
	Model
	ID int `db:"id" load:"primary"`
}

func (m *MissingPrimaryMethod) Deserialize(row map[string]any) error {
	return nil
}

type MultiplePrimaryKeys struct {
	Model
	ID   int `db:"id" load:"primary"`
	UUID int `db:"uuid" load:"primary"`
}

func (m *MultiplePrimaryKeys) Deserialize(row map[string]any) error {
	return nil
}

type InvalidCompositeKey struct {
	Model
	UserID int `db:"user_id" load:"composite:userpost"`
	// Only one field in composite - should fail
}

func (i *InvalidCompositeKey) Deserialize(row map[string]any) error {
	return nil
}

type MissingCompositeMethod struct {
	Model
	UserID int `db:"user_id" load:"composite:userpost"`
	PostID int `db:"post_id" load:"composite:userpost"`
}

func (m *MissingCompositeMethod) Deserialize(row map[string]any) error {
	return nil
}

type MissingUniqueMethod struct {
	Model
	ID    int    `db:"id" load:"primary"`
	Email string `db:"email" load:"unique"`
}

func (m *MissingUniqueMethod) QueryByID() string {
	return "SELECT id, email FROM users WHERE id = $1"
}

func (m *MissingUniqueMethod) Deserialize(row map[string]any) error {
	return nil
}

type WrongMethodSignature struct {
	Model
	ID int `db:"id" load:"primary"`
}

func (w *WrongMethodSignature) QueryByID() int {
	return 123
}

func (w *WrongMethodSignature) Deserialize(row map[string]any) error {
	return nil
}

type WrongMethodParams struct {
	Model
	ID int `db:"id" load:"primary"`
}

func (w *WrongMethodParams) QueryByID(id int) string {
	return "SELECT * FROM users WHERE id = $1"
}

func (w *WrongMethodParams) Deserialize(row map[string]any) error {
	return nil
}

type WrongMethodReturnCount struct {
	Model
	ID int `db:"id" load:"primary"`
}

func (w *WrongMethodReturnCount) QueryByID() (string, error) {
	return "SELECT * FROM users WHERE id = $1", nil
}

func (w *WrongMethodReturnCount) Deserialize(row map[string]any) error {
	return nil
}

func TestValidateModel_ValidPrimary(t *testing.T) {
	user := &ValidUser{ID: 1}
	err := ValidateModel(user)
	if err != nil {
		t.Errorf("Expected no error for valid primary key model, got: %v", err)
	}
}

func TestValidateModel_ValidUnique(t *testing.T) {
	user := &ValidUser{ID: 1, Email: "test@example.com"}
	err := ValidateModel(user)
	if err != nil {
		t.Errorf("Expected no error for valid unique field model, got: %v", err)
	}
}

func TestValidateModel_ValidComposite(t *testing.T) {
	userPost := &ValidUserPost{UserID: 1, PostID: 2}
	err := ValidateModel(userPost)
	if err != nil {
		t.Errorf("Expected no error for valid composite key model, got: %v", err)
	}
}

func TestValidateModel_ValidPrimaryAndComposite(t *testing.T) {
	userRole := &ValidUserRole{UserID: 1, RoleID: 2}
	err := ValidateModel(userRole)
	if err != nil {
		t.Errorf("Expected no error for valid primary and composite key model, got: %v", err)
	}
}

func TestValidateModel_MissingPrimaryMethod(t *testing.T) {
	model := &MissingPrimaryMethod{ID: 1}
	err := ValidateModel(model)
	if err == nil {
		t.Error("Expected error for missing QueryByID method")
	}
	if !strings.Contains(err.Error(), "QueryByID") {
		t.Errorf("Expected error to mention QueryByID, got: %v", err)
	}
}

func TestValidateModel_MultiplePrimaryKeys(t *testing.T) {
	model := &MultiplePrimaryKeys{ID: 1, UUID: 2}
	err := ValidateModel(model)
	if err == nil {
		t.Error("Expected error for multiple primary keys")
	}
	if !strings.Contains(err.Error(), "multiple primary key fields") {
		t.Errorf("Expected error about multiple primary keys, got: %v", err)
	}
}

func TestValidateModel_InvalidCompositeKey(t *testing.T) {
	model := &InvalidCompositeKey{UserID: 1}
	err := ValidateModel(model)
	if err == nil {
		t.Error("Expected error for composite key with only one field")
	}
	if !strings.Contains(err.Error(), "at least 2 required") {
		t.Errorf("Expected error about needing at least 2 fields, got: %v", err)
	}
}

func TestValidateModel_MissingCompositeMethod(t *testing.T) {
	model := &MissingCompositeMethod{UserID: 1, PostID: 2}
	err := ValidateModel(model)
	if err == nil {
		t.Error("Expected error for missing composite query method")
	}
	if !strings.Contains(err.Error(), "QueryByPostIDUserID") {
		t.Errorf("Expected error to mention QueryByPostIDUserID, got: %v", err)
	}
}

func TestValidateModel_WrongMethodSignature(t *testing.T) {
	model := &WrongMethodSignature{ID: 1}
	err := ValidateModel(model)
	if err == nil {
		t.Error("Expected error for wrong method return type")
	}
	if !strings.Contains(err.Error(), "should return string") {
		t.Errorf("Expected error about return type, got: %v", err)
	}
}

func TestValidateModel_WrongMethodParams(t *testing.T) {
	model := &WrongMethodParams{ID: 1}
	err := ValidateModel(model)
	if err == nil {
		t.Error("Expected error for method with parameters")
	}
	if !strings.Contains(err.Error(), "should have no parameters") {
		t.Errorf("Expected error about parameters, got: %v", err)
	}
}

func TestValidateModel_WrongMethodReturnCount(t *testing.T) {
	model := &WrongMethodReturnCount{ID: 1}
	err := ValidateModel(model)
	if err == nil {
		t.Error("Expected error for method returning multiple values")
	}
	if !strings.Contains(err.Error(), "should return exactly one value") {
		t.Errorf("Expected error about return count, got: %v", err)
	}
}

func TestValidateModel_MissingUniqueMethod(t *testing.T) {
	model := &MissingUniqueMethod{ID: 1, Email: "test@example.com"}
	err := ValidateModel(model)
	if err == nil {
		t.Error("Expected error for missing QueryByEmail method")
	}
	if !strings.Contains(err.Error(), "QueryByEmail") {
		t.Errorf("Expected error to mention QueryByEmail, got: %v", err)
	}
}

func TestValidateAllRegistered_NoModels(t *testing.T) {
	// Reset registry
	registeredModels = nil

	err := ValidateAllRegistered()
	if err != nil {
		t.Errorf("Expected no error when no models registered, got: %v", err)
	}
}

func TestValidateAllRegistered_ValidModels(t *testing.T) {
	// Reset registry
	registeredModels = nil

	RegisterModel[*ValidUser]()
	RegisterModel[*ValidUserPost]()

	err := ValidateAllRegistered()
	if err != nil {
		t.Errorf("Expected no error for valid models, got: %v", err)
	}
}

func TestValidateAllRegistered_InvalidModels(t *testing.T) {
	// Reset registry
	registeredModels = nil

	RegisterModel[*ValidUser]()
	
	// Registration-time validation should panic for invalid models
	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
				// Verify the panic message contains validation error
				panicMsg := fmt.Sprintf("%v", r)
				if !strings.Contains(panicMsg, "validation failed") {
					t.Errorf("Expected panic message to contain 'validation failed', got: %s", panicMsg)
				}
				if !strings.Contains(panicMsg, "MissingPrimaryMethod") {
					t.Errorf("Expected panic message to contain 'MissingPrimaryMethod', got: %s", panicMsg)
				}
			}
		}()
		
		RegisterModel[*MissingPrimaryMethod]()
	}()
	
	if !panicked {
		t.Error("Expected panic when registering invalid model MissingPrimaryMethod")
	}
}

func TestMustValidateAllRegistered_NoError(t *testing.T) {
	// Reset registry
	registeredModels = nil

	RegisterModel[*ValidUser]()

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustValidateAllRegistered should not panic for valid models, panicked with: %v", r)
		}
	}()

	MustValidateAllRegistered()
}

func TestMustValidateAllRegistered_Panic(t *testing.T) {
	// Reset registry and validation cache
	registeredModels = nil
	ResetValidation()

	// Registration-time validation should panic for invalid models
	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		
		RegisterModel[*MissingPrimaryMethod]()
	}()
	
	if !panicked {
		t.Error("Expected panic when registering invalid model MissingPrimaryMethod")
	}
	
	// Since registration panics, MustValidateAllRegistered should not be called
	// But if it were called with valid models, it should not panic
	registeredModels = nil
	RegisterModel[*ValidUser]()
	
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustValidateAllRegistered should not panic for valid models, panicked with: %v", r)
		}
	}()
	
	MustValidateAllRegistered()
}

func TestValidationError_Error(t *testing.T) {
	ve := &ValidationError{
		ModelName: "TestModel",
		Errors:    []string{"error 1", "error 2"},
	}

	errMsg := ve.Error()
	if !strings.Contains(errMsg, "TestModel") {
		t.Errorf("Expected error message to contain model name, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "error 1") {
		t.Errorf("Expected error message to contain first error, got: %s", errMsg)
	}
}

func TestValidationErrors_Error(t *testing.T) {
	ve := &ValidationErrors{
		Errors: []*ValidationError{
			{ModelName: "Model1", Errors: []string{"error 1"}},
			{ModelName: "Model2", Errors: []string{"error 2"}},
		},
	}

	errMsg := ve.Error()
	if !strings.Contains(errMsg, "Model1") {
		t.Errorf("Expected error message to contain Model1, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "Model2") {
		t.Errorf("Expected error message to contain Model2, got: %s", errMsg)
	}
}
