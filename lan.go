package gontpd

import (
	"log"
	"net"
)

var (
	lan192 *net.IPNet
	lan172 *net.IPNet
	lan10  *net.IPNet
)

func init() {
	var err error
	_, lan192, err = net.ParseCIDR("192.168.0.0/16")
	if err != nil {
		log.Fatal(err)
	}

	_, lan172, err = net.ParseCIDR("172.16.0.0/12")
	if err != nil {
		log.Fatal(err)
	}

	_, lan10, err = net.ParseCIDR("10.0.0.0/8")
	if err != nil {
		log.Fatal(err)
	}

}

func isLan(ip net.IP) (in bool) {
	in = lan10.Contains(ip)
	if in {
		return
	}
	in = lan172.Contains(ip)
	if in {
		return
	}
	in = lan192.Contains(ip)
	if in {
		return
	}
	return false
}
