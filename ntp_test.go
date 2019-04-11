package gontpd

import (
	"testing"
	"time"
)

func TestToNTPTime(t *testing.T) {
	ts := time.Now()
	a := naiveNtpTime(ts)
	b := toNTPTime(ts)
	if a != b {
		t.Errorf("ntp time failed %d != %d", a, b)
	}
}

var ntpEpoch = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)

func naiveNtpTime(t time.Time) uint64 {
	nsec := uint64(t.Sub(ntpEpoch))
	sec := nsec / nanoPerSec
	frac := (nsec - sec*nanoPerSec) << 32 / nanoPerSec
	return uint64(sec<<32 | frac)
}

func BenchmarkNTPTime(b *testing.B) {
	g := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		toNTPTime(g)
	}
}
