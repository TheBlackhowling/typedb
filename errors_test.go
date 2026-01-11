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
