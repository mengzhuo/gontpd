package gontpd

import (
	"net"
	"testing"
)

func TestLRUAdd(t *testing.T) {
	ip := net.IP{1, 2, 3, 4}
	l := newLRU(3)
	l.Add(ip, 1)
	if z, ok := l.Get(ip); z != 1 || !ok {
		t.Error(z, ok)
	}
	l.Add(ip, 2)
	if z, ok := l.Get(ip); z != 2 || !ok {
		t.Error(z, ok)
	}
}
