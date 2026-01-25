package typedb

import (
	"errors"
	"testing"
)

func TestErrNotFound(t *testing.T) {
	if ErrNotFound == nil {
		t.Fatal("ErrNotFound should not be nil")
	}

	if ErrNotFound.Error() != "typedb: record not found" {
		t.Errorf("ErrNotFound.Error() = %q, want %q", ErrNotFound.Error(), "typedb: record not found")
	}

	// Test that it's a distinct error
	if errors.Is(ErrNotFound, errors.New("different error")) {
		t.Error("ErrNotFound should not match different errors")
	}
}

func TestErrFieldNotFound(t *testing.T) {
	if ErrFieldNotFound == nil {
		t.Fatal("ErrFieldNotFound should not be nil")
	}

	if ErrFieldNotFound.Error() != "typedb: field not found" {
		t.Errorf("ErrFieldNotFound.Error() = %q, want %q", ErrFieldNotFound.Error(), "typedb: field not found")
	}

	// Test that it can be checked with errors.Is
	if !errors.Is(ErrFieldNotFound, ErrFieldNotFound) {
		t.Error("errors.Is should return true for ErrFieldNotFound")
	}

	// Test that it's distinct from other errors
	if errors.Is(ErrFieldNotFound, ErrNotFound) {
		t.Error("ErrFieldNotFound should not be equal to ErrNotFound")
	}
}

func TestErrMethodNotFound(t *testing.T) {
	if ErrMethodNotFound == nil {
		t.Fatal("ErrMethodNotFound should not be nil")
	}

	if ErrMethodNotFound.Error() != "typedb: method not found" {
		t.Errorf("ErrMethodNotFound.Error() = %q, want %q", ErrMethodNotFound.Error(), "typedb: method not found")
	}

	// Test that it can be checked with errors.Is
	if !errors.Is(ErrMethodNotFound, ErrMethodNotFound) {
		t.Error("errors.Is should return true for ErrMethodNotFound")
	}

	// Test that it's distinct from other errors
	if errors.Is(ErrMethodNotFound, ErrNotFound) {
		t.Error("ErrMethodNotFound should not be equal to ErrNotFound")
	}
	if errors.Is(ErrMethodNotFound, ErrFieldNotFound) {
		t.Error("ErrMethodNotFound should not be equal to ErrFieldNotFound")
	}
}
