package gontpd

import (
	"math/rand"
	"net"
	"testing"
	"time"
)

func randIP() net.IP {
	a := make(net.IP, 4)
	rand.Read(a)
	return a
}

func TestMakeKey(t *testing.T) {
	a := make(net.IP, 4)
	a[0] = 0x32
	a[1] = 0x54
	a[2] = 0x18
	a[3] = 0x38
	if makeKey(a) != 0x32541838 {
		t.Error(makeKey(a), a)
	}
}

func TestSecondLimiter(t *testing.T) {
	s := newLimitter(4)
	a := randIP()
	ts := time.Unix(0, 0)
	if !s.allow(a, ts) {
		t.Error(a, "in limtter")
	}
	if s.allow(a, ts) {
		t.Error(a, "not in limtter")
	}
	at := ts.Add(time.Second)
	s.clear(at)

	if !s.allow(a, at) {
		t.Error(a, "in limtter", s.bucket)
	}
	if s.allow(a, at) {
		t.Error(a, "not in limtter")
	}
}

func TestFullCheck(t *testing.T) {

	s := newLimitter(4)
	a := randIP()
	ts := time.Unix(0, 0)
	s.allow(a, ts)
	for i := 0; i < 15; i += 4 {
		ts = ts.Add(time.Second * 4)
		if !s.allow(a, ts) {
			t.Error(a, ts, "in limtter")
		}
	}

}

func BenchmarkSecondLimiter(b *testing.B) {
	s := newLimitter(4)
	a := randIP()
	t := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.allow(a, t)
	}
}
