//go:build goexperiment.simd && amd64

package swisssymbols

import "math/bits"

// groupBits is a bitmask representing matches in a group. A match is indicated by the
// top bit in the byte being set.
type groupBits uint16

// firstSet returns the index of the first set bit in the bits mask. If no bits
// are set, it returns 8.
func (b groupBits) firstSet() int {
	// TrailingZeros will return 7 if the first bit is set, 15 if the second bit is
	// set, etc. So we divide by 8 to get the index.

	return bits.TrailingZeros16(uint16(b))
}

// clearFirstBit clears the least significant set bit and returns the result.
func (b groupBits) clearFirstBit() groupBits {
	return b & (b - 1)
}
