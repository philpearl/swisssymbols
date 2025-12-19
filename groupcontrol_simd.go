//go:build goexperiment.simd && amd64

package swisssymbols

import "simd/archsimd"

type groupControl [groupSize]uint8

var (
	emptyGroupControl = archsimd.BroadcastUint8x16(0x80)
	fullGroupControl  = archsimd.BroadcastUint8x16(0x00)
)

const (
	controlHashMask    = 0x7F
	groupControlExpand = 0x0101010101010101
)

func (g *group) init() {
	g.control = [groupSize]uint8{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80}
}

// findMatches returns a bits mask of which entries in the group match the given
// hash value.
func (gc groupControl) findMatches(hash hashValue) groupBits {
	ctrlHash := byte(hash & controlHashMask)
	vec := archsimd.LoadUint8x16SlicePart(gc[:])
	lookingFor := archsimd.BroadcastUint8x16(ctrlHash)
	return groupBits(vec.Equal(lookingFor).ToBits())
}

// findEmpty returns a bits mask of which entries in the group are empty.
func (gc groupControl) findEmpty() groupBits {
	vec := archsimd.LoadUint8x16SlicePart(gc[:])
	return groupBits(vec.And(emptyGroupControl).Equal(emptyGroupControl).ToBits())
}

func (gc groupControl) findFull() groupBits {
	vec := archsimd.LoadUint8x16SlicePart(gc[:])
	return groupBits(vec.AndNot(emptyGroupControl).Equal(fullGroupControl).ToBits())
}

func (gc *groupControl) set(index int, hash byte) {
	(*gc)[index] = hash
}
