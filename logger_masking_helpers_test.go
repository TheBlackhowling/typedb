package typedb

import (
	"context"
	"testing"
)

// TestMaskArgs verifies that maskArgs function correctly masks arguments at specified indices
func TestMaskArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []any
		maskIndices []int
		expected    []any
	}{
		{
			name:        "no indices to mask",
			args:        []any{"John", "john@example.com", "password123"},
			maskIndices: []int{},
			expected:    []any{"John", "john@example.com", "password123"},
		},
		{
			name:        "mask single index",
			args:        []any{"John", "john@example.com", "password123"},
			maskIndices: []int{2},
			expected:    []any{"John", "john@example.com", "[REDACTED]"},
		},
		{
			name:        "mask first index",
			args:        []any{"secret123", "john@example.com", "John"},
			maskIndices: []int{0},
			expected:    []any{"[REDACTED]", "john@example.com", "John"},
		},
		{
			name:        "mask multiple indices",
			args:        []any{"John", "secret123", "anotherSecret"},
			maskIndices: []int{1, 2},
			expected:    []any{"John", "[REDACTED]", "[REDACTED]"},
		},
		{
			name:        "mask all indices",
			args:        []any{"secret1", "secret2", "secret3"},
			maskIndices: []int{0, 1, 2},
			expected:    []any{"[REDACTED]", "[REDACTED]", "[REDACTED]"},
		},
		{
			name:        "mask out of bounds index (should not panic)",
			args:        []any{"John", "john@example.com"},
			maskIndices: []int{5},
			expected:    []any{"John", "john@example.com"},
		},
		{
			name:        "mask negative index (should not panic)",
			args:        []any{"John", "john@example.com"},
			maskIndices: []int{-1},
			expected:    []any{"John", "john@example.com"},
		},
		{
			name:        "empty args",
			args:        []any{},
			maskIndices: []int{0},
			expected:    []any{},
		},
		{
			name:        "mask middle index",
			args:        []any{"John", "secret123", "Doe", "john@example.com"},
			maskIndices: []int{1},
			expected:    []any{"John", "[REDACTED]", "Doe", "john@example.com"},
		},
		{
			name:        "duplicate indices",
			args:        []any{"John", "secret123"},
			maskIndices: []int{1, 1},
			expected:    []any{"John", "[REDACTED]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskArgs(tt.args, tt.maskIndices)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected result length %d, got %d", len(tt.expected), len(result))
				return
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("Index %d: expected %v, got %v", i, tt.expected[i], result[i])
				}
			}

			// Verify original args are not modified
			if len(tt.args) > 0 && len(tt.maskIndices) > 0 {
				// Check that original args still contain original values
				for i, idx := range tt.maskIndices {
					if idx >= 0 && idx < len(tt.args) {
						// Original should still have the original value
						if result[idx] == "[REDACTED]" && tt.args[idx] == "[REDACTED]" {
							t.Errorf("Original args were modified at index %d", idx)
						}
					}
					_ = i // avoid unused variable
				}
			}
		})
	}
}

// TestWithMaskIndices verifies that WithMaskIndices stores indices in context correctly
func TestWithMaskIndices(t *testing.T) {
	ctx := context.Background()

	t.Run("store single index", func(t *testing.T) {
		ctx := WithMaskIndices(ctx, []int{1})
		maskIndices := ctx.Value(maskIndicesKey{})
		if maskIndices == nil {
			t.Fatal("Expected maskIndices to be stored in context")
		}
		indices, ok := maskIndices.([]int)
		if !ok {
			t.Fatalf("Expected []int, got %T", maskIndices)
		}
		if len(indices) != 1 || indices[0] != 1 {
			t.Errorf("Expected [1], got %v", indices)
		}
	})

	t.Run("store multiple indices", func(t *testing.T) {
		ctx := WithMaskIndices(ctx, []int{0, 2, 5})
		maskIndices := ctx.Value(maskIndicesKey{})
		if maskIndices == nil {
			t.Fatal("Expected maskIndices to be stored in context")
		}
		indices, ok := maskIndices.([]int)
		if !ok {
			t.Fatalf("Expected []int, got %T", maskIndices)
		}
		expected := []int{0, 2, 5}
		if len(indices) != len(expected) {
			t.Errorf("Expected length %d, got %d", len(expected), len(indices))
		}
		for i := range indices {
			if indices[i] != expected[i] {
				t.Errorf("Index %d: expected %d, got %d", i, expected[i], indices[i])
			}
		}
	})

	t.Run("store empty indices", func(t *testing.T) {
		ctx := WithMaskIndices(ctx, []int{})
		maskIndices := ctx.Value(maskIndicesKey{})
		if maskIndices == nil {
			t.Fatal("Expected maskIndices to be stored in context (even if empty)")
		}
		indices, ok := maskIndices.([]int)
		if !ok {
			t.Fatalf("Expected []int, got %T", maskIndices)
		}
		if len(indices) != 0 {
			t.Errorf("Expected empty slice, got %v", indices)
		}
	})

	t.Run("chaining WithMaskIndices overwrites previous", func(t *testing.T) {
		ctx := WithMaskIndices(ctx, []int{1, 2})
		ctx = WithMaskIndices(ctx, []int{3, 4})
		maskIndices := ctx.Value(maskIndicesKey{})
		if maskIndices == nil {
			t.Fatal("Expected maskIndices to be stored in context")
		}
		indices, ok := maskIndices.([]int)
		if !ok {
			t.Fatalf("Expected []int, got %T", maskIndices)
		}
		expected := []int{3, 4}
		if len(indices) != len(expected) {
			t.Errorf("Expected length %d, got %d", len(expected), len(indices))
		}
		for i := range indices {
			if indices[i] != expected[i] {
				t.Errorf("Index %d: expected %d, got %d", i, expected[i], indices[i])
			}
		}
	})
}

// TestGetLoggingFlagsAndArgs verifies that getLoggingFlagsAndArgs correctly retrieves and applies masking
func TestGetLoggingFlagsAndArgs(t *testing.T) {
	ctx := context.Background()

	t.Run("no masking when no mask indices in context", func(t *testing.T) {
		args := []any{"John", "john@example.com", "password123"}
		logQueries, logArgs, logArgsCopy := getLoggingFlagsAndArgs(ctx, true, true, args)

		if !logQueries {
			t.Error("Expected logQueries to be true")
		}
		if !logArgs {
			t.Error("Expected logArgs to be true")
		}
		if len(logArgsCopy) != len(args) {
			t.Errorf("Expected args length %d, got %d", len(args), len(logArgsCopy))
		}
		for i := range args {
			if logArgsCopy[i] != args[i] {
				t.Errorf("Index %d: expected %v, got %v", i, args[i], logArgsCopy[i])
			}
		}
	})

	t.Run("masking applied when mask indices in context", func(t *testing.T) {
		ctx := WithMaskIndices(ctx, []int{2})
		args := []any{"John", "john@example.com", "password123"}
		logQueries, logArgs, logArgsCopy := getLoggingFlagsAndArgs(ctx, true, true, args)

		if !logQueries {
			t.Error("Expected logQueries to be true")
		}
		if !logArgs {
			t.Error("Expected logArgs to be true")
		}
		if len(logArgsCopy) != len(args) {
			t.Errorf("Expected args length %d, got %d", len(args), len(logArgsCopy))
		}
		// First two should be unchanged
		if logArgsCopy[0] != "John" {
			t.Errorf("Index 0: expected 'John', got %v", logArgsCopy[0])
		}
		if logArgsCopy[1] != "john@example.com" {
			t.Errorf("Index 1: expected 'john@example.com', got %v", logArgsCopy[1])
		}
		// Third should be masked
		if logArgsCopy[2] != "[REDACTED]" {
			t.Errorf("Index 2: expected '[REDACTED]', got %v", logArgsCopy[2])
		}
		// Original args should not be modified
		if args[2] != "password123" {
			t.Error("Original args were modified")
		}
	})

	t.Run("masking multiple indices", func(t *testing.T) {
		ctx := WithMaskIndices(ctx, []int{0, 2})
		args := []any{"secret1", "john@example.com", "secret2"}
		logQueries, logArgs, logArgsCopy := getLoggingFlagsAndArgs(ctx, true, true, args)

		if !logQueries {
			t.Error("Expected logQueries to be true")
		}
		if !logArgs {
			t.Error("Expected logArgs to be true")
		}
		if logArgsCopy[0] != "[REDACTED]" {
			t.Errorf("Index 0: expected '[REDACTED]', got %v", logArgsCopy[0])
		}
		if logArgsCopy[1] != "john@example.com" {
			t.Errorf("Index 1: expected 'john@example.com', got %v", logArgsCopy[1])
		}
		if logArgsCopy[2] != "[REDACTED]" {
			t.Errorf("Index 2: expected '[REDACTED]', got %v", logArgsCopy[2])
		}
	})

	t.Run("masking with LogArgs=false (should not mask)", func(t *testing.T) {
		ctx := WithMaskIndices(ctx, []int{2})
		args := []any{"John", "john@example.com", "password123"}
		logQueries, logArgs, logArgsCopy := getLoggingFlagsAndArgs(ctx, true, false, args)

		if !logQueries {
			t.Error("Expected logQueries to be true")
		}
		if logArgs {
			t.Error("Expected logArgs to be false")
		}
		// When LogArgs=false, masking is not applied, but args are still returned
		// (they just won't be logged). The function returns the original args.
		if len(logArgsCopy) != len(args) {
			t.Errorf("Expected args length %d, got %d", len(args), len(logArgsCopy))
		}
		// Args should not be masked when LogArgs=false
		if logArgsCopy[2] == "[REDACTED]" {
			t.Error("Expected args not to be masked when LogArgs=false")
		}
		if logArgsCopy[2] != "password123" {
			t.Errorf("Expected password123, got %v", logArgsCopy[2])
		}
	})

	t.Run("masking with LogQueries=false and LogArgs=true", func(t *testing.T) {
		ctx := WithMaskIndices(ctx, []int{2})
		args := []any{"John", "john@example.com", "password123"}
		logQueries, logArgs, logArgsCopy := getLoggingFlagsAndArgs(ctx, false, true, args)

		if logQueries {
			t.Error("Expected logQueries to be false")
		}
		if !logArgs {
			t.Error("Expected logArgs to be true")
		}
		// Args should still be masked even if queries are not logged
		if len(logArgsCopy) != len(args) {
			t.Errorf("Expected args length %d, got %d", len(args), len(logArgsCopy))
		}
		if logArgsCopy[2] != "[REDACTED]" {
			t.Errorf("Index 2: expected '[REDACTED]', got %v", logArgsCopy[2])
		}
	})
}
