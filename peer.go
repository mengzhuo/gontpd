package gontpd

import (
	"crypto/md5"
	"log"
	"net"
	"time"

	"github.com/beevik/ntp"
)

const (
	replyNum       = 3
	goodFilter     = 2
	invalidStratum = 16
	maxPoll        = 16
	minPoll        = 5
)

type peer struct {
	addr       string
	reply      [replyNum]*ntp.Response
	offset     time.Duration
	delay      time.Duration
	err        time.Duration
	refId      uint32
	stratum    uint8
	trustLevel uint8
	good       bool
}

func newPeer(addr string) (p *peer) {
	log.Printf("new peer:%s", addr)
	p = &peer{
		addr:       addr,
		trustLevel: minPoll,
	}
	p.refId = makeSendRefId(addr)
	return
}

func (p *peer) update() {

	goodCount := 0

	for i := 0; i < replyNum; i++ {
		time.Sleep(2 * time.Second)
		resp, err := ntp.Query(p.addr)
		if err != nil {
			log.Printf("%s update failed %s", p.addr, err)
			p.reply[i] = &ntp.Response{Stratum: invalidStratum}
			continue
		}
		goodCount += 1
		p.reply[i] = resp
	}

	p.good = goodCount > goodFilter

	if debug {
		log.Printf("%s is good=%v", p.addr, p.good)
	}

	if p.good {
		if p.trustLevel < maxPoll {
			p.trustLevel += 1
		}
	} else {
		if p.trustLevel > minPoll {
			p.trustLevel -= 1
		}
	}
}

func makeSendRefId(addr string) (id uint32) {

	ips, err := net.LookupIP(addr)
	if err != nil || len(ips) == 0 {
		log.Print(err)
		return 0
	}
	ip := ips[0]

	if len(ip) > 10 && ip[11] == 255 {
		// ipv4
		id = uint32(ip[12])<<24 + uint32(ip[13])<<16 + uint32(ip[14])<<
			8 + uint32(ip[15])
	} else {
		h := md5.New()
		hr := h.Sum(ip)
		// 255.b2.b3.b4 for ipv6 hash
		// https://support.ntp.org/bin/view/Dev/UpdatingTheRefidFormat
		id = uint32(255)<<24 + uint32(hr[1])<<16 + uint32(hr[2])<<8 + uint32(hr[3])
	}
	return
}
