package gontpd

import (
	"testing"
	"time"
)

func TestToNtpShortTime(t *testing.T) {
	a := toNtpShortTime(time.Second)
	if a != 65536 {
		t.Fatal(a)
	}
}

func TestIntSqrt(t *testing.T) {
	gold := []struct{ v, g uint64 }{
		{0, 0},
		{1, 1},
		{2, 1},
		{3, 1},
		{4, 2},
		{9, 3},
		{16, 4},
		{25, 5},
		{26, 5},
		{35, 5},
		{36, 6},
	}

	for _, p := range gold {
		got := uintSqrt(p.v)
		if got != p.g {
			t.Errorf("%d sqrt = %d expecting=%d", p.v, got, p.g)
		}
	}
}

func TestStddev(t *testing.T) {
	gold := []struct {
		g time.Duration
		v []time.Duration
	}{
		{time.Millisecond,
			[]time.Duration{49 * time.Millisecond, 50 * time.Millisecond, 51 * time.Millisecond}},
	}
	for _, p := range gold {
		got := stddev(p.v)
		if got > p.g {
			t.Errorf("%s stddev = %s expecting=%s", p.v, got, p.g)
		}
	}
}
