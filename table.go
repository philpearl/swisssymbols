package swisssymbols

import (
	"unsafe"
)

type hashValue uint32

const tableSize = 4096

// Table is a fixed-size hash table containing groups.
type table struct {
	groups groups

	// localDepth is the number of bits of the hash used to pick this table in
	// the extensible hashing scheme.
	localDepth int
	// used is the number of entries in the table
	used int
	// This is the index of this table in the map's table index.
	index int
}

type groups [tableSize]group

func (t *table) init() {
	if t == nil {
		panic("initializing nil table")
	}
	for i := range t.groups {
		t.groups[i].init()
	}
	t.localDepth = 0
	t.used = 0
	t.index = 0
}

func (gs *groups) getGroup(i hashValue) *group {
	return (*group)(unsafe.Add(unsafe.Pointer(gs), uintptr(i)*unsafe.Sizeof(group{})))
}

func hash(key string) hashValue {
	return hashValue(runtime_memhash(
		unsafe.Pointer(unsafe.StringData(key)),
		0,
		uintptr(len(key)),
	))
}

// We use the runtime's map hash function without the overhead of using
// hash/maphash
//
//go:linkname runtime_memhash runtime.memhash
//go:noescape
func runtime_memhash(p unsafe.Pointer, seed, s uintptr) uintptr

func (t *table) insert(m *Map, ent entry) {
	if t == nil {
		panic("inserting into nil table")
	}
	l := hashValue(len(t.groups))

	groupIndex := (ent.hash >> 7) % l
	for range t.groups {
		group := t.groups.getGroup(groupIndex)
		// We're not looking for matches, only empty spaces
		if empty := group.control.findEmpty(); empty != 0 {
			// There is an empty slot, so we've reached the end of the probe
			// sequence and the key is not present in the map.

			// This horrendous line sets the entry at index without doing a bounds check or nil check
			*(*entry)(unsafe.Add(unsafe.Pointer(&group.entries), uintptr(empty.firstSet())*unsafe.Sizeof(entry{}))) = ent

			group.control = (group.control &^ (groupControl(0x80) << (empty.firstSet() * 8))) | groupControl(byte(ent.hash&0x7F))<<(empty.firstSet()*8)
			t.used++
			if t.used > growthThreshold {
				// Table is too full, need to grow
				m.onGrowthNeeded(t)
			}
			return
		}
		// Continue to next group in case of hash collision
		// TODO: try a different probe sequence
		groupIndex = (groupIndex + 1) % l
	}
	panic("table is full")
}

// split splits the table, returning a new table containing hopefully half of
// the entries.
func (t *table) split(m *Map) (oldTab, newTab *table) {
	// We create two whole new tables rather than reusing the existing one,
	// because we can't enumerate over the old table and modify it at the same
	// time.
	//
	// We do re-use the old table next time we grow - it gets cleared and put
	// into the spare table pool (which has 1 slot!).
	if t == nil {
		panic("splitting nil table")
	}
	newTab = m.newTable()
	oldTab = m.newTable()
	oldTab.localDepth = t.localDepth + 1
	oldTab.index = t.index * 2
	newTab.localDepth = t.localDepth + 1
	newTab.index = t.index*2 + 1

	// We create a new table, then split the data in the current table between
	// the current table and the new, based on the hash bit that the new local
	// depth exposes.
	mask := hashValue(1 << (32 - t.localDepth - 1))

	for i := range t.groups {
		group := t.groups.getGroup(hashValue(i))
		// Find all the used entries in this group and iterate over them
		matches := group.control.findFull()
		for matches != 0 {
			index := matches.firstSet()
			// This horrendous line gets the entry at index without doing a bounds check or nil check
			ent := *(*entry)(unsafe.Add(unsafe.Pointer(&group.entries), uintptr(index)*unsafe.Sizeof(entry{})))

			// We need to recalculate the hash so that we can find the correct
			// bit to decide what to split. We also need to re-insert the entry
			// into the tables, and the location won't be the same because of
			// probing.
			tab := oldTab
			if ent.hash&mask != 0 {
				tab = newTab
			}
			tab.insert(m, ent)

			matches = matches.clearFirstBit()
		}
	}

	return oldTab, newTab
}
