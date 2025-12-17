package swisssymbols

import (
	"testing"
)

func TestIntbank(t *testing.T) {
	ib := intbank{}
	ib.save(1, 37)
	ib.save(2, 43)

	if v := ib.lookup(1); v != 37 {
		t.Fatalf("expected 37, got %d", v)
	}
	if v := ib.lookup(2); v != 43 {
		t.Fatalf("expected 43, got %d", v)
	}

	if v := ib.lookup(1); v != 37 {
		t.Fatalf("expected 37, got %d", v)
	}
}
