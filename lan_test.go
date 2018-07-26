package gontpd

import (
	"net"
	"testing"
)

func TestIsLan(t *testing.T) {
	tt := []struct {
		ip string
		in bool
	}{
		{"192.168.1.1", true},
		{"172.16.66.77", true},
		{"10.0.0.1", true},
		{"255.255.255.255", false},
		{"1.1.1.1", false},
		{"74.120.168.214", false},
		{"45.76.218.213", false},
	}
	for _, g := range tt {
		ip := net.ParseIP(g.ip)
		if got := isLan(ip); got != g.in {
			t.Errorf(" %s expecting=%v got=%v", ip, g.in, got)
		}
	}
}
