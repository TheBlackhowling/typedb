package typedb

import (
	"math"
	"testing"
)

func TestDeserializeInt_OverflowErrors(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	maxInt64 := int64(math.MaxInt64)

	tests := []struct {
		name    string
		value   any
		wantErr string
	}{
		{
			name:    "uint overflow to int",
			value:   uint(math.MaxUint64),
			wantErr: "overflows int",
		},
		{
			name:    "uint64 overflow to int",
			value:   uint64(math.MaxUint64),
			wantErr: "overflows int",
		},
	}

	// Only test int64 overflow if int is smaller than int64 (32-bit systems)
	if int64(maxInt) < maxInt64 {
		tests = append(tests, struct {
			name    string
			value   any
			wantErr string
		}{
			name:    "int64 overflow to int (positive)",
			value:   int64(math.MaxInt64),
			wantErr: "overflows int",
		})
		tests = append(tests, struct {
			name    string
			value   any
			wantErr string
		}{
			name:    "int64 overflow to int (negative)",
			value:   int64(math.MinInt64),
			wantErr: "overflows int",
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := deserializeInt(tt.value)
			if err == nil {
				t.Fatal("Expected error for overflow, got nil")
			}
			if err.Error() == "" {
				t.Fatal("Expected error message, got empty string")
			}
			// Check that error message contains expected text
			if err.Error()[:len("typedb: ")] != "typedb: " {
				t.Errorf("Expected error to start with 'typedb: ', got: %q", err.Error())
			}
		})
	}
}

func TestDeserializeInt64_OverflowErrors(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr string
	}{
		{
			name:    "uint64 overflow to int64",
			value:   uint64(math.MaxUint64),
			wantErr: "overflows int64",
		},
		{
			name:    "uint overflow to int64 (on 64-bit systems)",
			value:   uint(math.MaxUint64),
			wantErr: "overflows int64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := deserializeInt64(tt.value)
			if err == nil {
				t.Fatal("Expected error for overflow, got nil")
			}
			if err.Error() == "" {
				t.Fatal("Expected error message, got empty string")
			}
			if err.Error()[:len("typedb: ")] != "typedb: " {
				t.Errorf("Expected error to start with 'typedb: ', got: %q", err.Error())
			}
		})
	}
}

func TestDeserializeUint32_OverflowErrors(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr string
	}{
		{
			name:    "uint overflow to uint32",
			value:   uint(math.MaxUint64),
			wantErr: "overflows uint32",
		},
		{
			name:    "int overflow to uint32 (negative)",
			value:   -1,
			wantErr: "cannot convert negative int to uint32",
		},
		{
			name:    "int overflow to uint32 (too large)",
			value:   int(math.MaxInt64),
			wantErr: "overflows uint32",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := deserializeUint32(tt.value)
			if err == nil {
				t.Fatal("Expected error for overflow, got nil")
			}
			if err.Error() == "" {
				t.Fatal("Expected error message, got empty string")
			}
			if err.Error()[:len("typedb: ")] != "typedb: " {
				t.Errorf("Expected error to start with 'typedb: ', got: %q", err.Error())
			}
		})
	}
}

func TestDeserializeInt32_OverflowErrors(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr string
	}{
		{
			name:    "int overflow to int32 (positive)",
			value:   int(math.MaxInt64),
			wantErr: "overflows int32",
		},
		{
			name:    "int overflow to int32 (negative)",
			value:   int(math.MinInt64),
			wantErr: "overflows int32",
		},
		{
			name:    "int64 overflow to int32 (positive)",
			value:   int64(math.MaxInt64),
			wantErr: "overflows int32",
		},
		{
			name:    "int64 overflow to int32 (negative)",
			value:   int64(math.MinInt64),
			wantErr: "overflows int32",
		},
		{
			name:    "uint32 overflow to int32",
			value:   uint32(math.MaxUint32),
			wantErr: "overflows int32",
		},
		{
			name:    "uint overflow to int32",
			value:   uint(math.MaxUint64),
			wantErr: "overflows int32",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := deserializeInt32(tt.value)
			if err == nil {
				t.Fatal("Expected error for overflow, got nil")
			}
			if err.Error() == "" {
				t.Fatal("Expected error message, got empty string")
			}
			if err.Error()[:len("typedb: ")] != "typedb: " {
				t.Errorf("Expected error to start with 'typedb: ', got: %q", err.Error())
			}
		})
	}
}

func TestConvertUintSlice_OverflowHandling(t *testing.T) {
	// Test that convertUintSlice handles overflow gracefully
	// It should set overflow values to 0
	largeUint := uint(math.MaxUint64)
	smallUint := uint(123)

	v := []uint{smallUint, largeUint, smallUint}
	result := convertUintSlice(v)

	if len(result) != len(v) {
		t.Fatalf("Expected result length %d, got %d", len(v), len(result))
	}

	if result[0] != int(smallUint) {
		t.Errorf("Expected result[0] = %d, got %d", int(smallUint), result[0])
	}

	// Overflow value should be set to 0
	if result[1] != 0 {
		t.Errorf("Expected overflow value to be 0, got %d", result[1])
	}

	if result[2] != int(smallUint) {
		t.Errorf("Expected result[2] = %d, got %d", int(smallUint), result[2])
	}
}

func TestConvertUint64Slice_OverflowHandling(t *testing.T) {
	// Test that convertUint64Slice handles overflow gracefully
	// It should set overflow values to 0
	largeUint64 := uint64(math.MaxUint64)
	smallUint64 := uint64(123)

	v := []uint64{smallUint64, largeUint64, smallUint64}
	result := convertUint64Slice(v)

	if len(result) != len(v) {
		t.Fatalf("Expected result length %d, got %d", len(v), len(result))
	}

	if result[0] != int(smallUint64) {
		t.Errorf("Expected result[0] = %d, got %d", int(smallUint64), result[0])
	}

	// Overflow value should be set to 0
	if result[1] != 0 {
		t.Errorf("Expected overflow value to be 0, got %d", result[1])
	}

	if result[2] != int(smallUint64) {
		t.Errorf("Expected result[2] = %d, got %d", int(smallUint64), result[2])
	}
}

// Test that valid conversions still work (no false positives)
func TestDeserializeInt_ValidConversions(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	minInt := ^maxInt
	maxInt32 := int32(math.MaxInt32)
	minInt32 := int32(math.MinInt32)

	tests := []struct {
		value any    // 16 bytes (interface{})
		name  string // 16 bytes
		want  int    // 8 bytes
	}{
		{0, "int (zero)", 0},
		{123, "int (small positive)", 123},
		{-123, "int (small negative)", -123},
		{maxInt, "int (max)", maxInt},
		{minInt, "int (min)", minInt},
		{maxInt32, "int32 (max)", int(maxInt32)},
		{minInt32, "int32 (min)", int(minInt32)},
		{int32(456), "int32 (small)", 456},
		{int16(math.MaxInt16), "int16 (max)", int(math.MaxInt16)},
		{int16(math.MinInt16), "int16 (min)", int(math.MinInt16)},
		{int16(789), "int16 (small)", 789},
		{int8(math.MaxInt8), "int8 (max)", int(math.MaxInt8)},
		{int8(math.MinInt8), "int8 (min)", int(math.MinInt8)},
		{int8(42), "int8 (small)", 42},
		{uint32(math.MaxUint32), "uint32 (max)", int(math.MaxUint32)},
		{uint32(999), "uint32 (small)", 999},
		{uint16(math.MaxUint16), "uint16 (max)", int(math.MaxUint16)},
		{uint16(888), "uint16 (small)", 888},
		{uint8(math.MaxUint8), "uint8 (max)", int(math.MaxUint8)},
		{uint8(77), "uint8 (small)", 77},
		{uint(maxInt), "uint (max valid)", maxInt},
		{uint(555), "uint (small)", 555},
		{uint64(maxInt), "uint64 (max valid)", maxInt},
		{uint64(333), "uint64 (small)", 333},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := deserializeInt(tt.value)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Expected %d, got %d", tt.want, got)
			}
		})
	}
}

func TestDeserializeInt64_ValidConversions(t *testing.T) {
	maxInt64 := int64(math.MaxInt64)
	minInt64 := int64(math.MinInt64)
	maxInt32 := int32(math.MaxInt32)
	minInt32 := int32(math.MinInt32)
	maxUint32 := uint32(math.MaxUint32)
	maxUint64Valid := uint64(maxInt64) // Largest uint64 that fits in int64

	tests := []struct {
		value any    // 16 bytes (interface{})
		name  string // 16 bytes
		want  int64  // 8 bytes
	}{
		{value: int64(0), name: "int64 (zero)", want: 0},
		{value: maxInt64, name: "int64 (max)", want: maxInt64},
		{value: minInt64, name: "int64 (min)", want: minInt64},
		{value: int64(123), name: "int64 (small)", want: 123},
		{value: int(math.MaxInt), name: "int (max)", want: int64(math.MaxInt)},
		{value: int(math.MinInt), name: "int (min)", want: int64(math.MinInt)},
		{value: 456, name: "int (small)", want: 456},
		{value: maxInt32, name: "int32 (max)", want: int64(maxInt32)},
		{value: minInt32, name: "int32 (min)", want: int64(minInt32)},
		{value: int32(789), name: "int32 (small)", want: 789},
		{value: int16(math.MaxInt16), name: "int16 (max)", want: int64(math.MaxInt16)},
		{value: int16(math.MinInt16), name: "int16 (min)", want: int64(math.MinInt16)},
		{value: int8(math.MaxInt8), name: "int8 (max)", want: int64(math.MaxInt8)},
		{value: int8(math.MinInt8), name: "int8 (min)", want: int64(math.MinInt8)},
		{value: maxUint32, name: "uint32 (max)", want: int64(maxUint32)},
		{value: uint32(999), name: "uint32 (small)", want: 999},
		{value: uint16(math.MaxUint16), name: "uint16 (max)", want: int64(math.MaxUint16)},
		{value: uint8(math.MaxUint8), name: "uint8 (max)", want: int64(math.MaxUint8)},
		{value: uint(maxUint64Valid), name: "uint (max valid)", want: int64(maxUint64Valid)},
		{value: uint(555), name: "uint (small)", want: 555},
		{value: maxUint64Valid, name: "uint64 (max valid)", want: int64(maxUint64Valid)},
		{value: uint64(333), name: "uint64 (small)", want: 333},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := deserializeInt64(tt.value)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Expected %d, got %d", tt.want, got)
			}
		})
	}
}

func TestDeserializeUint32_ValidConversions(t *testing.T) {
	maxUint32 := uint32(math.MaxUint32)
	maxUint16 := uint16(math.MaxUint16)
	maxUint8 := uint8(math.MaxUint8)
	maxInt32 := int32(math.MaxInt32)
	maxIntValid := int(maxUint32) // Largest int that fits in uint32

	tests := []struct {
		value any    // 16 bytes (interface{})
		name  string // 16 bytes
		want  uint32 // 4 bytes
	}{
		{uint32(0), "uint32 (zero)", 0},
		{maxUint32, "uint32 (max)", maxUint32},
		{uint32(123), "uint32 (small)", 123},
		{maxUint16, "uint16 (max)", uint32(maxUint16)},
		{uint16(456), "uint16 (small)", 456},
		{maxUint8, "uint8 (max)", uint32(maxUint8)},
		{uint8(77), "uint8 (small)", 77},
		{maxInt32, "int32 (max valid)", uint32(maxInt32)},
		{int32(555), "int32 (small positive)", 555},
		{int32(0), "int32 (zero)", 0},
		{maxIntValid, "int (max valid)", uint32(maxIntValid)},
		{999, "int (small positive)", 999},
		{0, "int (zero)", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := deserializeUint32(tt.value)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Expected %d, got %d", tt.want, got)
			}
		})
	}
}

func TestDeserializeInt32_ValidConversions(t *testing.T) {
	maxInt32 := int32(math.MaxInt32)
	minInt32 := int32(math.MinInt32)
	maxInt16 := int16(math.MaxInt16)
	minInt16 := int16(math.MinInt16)
	maxInt8 := int8(math.MaxInt8)
	minInt8 := int8(math.MinInt8)
	maxUint16 := uint16(math.MaxUint16)
	maxUint8 := uint8(math.MaxUint8)
	maxUint32Valid := uint32(maxInt32) // Largest uint32 that fits in int32
	maxUintValid := uint(maxInt32)     // Largest uint that fits in int32

	tests := []struct {
		value any    // 16 bytes (interface{})
		name  string // 16 bytes
		want  int32  // 4 bytes
	}{
		{int32(0), "int32 (zero)", 0},
		{maxInt32, "int32 (max)", maxInt32},
		{minInt32, "int32 (min)", minInt32},
		{int32(123), "int32 (small)", 123},
		{maxInt16, "int16 (max)", int32(maxInt16)},
		{minInt16, "int16 (min)", int32(minInt16)},
		{int16(456), "int16 (small)", 456},
		{maxInt8, "int8 (max)", int32(maxInt8)},
		{minInt8, "int8 (min)", int32(minInt8)},
		{int8(77), "int8 (small)", 77},
		{int(maxInt32), "int (max valid)", maxInt32},
		{int(minInt32), "int (min valid)", minInt32},
		{999, "int (small)", 999},
		{-999, "int (small negative)", -999},
		{int64(maxInt32), "int64 (max valid)", maxInt32},
		{int64(minInt32), "int64 (min valid)", minInt32},
		{maxUint32Valid, "uint32 (max valid)", maxInt32},
		{uint32(123), "uint32 (small)", 123},
		{maxUint16, "uint16 (max)", int32(maxUint16)},
		{uint16(555), "uint16 (small)", 555},
		{maxUint8, "uint8 (max)", int32(maxUint8)},
		{uint8(33), "uint8 (small)", 33},
		{maxUintValid, "uint (max valid)", maxInt32},
		{uint(123), "uint (small)", 123},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := deserializeInt32(tt.value)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Expected %d, got %d", tt.want, got)
			}
		})
	}
}
