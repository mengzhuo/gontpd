package gontpd

import (
	"net"
	"sync"
	"time"
)

type secondLimitter struct {
	mu       sync.RWMutex
	interval int
	bucket   []map[uint32]struct{}
}

func (s *secondLimitter) allow(ip net.IP, ts time.Time) (a bool) {
	buck := s.bucket[ts.Second()/s.interval]
	var key uint32
	key = uint32(ip[0]<<24) + uint32(ip[1]<<16) + uint32(ip[2]<<8) + uint32(ip[3])
	_, in := buck[key]
	if !in {
		buck[key] = struct{}{}
	}
	return !in
}

func (s *secondLimitter) step(ts time.Time) {
	key := ts.Second() / s.interval
	l := len(s.bucket[key])
	prev := key - 1
	if prev < 0 {
		prev = len(s.bucket) - 1
	}
	s.bucket[prev] = make(map[uint32]struct{}, l/2)
}

func newLimitter(interval int) *secondLimitter {
	m := make([]map[uint32]struct{}, 60/interval)
	for i := 0; i < len(m); i++ {
		m[i] = map[uint32]struct{}{}
	}
	return &secondLimitter{
		mu:       sync.RWMutex{},
		interval: interval,
		bucket:   m,
	}
}
