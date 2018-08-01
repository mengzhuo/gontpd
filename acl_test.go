package gontpd

import (
	"net"
	"testing"
)

func TestSubNet(t *testing.T) {
	gold := []struct {
		cidr       string
		start, end string
	}{
		{"127.0.0.1/32", "127.0.0.1", "127.0.0.1"},
		{"172.16.0.0/12", "172.16.0.0", "172.31.255.255"},
		{"10.0.0.0/8", "10.0.0.0", "10.255.255.255"},
		{"45.40.192.0/19", "45.40.192.0", "45.40.223.255"},
		{"0::0/0", "::", "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"},
		{"2001:df6:f400::/48", "2001:df6:f400::", "2001:df6:f400:ffff:ffff:ffff:ffff:ffff"},
	}

	for _, g := range gold {
		start := net.ParseIP(g.start)
		end := net.ParseIP(g.end)
		_, ipnet, err := net.ParseCIDR(g.cidr)
		if err != nil {
			t.Error(err)
		}
		gotS, gotE := subNet(ipnet)
		if !gotS.Equal(start) {
			t.Errorf("start=%s CIDR=%s got=%s", g.start, g.cidr, gotS)
		}
		if !gotE.Equal(end) {
			t.Errorf("end=%s CIDR=%s got=%s", g.end, g.cidr, gotE)
		}
	}
}

func TestDropTableMulti(t *testing.T) {
	dt, err := newDropTable([]string{
		"127.0.0.1/8", "172.16.0.0/12", "10.0.0.0/8", "192.168.0.0/16",
	})
	if err != nil {
		t.Fatal(err)
	}
	gold := []struct {
		ip string
		in bool
	}{
		{"127.0.0.1", true},
		{"172.16.3.1", true},
		{"10.44.22.196", true},
		{"9.44.22.196", false},
		{"1.1.1.1", false},
		{"255.255.1.1", false},
	}

	for _, g := range gold {
		ip := net.ParseIP(g.ip)
		if dt.contains(ip) != g.in {
			t.Errorf("table contains:%s expect:%v got:%v", ip, g.in, !g.in)
			t.Error(dt)
		}
	}
}

func BenchmarkDropTableMulti(b *testing.B) {

	dt, err := newDropTable([]string{
		"127.0.0.1/8", "172.16.0.0/12", "10.0.0.0/8", "192.168.0.0/16",
	})
	if err != nil {
		b.Fatal(err)
	}
	ip := net.IP{5, 6, 7, 8}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dt.contains(ip)
	}
}
