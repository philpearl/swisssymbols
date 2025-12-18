package swisssymbols

import "unsafe"

const groupSize = 8

type group struct {
	control groupControl
	entries [groupSize]entry
}

type entry struct {
	seq  uint32
	hash hashValue
}

type groupControl uint64

const (
	emptyGroupControl  = 0x8080808080808080
	controlHashMask    = 0x7F
	groupControlExpand = 0x0101010101010101
)

func (g *group) init() {
	g.control = emptyGroupControl
}

// findMatches returns a bits mask of which entries in the group match the given
// hash value.
func (gc groupControl) findMatches(hash hashValue) groupBits {
	ctrlHash := byte(hash & controlHashMask)
	// Find the entries where the control byte matches ctrlHash
	//
	// We expand the ctrlHash to a groupControl where each byte is ctrlHash,
	// then XOR that with the group control. Any byte that was equal will now be
	// zero. We then subtract 0x01 from each byte, so any byte that was zero
	// will now have its high bit set. Finally we AND with 0x80 to keep only the
	// high bits.
	//
	// Note this does give false positives!
	matchesAreZero := uint64(gc) ^ (uint64(ctrlHash) * groupControlExpand)
	return groupBits(((matchesAreZero - 0x0101010101010101) &^ matchesAreZero) & 0x8080808080808080)
}

// findEmpty returns a bits mask of which entries in the group are empty.
func (gc groupControl) findEmpty() groupBits {
	return groupBits(uint64(gc) & uint64(emptyGroupControl))
}

func (gc groupControl) findFull() groupBits {
	return groupBits(^uint64(gc) & uint64(emptyGroupControl))
}

func (gc *groupControl) set(index int, hash byte) {
	// Only works on little-endian systems, but all systems we see currently are.
	*(*byte)(unsafe.Add(unsafe.Pointer(gc), uintptr(index)*unsafe.Sizeof(byte(0)))) = hash
}
