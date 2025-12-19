//go:build !goexperiment.simd || !amd64

package swisssymbols

import "testing"

func TestGroupFindMatches(t *testing.T) {
	tests := []struct {
		name     string
		control  groupControl
		hash     hashValue
		expected groupBits
	}{
		{
			name:     "no matches",
			control:  0x0102030405060708,
			hash:     0x09,
			expected: 0x0,
		},
		{
			name:     "one match with expected false positive",
			control:  0x0102030405060708,
			hash:     0x23823803,
			expected: 0b0000000010000000100000000000000000000000000000000000000000000000,
			//          7654321076543210765432107654321076543210765432107654321076543210
			//          7.      6       5       4       3       2       1       0
		},
		{
			name:     "multiple matches",
			control:  0x0202030402060702,
			hash:     0x02,
			expected: 0b1000000010000000000000000000000010000000000000000000000010000000,
			//          7654321076543210765432107654321076543210765432107654321076543210
			//          7       6       5       4       3       2       1       0
		},
		{
			name:     "all matches",
			control:  0x0505050505050505,
			hash:     0x05,
			expected: 0b1000000010000000100000001000000010000000100000001000000010000000,
			//          7654321076543210765432107654321076543210765432107654321076543210
			//          7       6       5       4       3       2       1       0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.control.findMatches(tt.hash & 0x7F)
			if result != tt.expected {
				t.Errorf("expected bits %08b, got %08b", tt.expected, result)
			}
		})
	}
}

func TestGroupFindEmpty(t *testing.T) {
	tests := []struct {
		name     string
		control  groupControl
		expected groupBits
	}{
		{
			name:     "no empty",
			control:  0x0102030405060708,
			expected: 0x0,
		},
		{
			name:     "one empty",
			control:  0x8002030405060708,
			expected: 0b1000000000000000000000000000000000000000000000000000000000000000,
		},
		{
			name:     "multiple empty",
			control:  0x8002_0304_0506_0780,
			expected: 0b1000000000000000000000000000000000000000000000000000000010000000,
		},
		{
			name:     "all empty",
			control:  0x8080808080808080,
			expected: 0x8080808080808080,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.control.findEmpty()
			if result != tt.expected {
				t.Errorf("expected bits %08b, got %08b", tt.expected, result)
			}
		})
	}
}
