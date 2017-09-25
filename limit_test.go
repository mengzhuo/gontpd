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

func TestSecondLimiter(t *testing.T) {
	s := newLimitter(4)
	a := randIP()
	ts := time.Now()
	if !s.allow(a, ts) {
		t.Error(a, "in limtter")
	}
	if s.allow(a, ts) {
		t.Error(a, "not in limtter")
	}
	at := ts.Add(time.Second)
	s.step(at)

	if s.allow(a, at) {
		t.Error(a, "in limtter")
	}
	at = ts.Add(3 * time.Second)
	s.step(at)
	if !s.allow(a, at) {
		t.Error(a, "not in limtter")
	}
}

func TestFullCheck(t *testing.T) {

	s := newLimitter(4)
	a := randIP()
	ts := time.Now()

	for i := 0; i < 120; i += 4 {
		ts = ts.Add(4 * time.Second)
		if s.allow(a, ts) {
			t.Error(a, ts, "in limtter")
		}
		s.step(ts)
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
