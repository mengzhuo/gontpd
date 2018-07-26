package gontpd

import "testing"

func TestLRUAdd(t *testing.T) {
	l := newLRU(3)
	l.Add("ok", 1)
	if z, ok := l.Get("ok"); z != 1 || !ok {
		t.Error(z, ok)
	}
	l.Add("ok", 2)
	if z, ok := l.Get("ok"); z != 2 || !ok {
		t.Error(z, ok)
	}
}
