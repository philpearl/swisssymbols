// Package swisssymbols maps strings to integers. The first string maps to 1, the
// second to 2, etc. It also maps back from integers to strings. It is designed to be
// very memory efficient and fast. It holds all data off-heap.
//
// It is based on the swiss-table design for a hash table.
package swisssymbols

import (
	"unsafe"

	"github.com/philpearl/mmap"
	stringbank "github.com/philpearl/stringbank/offheap"
)

type SymbolTab struct {
	tables []*table

	spareTable      *table
	sb              stringbank.Stringbank
	ib              intbank
	count           int
	tableCount      int
	tableIndexShift uint16
	tableIndexDepth uint16
}

func New() *SymbolTab {
	m := SymbolTab{
		tableIndexShift: hashBits,
	}

	var err error
	m.tables, err = mmap.Alloc[*table](1)
	if err != nil {
		panic(err)
	}
	m.tables[0] = m.newTable()

	return &m
}

func (m *SymbolTab) Close() {
	m.sb.Close()
	m.ib.close()
	for _, t := range m.tables {
		m.freeTable(t)
	}
	if m.spareTable != nil {
		m.freeTable(m.spareTable)
	}
	mmap.Free(m.tables)
}

// Len returns the number of unique strings stored
func (m *SymbolTab) Len() int {
	return m.count
}

// Cap returns the size of the SymbolTab table
func (m *SymbolTab) Cap() int {
	return m.tableCount * tableSize * groupSize
}

// SymbolSize contains the approximate size of string storage in the symboltable. This will be an over-estimate and
// includes as yet unused and wasted space
func (m *SymbolTab) SymbolSize() int {
	return m.sb.Size()
}

// SequenceToString looks up a string by its sequence number. Obtain the sequence number
// for a string with StringToSequence
func (m *SymbolTab) SequenceToString(seq uint32) string {
	// Look up the stringbank offset for this sequence number, then get the string
	offset := m.ib.lookup(seq)
	return m.sb.Get(offset)
}

// Experiments to try
// - [X] Lower growth threshold - seems to help!
// - [X] use full hash - wildly better!
// - [X] bigger machine without these above changes (was it swapping?) - much slower again
// - [X] Just use full hash, not lower growth threshold - seems good
// - [ ] instrinsics for bit ops - can't make a version that works without AVX512
// - [X] different probe sequence. Maybe a bit better?
// - [ ] prefetch next group in probe sequence

const growthThreshold = tableSize * groupSize * 3 / 4

// StringToSequence looks up the string val and returns its sequence number seq. If val does
// not currently exist in the symbol table, it will add it if addNew is true. found indicates
// whether val was already present in the SymbolTab
func (m *SymbolTab) StringToSequence(val string, addNew bool) (seq uint32, found bool) {
	hash := hash(val)
	t := m.tables[hash>>hashValue(m.tableIndexShift)]
	if t == nil {
		// remove repeated nilcheck by checking here
		panic("nil table found in map")
	}

	probe := makeProbeSeq(hash>>7, tableMask)
	groupHash := hashValue(hash & 0x7F)
	for range t.groups {
		group := t.groups.getGroup(probe.offset)
		matches := group.control.findMatches(groupHash)
		for matches != 0 {
			index := matches.firstSet()
			// This horrendous line gets the entry at index without doing a bounds check or nil check
			ent := (*entry)(unsafe.Add(unsafe.Pointer(&group.entries), uintptr(index)*unsafe.Sizeof(entry{})))
			if ent.hash == hash {
				if seq := ent.seq; m.sb.Get(m.ib.lookup(seq)) == val {
					return ent.seq, true
				}
			}
			matches = matches.clearFirstBit()
		}

		if empty := group.control.findEmpty(); empty != 0 {
			// There is an empty slot, so we've reached the end of the probe
			// sequence and the key is not present in the map.
			if !addNew {
				return 0, false
			}

			index := empty.firstSet()
			m.count++
			seq = uint32(m.count)
			m.ib.save(seq, m.sb.Save(val))

			// This horrendous line sets the entry at index without doing a bounds check or nil check
			*(*entry)(unsafe.Add(unsafe.Pointer(&group.entries), uintptr(index)*unsafe.Sizeof(entry{}))) = entry{seq: seq, hash: hash}
			group.control.set(index, groupHash)
			t.used++
			if t.used > growthThreshold {
				// Table is too full, need to grow
				m.onGrowthNeeded(t)
			}

			return seq, false
		}
		// Continue to next group in case of hash collision
		// TODO: try a different probe sequence
		probe = probe.next()
	}
	panic("table is full")
}

func (m *SymbolTab) newTable() *table {
	m.tableCount++
	if m.spareTable != nil {
		t := m.spareTable
		m.spareTable = nil
		return t
	}
	tables, err := mmap.Alloc[table](1)
	if err != nil {
		panic(err)
	}
	t := &tables[0]
	t.init()
	return t
}

func (m *SymbolTab) freeTable(t *table) {
	m.tableCount--
	if m.spareTable == nil {
		t.init()
		m.spareTable = t
		return
	}
	if err := mmap.Free(unsafe.Slice(t, 1)); err != nil {
		panic(err)
	}
}

// This is called when a table detects it is too full and needs to grow.
func (m *SymbolTab) onGrowthNeeded(t *table) {
	if t.localDepth == m.tableIndexDepth {
		// Need to grow the directory. This will take care of splitting tables as needed.
		m.grow()
	}

	// We can just split this table, and split up the slots it is currently
	// installed in in the directory.
	oldTab, newTab := t.split(m)
	m.insertTable(oldTab)
	m.insertTable(newTab)
	m.freeTable(t)
}

func (m *SymbolTab) insertTable(t *table) {
	depthDifference := m.tableIndexDepth - t.localDepth
	index := t.index * (depthDifference + 1)
	tableWidth := uint16(1 << depthDifference)
	for i := range tableWidth {
		m.tables[index+i] = t
	}
}

// grow grows the map by splitting tables as needed. We always double the number
// of entries in the table index, but only split tables as needed. If we don't
// need to split a table we double the number of entries that point to the same
// table.
func (m *SymbolTab) grow() {
	newTables, err := mmap.Alloc[*table](len(m.tables) * 2)
	if err != nil {
		panic(err)
	}
	for i, table := range m.tables {
		newTables[i*2] = table
		newTables[i*2+1] = table
	}
	m.tableIndexShift--
	m.tableIndexDepth++
	mmap.Free(m.tables)
	m.tables = newTables
}
