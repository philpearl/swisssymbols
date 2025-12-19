//go:build goexperiment.simd && amd64

package swisssymbols

import "testing"

func TestBitsFirstSet(t *testing.T) {
	tests := []struct {
		name     string
		bits     groupBits
		expected int
	}{
		{
			name:     "no bits set",
			bits:     0x0,
			expected: 16, // TrailingZeros returns top value if no bits are set
		},
		{
			name:     "top bit set",
			bits:     0b10000000,
			expected: 7,
		},
		{
			name:     "second top bit set",
			bits:     0b01000000,
			expected: 6,
		},
		{
			name:     "third top bit set",
			bits:     0b00100000,
			expected: 5,
		},
		{
			name:     "4th top bit set",
			bits:     0b00010000,
			expected: 4,
		},
		{
			name:     "5th top bit set",
			bits:     0b00001000,
			expected: 3,
		},
		{
			name:     "6th top bit set",
			bits:     0b00000100,
			expected: 2,
		},

		{
			name:     "second bottom bit set",
			bits:     0b00000010,
			expected: 1,
		},

		{
			name:     "bottom bit set",
			bits:     0b00000001,
			expected: 0,
		},
		{
			name:     "multiple bits set",
			bits:     0b11111111,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.bits.firstSet()
			if result != tt.expected {
				t.Errorf("expected index %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestBitsEnumerate(t *testing.T) {
	bits := groupBits(0b11111111)

	exp := 0
	for {
		index := bits.firstSet()
		if index != exp {
			t.Errorf("expected index %d, got %d", exp, index)
		}

		if index == 16 {
			break
		}
		bits = bits.clearFirstBit()
		exp++
	}
}
