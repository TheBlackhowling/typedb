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
		name  string
		value any
		want  int
	}{
		{"int (zero)", 0, 0},
		{"int (small positive)", 123, 123},
		{"int (small negative)", -123, -123},
		{"int (max)", maxInt, maxInt},
		{"int (min)", minInt, minInt},
		{"int32 (max)", maxInt32, int(maxInt32)},
		{"int32 (min)", minInt32, int(minInt32)},
		{"int32 (small)", int32(456), 456},
		{"int16 (max)", int16(math.MaxInt16), int(math.MaxInt16)},
		{"int16 (min)", int16(math.MinInt16), int(math.MinInt16)},
		{"int16 (small)", int16(789), 789},
		{"int8 (max)", int8(math.MaxInt8), int(math.MaxInt8)},
		{"int8 (min)", int8(math.MinInt8), int(math.MinInt8)},
		{"int8 (small)", int8(42), 42},
		{"uint32 (max)", uint32(math.MaxUint32), int(math.MaxUint32)},
		{"uint32 (small)", uint32(999), 999},
		{"uint16 (max)", uint16(math.MaxUint16), int(math.MaxUint16)},
		{"uint16 (small)", uint16(888), 888},
		{"uint8 (max)", uint8(math.MaxUint8), int(math.MaxUint8)},
		{"uint8 (small)", uint8(77), 77},
		{"uint (max valid)", uint(maxInt), maxInt},
		{"uint (small)", uint(555), 555},
		{"uint64 (max valid)", uint64(maxInt), maxInt},
		{"uint64 (small)", uint64(333), 333},
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
		name  string
		value any
		want  int64
	}{
		{"int64 (zero)", int64(0), 0},
		{"int64 (max)", maxInt64, maxInt64},
		{"int64 (min)", minInt64, minInt64},
		{"int64 (small)", int64(123), 123},
		{"int (max)", int(math.MaxInt), int64(math.MaxInt)},
		{"int (min)", int(math.MinInt), int64(math.MinInt)},
		{"int (small)", 456, 456},
		{"int32 (max)", maxInt32, int64(maxInt32)},
		{"int32 (min)", minInt32, int64(minInt32)},
		{"int32 (small)", int32(789), 789},
		{"int16 (max)", int16(math.MaxInt16), int64(math.MaxInt16)},
		{"int16 (min)", int16(math.MinInt16), int64(math.MinInt16)},
		{"int8 (max)", int8(math.MaxInt8), int64(math.MaxInt8)},
		{"int8 (min)", int8(math.MinInt8), int64(math.MinInt8)},
		{"uint32 (max)", maxUint32, int64(maxUint32)},
		{"uint32 (small)", uint32(999), 999},
		{"uint16 (max)", uint16(math.MaxUint16), int64(math.MaxUint16)},
		{"uint8 (max)", uint8(math.MaxUint8), int64(math.MaxUint8)},
		{"uint (max valid)", uint(maxUint64Valid), int64(maxUint64Valid)},
		{"uint (small)", uint(555), 555},
		{"uint64 (max valid)", maxUint64Valid, int64(maxUint64Valid)},
		{"uint64 (small)", uint64(333), 333},
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
		name  string
		value any
		want  uint32
	}{
		{"uint32 (zero)", uint32(0), 0},
		{"uint32 (max)", maxUint32, maxUint32},
		{"uint32 (small)", uint32(123), 123},
		{"uint16 (max)", maxUint16, uint32(maxUint16)},
		{"uint16 (small)", uint16(456), 456},
		{"uint8 (max)", maxUint8, uint32(maxUint8)},
		{"uint8 (small)", uint8(77), 77},
		{"int32 (max valid)", maxInt32, uint32(maxInt32)},
		{"int32 (small positive)", int32(555), 555},
		{"int32 (zero)", int32(0), 0},
		{"int (max valid)", maxIntValid, uint32(maxIntValid)},
		{"int (small positive)", 999, 999},
		{"int (zero)", 0, 0},
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
	maxUintValid := uint(maxInt32) // Largest uint that fits in int32
	
	tests := []struct {
		name  string
		value any
		want  int32
	}{
		{"int32 (zero)", int32(0), 0},
		{"int32 (max)", maxInt32, maxInt32},
		{"int32 (min)", minInt32, minInt32},
		{"int32 (small)", int32(123), 123},
		{"int16 (max)", maxInt16, int32(maxInt16)},
		{"int16 (min)", minInt16, int32(minInt16)},
		{"int16 (small)", int16(456), 456},
		{"int8 (max)", maxInt8, int32(maxInt8)},
		{"int8 (min)", minInt8, int32(minInt8)},
		{"int8 (small)", int8(77), 77},
		{"int (max valid)", int(maxInt32), maxInt32},
		{"int (min valid)", int(minInt32), minInt32},
		{"int (small)", 999, 999},
		{"int (small negative)", -999, -999},
		{"int64 (max valid)", int64(maxInt32), maxInt32},
		{"int64 (min valid)", int64(minInt32), minInt32},
		{"uint32 (max valid)", maxUint32Valid, maxInt32},
		{"uint32 (small)", uint32(123), 123},
		{"uint16 (max)", maxUint16, int32(maxUint16)},
		{"uint16 (small)", uint16(555), 555},
		{"uint8 (max)", maxUint8, int32(maxUint8)},
		{"uint8 (small)", uint8(33), 33},
		{"uint (max valid)", maxUintValid, maxInt32},
		{"uint (small)", uint(123), 123},
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
