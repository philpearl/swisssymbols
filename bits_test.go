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
			expected: 8, // TrailingZeros returns 64 if no bits are set
		},
		{
			name:     "top bit set",
			bits:     0x8000_0000_0000_0000,
			expected: 7,
		},
		{
			name:     "second top bit set",
			bits:     0x0080_0000_0000_0000,
			expected: 6,
		},
		{
			name:     "third top bit set",
			bits:     0x0000_8000_0000_0000,
			expected: 5,
		},
		{
			name:     "4th top bit set",
			bits:     0x0000_0080_0000_0000,
			expected: 4,
		},
		{
			name:     "5th top bit set",
			bits:     0x0000_0000_8000_0000,
			expected: 3,
		},
		{
			name:     "6th top bit set",
			bits:     0x0000_0000_0080_0000,
			expected: 2,
		},

		{
			name:     "second bottom bit set",
			bits:     0x8000,
			expected: 1,
		},

		{
			name:     "bottom bit set",
			bits:     0x80,
			expected: 0,
		},
		{
			name:     "multiple bits set",
			bits:     0x8080_8080_8080_8080,
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
	bits := groupBits(0x8080_8080_8080_8080)

	exp := 0
	for {
		index := bits.firstSet()
		if index != exp {
			t.Errorf("expected index %d, got %d", exp, index)
		}

		if index == 8 {
			break
		}
		bits = bits.clearFirstBit()
		exp++
	}
}
