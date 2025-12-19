package swisssymbols

const groupSize = 8

type group struct {
	control groupControl
	entries [groupSize]entry
}

type entry struct {
	hash hashValue
	seq  uint32
}
