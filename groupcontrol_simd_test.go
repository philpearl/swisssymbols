//go:build goexperiment.simd && amd64

package swisssymbols

import (
	"simd/archsimd"
	"testing"
)

func TestSIMDSupport(t *testing.T) {
	var cpu archsimd.X86Features
	if !cpu.AVX() {
		t.Error("expected AVX support")
	}
	if !cpu.AVX2() {
		t.Error("expected AVX2 support")
	}
	if !cpu.AVX512() {
		t.Error("expected AVX512 support")
	}
}

func TestGroupFindMatches(t *testing.T) {
	tests := []struct {
		name     string
		control  groupControl
		hash     hashValue
		expected groupBits
	}{
		{
			name:     "no matches",
			control:  groupControl{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			hash:     0x09,
			expected: 0x0,
		},
		{
			name:     "one match with expected false positive",
			control:  groupControl{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			hash:     0x23823803,
			expected: 0b00100000,
		},
		{
			name:     "multiple matches",
			control:  groupControl{0x02, 0x02, 0x03, 0x04, 0x02, 0x06, 0x07, 0x02},
			hash:     0x02,
			expected: 0b11001001,
		},
		{
			name:     "all matches",
			control:  groupControl{0x05, 0x05, 0x05, 0x05, 0x05, 0x05, 0x05, 0x05},
			hash:     0x05,
			expected: 0b11111111,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.control.findMatches(tt.hash)
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
			control:  groupControl{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			expected: 0x0,
		},
		{
			name:     "one empty",
			control:  groupControl{0x80, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			expected: 0b10000000,
		},
		{
			name:     "multiple empty",
			control:  groupControl{0x80, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x80},
			expected: 0b10000001,
		},
		{
			name:     "all empty",
			control:  groupControl{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80},
			expected: 0b11111111,
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
