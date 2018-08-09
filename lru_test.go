package gontpd

import (
	"encoding/binary"
	"net"
	"testing"
)

func TestLRUAddSame(t *testing.T) {
	ip := net.IP{1, 2, 3, 4}
	l := newLRU(3)
	for i := int64(0); i < 10; i++ {
		l.Add(ip, i)
		if z, ok := l.Get(ip); z != i || !ok || len(l.cache) != 1 {
			t.Error(z, ok)
		}
	}
}

func TestLRUAddFull(t *testing.T) {
	u := newLRU(10)
	for i := 0; i < 20; i++ {
		u.Add(net.IP{0, 0, 0, byte(i)}, int64(i))
	}

	next := u.root.next
	ips := []net.IP{}
	tt := []int64{}
	if len(u.cache) != 10 {
		t.Error("cache overflow", len(u.cache))
		for k, v := range u.cache {
			t.Error(k, v.lastUnix)
		}
	}

	// check order
	for i := 0; i < u.maxEntry; i++ {
		tt = append(tt, next.lastUnix)
		ips = append(ips, next.key)
		next = next.next
	}
	prev := int64(20)
	for i, j := range tt {
		if prev-j != 1 {
			t.Error(tt)
		}
		prev = j
		if ips[i][3] != byte(j) {
			t.Error(ips)
		}
	}
}

func BenchmarkLRUGet(b *testing.B) {
	ip := net.IP{1, 2, 3, 4}
	l := newLRU(3)
	l.Add(ip, 1)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Get(ip)
	}
}

func BenchmarkLRUAddFull(b *testing.B) {
	l := newLRU(100)
	ipList := []net.IP{}
	for i := 0; i < 1000; i++ {
		nip := net.IP{0, 0, 0, 0}
		binary.BigEndian.PutUint32([]byte(nip), uint32(i))
		ipList = append(ipList, nip)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Add(ipList[i%1000], 0)
	}
}

func BenchmarkLRUAddExisted(b *testing.B) {
	ip := net.IP{0, 0, 0, 0}
	l := newLRU(3)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Add(ip, 0)
	}
}
