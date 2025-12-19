package swisssymbols

import (
	"fmt"
	"runtime"
	"strconv"
	"testing"
	"time"
	"unsafe"

	stringbank "github.com/philpearl/stringbank/offheap"
)

func TestSetGet(t *testing.T) {
	m := New()
	defer m.Close()

	var sb stringbank.Stringbank
	var buf []byte
	for i := range 1_000_000 {
		buf = fmt.Appendf(buf[:0], "key%d", i)
		sb.Save(unsafe.String(&buf[0], len(buf)))
	}

	i := 0
	for key := range sb.All() {
		seq, found := m.StringToSequence(key, true)
		if found {
			t.Fatalf("expected key %s to not be found", key)
		}
		i++
		if seq != uint32(i) {
			t.Fatalf("expected seq %d for key %s, got %d", i, key, seq)
		}
	}

	i = 0
	for key := range sb.All() {
		seq, found := m.StringToSequence(key, false)
		if !found {
			t.Fatalf("expected key %s to be found", key)
		}
		i++
		if seq != uint32(i) {
			t.Fatalf("expected seq %d for key %s, got %d", i, key, seq)
		}
	}

	i = 0
	for key := range sb.All() {
		i++
		actual := m.SequenceToString(uint32(i))
		if key != actual {
			t.Fatalf("expected key %s to be found", key)
		}
	}
}

func BenchmarkSetGet(b *testing.B) {
	var sb stringbank.Stringbank
	var buf []byte
	for i := range 10_000_000 {
		buf = append(buf[:0], "key"...)
		buf = strconv.AppendInt(buf, int64(i), 10)
		sb.Save(unsafe.String(&buf[0], len(buf)))
	}

	var i int
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		m := New()
		i = 0
		for key := range sb.All() {
			seq, found := m.StringToSequence(key, true)
			if found {
				b.Fatalf("expected key %s to not be found", key)
			}

			i++
			if seq != uint32(i) {
				b.Fatalf("expected seq %d for key %s, got %d", i, key, seq)
			}
		}

		i = 0
		for key := range sb.All() {
			i++
			actual := m.SequenceToString(uint32(i))
			if actual != key {
				b.Fatalf("unexpected for key %s, got %s", key, actual)
			}
		}
		m.Close()
		runtime.GC()
	}
}

func TestBasic(t *testing.T) {
	st := New()
	defer st.Close()

	assertStringToSequence := func(seq uint32, existing bool, val string) {
		t.Helper()
		seqa, existinga := st.StringToSequence(val, true)
		if existing != existinga {
			t.Errorf("for value %s, expected ( %v), got ( %v)", val, existing, existinga)
		}
		if existinga {
			if seq != seqa {
				t.Errorf("for value %s, expected seq %d, got %d", val, seq, seqa)
			}
		}
	}

	assertStringToSequence(1, false, "a1")
	assertStringToSequence(2, false, "a2")
	assertStringToSequence(3, false, "a3")
	assertStringToSequence(2, true, "a2")
	assertStringToSequence(3, true, "a3")

	if s := st.SequenceToString(1); s != "a1" {
		t.Errorf("expected string a1, got %s", s)
	}
	if s := st.SequenceToString(2); s != "a2" {
		t.Errorf("expected string a2, got %s", s)
	}
	if s := st.SequenceToString(3); s != "a3" {
		t.Errorf("expected string a3, got %s", s)
	}
}

func TestGrowth2(t *testing.T) {
	st := New()
	defer st.Close()

	for i := range 10000 {
		seq, found := st.StringToSequence(strconv.Itoa(i), true)
		if found {
			t.Fatalf("expected value %d to not be found", i)
		}
		if seq != uint32(i+1) {
			t.Fatalf("expected seq %d for value %d, got %d", i+1, i, seq)
		}

		seq, found = st.StringToSequence(strconv.Itoa(i), true)
		if !found {
			t.Fatalf("expected value %d to be found", i)
		}
		if seq != uint32(i+1) {
			t.Fatalf("expected seq %d for value %d, got %d", i+1, i, seq)
		}
	}
}

func TestAddNew(t *testing.T) {
	st := New()
	defer st.Close()
	// Won't add entry if asked not to
	seq, existing := st.StringToSequence("hat", false)
	if existing {
		t.Errorf("expected not to find entry")
	}
	if seq != 0 {
		t.Errorf("expected seq 0, got %d", seq)
	}

	seq, existing = st.StringToSequence("hat", true)
	if existing {
		t.Errorf("expected not to find entry")
	}
	if seq != 1 {
		t.Errorf("expected seq 1, got %d", seq)
	}

	// Can find existing entry if not asked to add new
	seq, existing = st.StringToSequence("hat", false)
	if !existing {
		t.Errorf("expected to find entry")
	}
	if seq != 1 {
		t.Errorf("expected seq 1, got %d", seq)
	}
}

func TestLowGC(t *testing.T) {
	st := New()
	defer st.Close()
	for i := range 10000000 {
		st.StringToSequence(strconv.Itoa(i), true)
	}
	runtime.GC()
	start := time.Now()
	runtime.GC()
	if time.Since(start) >= time.Millisecond*5 {
		t.Errorf("expected GC to take less than 5ms, took %s", time.Since(start))
	}

	runtime.KeepAlive(st)
}

func BenchmarkSymbolTab(b *testing.B) {
	symbols := make([]string, b.N)
	for i := range symbols {
		symbols[i] = strconv.Itoa(i)
	}

	b.ReportAllocs()
	b.ResetTimer()
	st := New()
	defer st.Close()
	for _, sym := range symbols {
		st.StringToSequence(sym, true)
	}

	if symbols[0] != st.SequenceToString(1) {
		b.Errorf("first symbol doesn't match - get %s", st.SequenceToString(1))
	}
}

func BenchmarkSymbolTabSmall(b *testing.B) {
	symbols := make([]string, 10_000)
	for i := range symbols {
		symbols[i] = strconv.Itoa(i)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		st := New()
		for _, sym := range symbols {
			st.StringToSequence(sym, true)
		}
		st.Close()
	}
}

func BenchmarkSequenceToString(b *testing.B) {
	st := New()
	defer st.Close()
	for i := range 100_000 {
		st.StringToSequence(strconv.Itoa(i), true)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		var str string
		for i := range 100_000 {
			str = st.SequenceToString(uint32(i + 1))
		}

		if str != strconv.Itoa(100_000-1) {
			b.Errorf("last symbol doesn't match - get %s", str)
		}
	}
}

func BenchmarkExisting(b *testing.B) {
	st := New()
	defer st.Close()
	values := make([]string, b.N)
	for i := range values {
		values[i] = strconv.Itoa(i)
	}

	for _, val := range values {
		st.StringToSequence(val, true)
	}

	b.ReportAllocs()
	b.ResetTimer()

	var seq uint32
	for _, val := range values {
		seq, _ = st.StringToSequence(val, false)
	}

	if st.SequenceToString(seq) != strconv.Itoa(b.N-1) {
		b.Errorf("last symbol doesn't match - get %s", st.SequenceToString(seq))
	}
}

// TODO: This is a bit useless as it tests a miss with an empty table only
func BenchmarkMiss(b *testing.B) {
	st := New()
	defer st.Close()
	values := make([]string, b.N)
	for i := range values {
		values[i] = strconv.Itoa(i)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for _, val := range values {
		_, found := st.StringToSequence(val, false)
		if found {
			b.Errorf("found value %s", val)
		}
	}
}

func ExampleSymbolTab() {
	st := New()
	defer st.Close()
	seq, found := st.StringToSequence("10293-ahdb-28383-555", true)
	fmt.Println(found)
	fmt.Println(st.SequenceToString(seq))
	// Output: false
	// 10293-ahdb-28383-555
}
