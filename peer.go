package gontpd

import (
	"crypto/md5"
	"log"
	"net"
	"time"

	"github.com/beevik/ntp"
)

const (
	replyNum       = 4
	goodFilter     = 2
	invalidStratum = 16
	maxPoll        = 16
	minPoll        = 5
)

type peer struct {
	addr       net.IP
	reply      [replyNum]*ntp.Response
	offset     time.Duration
	delay      time.Duration
	err        time.Duration
	refId      uint32
	stratum    uint8
	trustLevel uint8
	good       bool
	enable     bool
}

func newPeer(addr net.IP) (p *peer) {
	log.Printf("new peer:%s", addr.String())
	p = &peer{
		addr:       addr,
		trustLevel: minPoll,
		enable:     true,
	}
	p.refId = makeSendRefId(addr)
	return
}

func (p *peer) update() {

	goodCount := 0
	p.good = false
	ts := 2 * time.Second

	for i := 0; i < replyNum; i++ {
		time.Sleep(ts)
		resp, err := ntp.Query(p.addr.String())
		if resp != nil && resp.Stratum == 0 {
			switch resp.KissCode {
			case "RATE":
				ts += time.Second
			case "DENY":
				p.enable = false
				return
			}
		}

		if err != nil {
			log.Printf("%s update failed %s", p.addr.String(), err)
			p.reply[i] = &ntp.Response{Stratum: invalidStratum}
			if nerr, ok := err.(net.Error); ok {
				if !nerr.Temporary() {
					log.Printf("%s can't be reach, disabled", p.addr.String())
					p.enable = false
					return
				}
			}
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

func makeSendRefId(ip net.IP) (id uint32) {

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
