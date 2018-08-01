package gontpd

import (
	"encoding/binary"
	"fmt"
	"net"
	"sort"
	"strings"
)

const (
	uint64Full = uint64(0xffffffffffffffff)
	uint32Full = uint32(0xffffffff)
)

func subNet(n *net.IPNet) (start, end net.IP) {
	start = n.IP
	ones, bits := n.Mask.Size()
	if ones == bits {
		end = n.IP
		return
	}

	switch bits {
	case net.IPv4len * 8:
		si := binary.BigEndian.Uint32([]byte(start))
		mask := uint32Full >> uint(ones)
		si ^= mask
		end = net.IP{0, 0, 0, 0}
		binary.BigEndian.PutUint32([]byte(end), si)
	case net.IPv6len * 8:
		sl := binary.BigEndian.Uint64([]byte(start)[:8])
		sh := binary.BigEndian.Uint64([]byte(start)[8:])

		ml := uint64Full >> uint(ones)
		mh := uint64Full
		if ones > 64 {
			mh >>= uint(ones - 64)
		}

		sl ^= ml
		sh ^= mh

		end = net.IP{
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0}
		binary.BigEndian.PutUint64([]byte(end)[:8], sl)
		binary.BigEndian.PutUint64([]byte(end)[8:], sh)
	}
	return
}

type dropTable struct {
	CIDR  []string
	slice []*cidrItem
	net   *net.IPNet
}

func newDropTable(cidr []string) (d *dropTable, err error) {
	d = &dropTable{CIDR: cidr}
	if len(cidr) == 1 {
		_, d.net, err = net.ParseCIDR(cidr[0])
	}

	if len(cidr) > 1 {
		nl := []*net.IPNet{}
		for _, c := range cidr {
			var n *net.IPNet
			_, n, err = net.ParseCIDR(c)
			if err != nil {
				return
			}
			nl = append(nl, n)
		}
		d.slice, err = newStaticSegTree(nl)
	}

	return
}

func ipToUint32(ip net.IP) uint32 {
	if len(ip) == net.IPv4len {
		return binary.BigEndian.Uint32([]byte(ip))
	}
	return binary.BigEndian.Uint32([]byte(ip)[12:])
}

func ipToUInt128(ip net.IP) (l, h uint64) {
	l = binary.BigEndian.Uint64([]byte(ip)[8:])
	h = binary.BigEndian.Uint64([]byte(ip)[:8])
	return
}

func (d *dropTable) snlContains(ip net.IP) bool {
	s := d.slice
	if ip.To4() != nil {
		ipN := ipToUint32(ip)
		for len(s) > 2 {
			mid := len(s) / 2
			if s[mid].mid4 < ipN {
				s = s[mid:]
			} else {
				s = s[:mid+1]
			}
		}

	} else {
		ipL, ipH := ipToUInt128(ip)
		for len(s) > 2 {
			mid := len(s) / 2
			sm := s[mid]
			if sm.mid6H < ipH {
				s = s[mid:]
				continue
			}
			if sm.mid6H > ipH {
				s = s[:mid+1]
				continue
			}

			if sm.mid6L < ipL {
				s = s[mid:]
			} else {
				s = s[:mid+1]
			}
		}
	}

	if s[0].ipnet.Contains(ip) {
		return true
	}

	if s[1].ipnet.Contains(ip) {
		return true
	}
	return false
}

func (d *dropTable) String() string {
	var w strings.Builder
	for _, r := range d.slice {
		fmt.Fprintf(&w, "%s\n", r.ipnet)
	}
	return w.String()
}

func (d *dropTable) contains(ip net.IP) bool {
	switch len(d.CIDR) {
	case 0:
		return false
	case 1:
		return d.net.Contains(ip)
	default:
		return d.snlContains(ip)
	}
}

func newStaticSegTree(nl []*net.IPNet) (snl []*cidrItem, err error) {

	for _, n := range nl {
		cidrItem := newSegNode(n)
		snl = append(snl, cidrItem)
	}
	sort.Sort(byIP(snl))

	for i := 0; i < len(snl)-1; i++ {
		if ipLess(snl[i+1].left, snl[i].right) {
			err = fmt.Errorf("cidr overlaped, snl[i]=%s snl[i+1]=%s",
				snl[i].right, snl[i].left,
			)
			return
		}
	}

	return
}

func ipLess(a, b net.IP) bool {
	if a.To4() != nil {
		return ipToUint32(a) < ipToUint32(b)
	}

	al, ah := ipToUInt128(a)
	bl, bh := ipToUInt128(b)
	if ah != bh {
		return ah < bh
	}
	return al < bl
}

type cidrItem struct {
	ipnet        *net.IPNet
	left, right  net.IP
	mid6L, mid6H uint64
	mid4         uint32
}

func newSegNode(ipnet *net.IPNet) (n *cidrItem) {

	l, r := subNet(ipnet)
	n = &cidrItem{ipnet: ipnet, left: l, right: r}
	if l.To4() != nil {
		low := ipToUint32(l)
		n.mid4 = (ipToUint32(r)-low)/2 + low
		return
	}
	ln := binary.BigEndian.Uint64([]byte(l)[:8])
	rn := binary.BigEndian.Uint64([]byte(r)[:8])
	carry := rn < ln
	if carry {
		n.mid6L = (uint64Full-ln+1+rn)/2 + ln
	} else {
		n.mid6L = (rn-ln)/2 + ln
	}

	ln = binary.BigEndian.Uint64([]byte(l)[8:])
	rn = binary.BigEndian.Uint64([]byte(r)[8:])
	if carry {
		rn -= 1
	}
	n.mid6H = (rn-ln)/2 + ln
	return
}

type byIP []*cidrItem

func (b byIP) Less(i, j int) bool {
	return ipLess(b[i].left, b[j].left)
}

func (b byIP) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b byIP) Len() (i int) {
	return len(b)
}
