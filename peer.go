package gontpd

import (
	"crypto/md5"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/beevik/ntp"
)

const (
	offsetSize = 8

	trustlevelBadpeer    = 6
	trustlevelPathetic   = 2
	trustlevelAggressive = 8
	trustlevelMax        = 10
)

const (
	NoLeap uint8 = iota
	LeapIns
	LeapDel
	NotSync
)

const (
	intervalQueryNormal     = 30 * time.Second /* sync to peers every n secs */
	intervalQueryPathetic   = 60 * time.Second
	intervalQueryAggressive = 5 * time.Second
	qScaleOffMax            = 50 * time.Millisecond
	qScaleOffMin            = time.Millisecond

	frequencySamples           = 8
	maxFrequencyAdjust float64 = 128e-5
	maxStratum                 = 16
	filterAdjFreq              = 0x01
)

const (
	stateNone peerState = iota
	stateNetworkTempfail
	stateQuerySent
	stateReplyReceived
	stateTimeout
	stateInvalid
)

type peerState uint8

type ntpStatus struct {
	rootDelay      time.Duration
	rootDispersion time.Duration
	refTime        time.Time
	refId          uint32
	sendRefId      uint32
	synced         bool
	leap           uint8
	percision      int8
	poll           uint8
	stratum        uint8
}

type ntpOffset struct {
	status ntpStatus
	offset time.Duration
	delay  time.Duration
	err    time.Duration
	// received
	rcvd time.Time
	good bool
}

var (
	epoch = time.Unix(0, 0)
)

type peer struct {
	addr       string
	reply      [offsetSize]*ntpOffset
	update     *ntpOffset
	next       time.Duration
	deadline   time.Time
	poll       time.Time
	lastErrors int
	sendErrors int
	id         uint32
	shift      uint8
	trustLevel uint8
	state      peerState
	sync.Mutex
}

func newPeer(addr string) *peer {
	p := &peer{
		addr:       addr,
		trustLevel: trustlevelPathetic,
		state:      stateNone,
	}
	for i := 0; i < offsetSize; i++ {
		p.reply[i] = &ntpOffset{}
	}
	return p
}

func (s *Service) run(p *peer) {

	for {
		resp, err := p.query()
		if err == nil && resp != nil {
			s.dispatch(p, resp)
		}
		time.Sleep(p.next)
	}

	return
}

func (p *peer) query() (resp *ntp.Response, err error) {
	p.Lock()
	defer p.Unlock()

	resp, err = ntp.Query(p.addr)
	if err != nil {
		Warn.Print(err)
		p.state = stateNetworkTempfail
		p.setNext(intervalQueryPathetic)
		return
	}
	if debug {
		log.Printf("%s -> %v, err:%s", p.addr, resp, err)
	}
	return
}

func (s *Service) dispatch(p *peer, resp *ntp.Response) {
	p.Lock()
	defer p.Unlock()

	if resp.Validate() != nil {
		p.next = errorInterval()
		return
	}

	// TODO: detect liars

	p.reply[p.shift].offset = resp.ClockOffset
	p.reply[p.shift].delay = resp.RTT
	p.reply[p.shift].status.stratum = resp.Stratum
	if resp.RTT < 0 {
		Warn.Printf("%s got neg rtt:%s", p.addr, resp.RTT)
		p.next = errorInterval()
		return
	}

	p.reply[p.shift].err = resp.MinError
	p.reply[p.shift].rcvd = time.Now()
	p.reply[p.shift].good = true

	p.reply[p.shift].status.leap = uint8(resp.Leap)
	p.reply[p.shift].status.percision = reverseToInterval(resp.Precision)
	p.reply[p.shift].status.rootDelay = resp.RootDelay
	p.reply[p.shift].status.rootDispersion = resp.RootDispersion
	p.reply[p.shift].status.refId = resp.ReferenceID
	p.reply[p.shift].status.refTime = resp.ReferenceTime
	p.reply[p.shift].status.poll = uint8(durationToPoll(resp.Poll))
	p.reply[p.shift].status.sendRefId = p.makeSendRefId()

	interval := intervalQueryPathetic
	if p.trustLevel < trustlevelPathetic {
		interval = s.scaleInterval(intervalQueryPathetic)
	} else if p.trustLevel < trustlevelAggressive {
		interval = s.scaleInterval(intervalQueryAggressive)
	} else {
		interval = s.scaleInterval(intervalQueryNormal)
	}

	p.setNext(interval)

	if p.trustLevel < trustlevelMax {
		if p.trustLevel < trustlevelBadpeer &&
			p.trustLevel+1 >= trustlevelBadpeer {
			Info.Printf("peer %s now valid", p.addr)
		}
		p.trustLevel++
	}

	if debug {
		log.Printf("reply from:%s, offset:%s delay:%s",
			p.addr,
			p.reply[p.shift].offset,
			p.reply[p.shift].delay,
		)
		log.Printf("%s will query at %s", p.addr, interval)
	}

	s.clockFilter(p)
}

func (p *peer) makeSendRefId() (id uint32) {

	addrs, err := net.LookupHost(p.addr)
	if err != nil {
		Warn.Print(err)
		return
	}

	if len(addrs) == 0 {
		return
	}

	ip := net.ParseIP(addrs[0])

	if ip[11] == 255 {
		// ipv4
		id = uint32(ip[12])<<24 + uint32(ip[13])<<16 + uint32(ip[14])<<8 + uint32(ip[15])
	} else {
		h := md5.New()
		hr := h.Sum(ip)
		// 255.b2.b3.b4 for ipv6 hash
		// https://support.ntp.org/bin/view/Dev/UpdatingTheRefidFormat
		id = uint32(255)<<24 + uint32(hr[1])<<16 + uint32(hr[2])<<8 + uint32(hr[3])
	}
	return
}

func (s *Service) clockFilter(p *peer) (err error) {
	/*
	 * find the offset which arrived with the lowest delay
	 * use that as the peer update
	 * invalidate it and all older ones
	 */
	var best, good int

	for i, r := range p.reply {
		if r.good {
			good++
			best = i
		}
	}

	for i := best; i < len(p.reply); i++ {
		if p.reply[i].good {
			good++
			if p.reply[i].delay < p.reply[best].delay {
				best = i
			}
		}
	}

	if good < 8 {
		return fmt.Errorf("peer:%s not good enough:%d", p.addr, good)
	}

	*p.update = *p.reply[best]

	if s.privAjdtime() == nil {
		for i, r := range p.reply {
			if !r.rcvd.After(p.reply[best].rcvd) {
				p.reply[i].good = false
			}
		}
	}
	p.shift++
	if p.shift >= offsetSize {
		p.shift = 0
	}

	return
}

func (s *Service) privAdjFreq(offset time.Duration) {
	var currentTime, freq float64

	if !s.status.synced {
		s.freq.samples = 0
		return
	}

	s.freq.samples++

	if s.freq.samples <= 0 {
		return
	}

	s.freq.overallOffset += offset
	offset = s.freq.overallOffset

	currentTime = gettimeCorrected()

	s.freq.xy += offset.Seconds() * currentTime
	s.freq.x += currentTime
	s.freq.y += offset.Seconds()
	s.freq.xx += currentTime * currentTime

	if s.freq.samples%frequencySamples != 0 {
		return
	}

	freq = (s.freq.xy - s.freq.x*s.freq.y/float64(s.freq.samples)) /
		(s.freq.xx - s.freq.x*s.freq.x/float64(s.freq.samples))

	if freq > maxFrequencyAdjust {
		freq = maxFrequencyAdjust
	} else if freq < -maxFrequencyAdjust {
		freq = -maxFrequencyAdjust
	}

	s.filters |= filterAdjFreq
	s.freq.xy = 0
	s.freq.x = 0
	s.freq.y = 0
	s.freq.xx = 0
	s.freq.samples = 0
	s.freq.overallOffset = 0
	s.freq.num++
}

func (s *Service) privAjdtime() (err error) {
	offsets := []*ntpOffset{}
	for _, p := range s.peerList {
		if !p.update.good {
			continue
		}
		offsets = append(offsets, p.update)
	}

	sort.Sort(byOffset(offsets))

	i := len(offsets) / 2
	if len(offsets)%2 == 0 {
		if offsets[i-1].delay < offsets[i].delay {
			i -= 1
		}
	}

	offsetMedian := offsets[i].offset
	s.status.rootDelay = offsets[i].delay
	s.status.stratum = offsets[i].status.stratum
	s.status.leap = offsets[i].status.leap

	s.privAdjFreq(offsetMedian)

	s.status.refTime = time.Now()
	s.status.stratum++
	if s.status.stratum > maxStratum {
		s.status.stratum = maxStratum
	}

	s.updateScale(offsetMedian)
	s.status.refId = offsets[i].status.sendRefId
	s.setTemplate(offsets[i])
	for _, p := range s.peerList {
		for j := 0; j < len(p.reply); j++ {
			p.reply[j].offset -= offsetMedian
		}
		p.update.good = false
	}
	return
}

func (s *Service) updateScale(offset time.Duration) {
	offset += getOffset()
	if offset < 0 {
		offset = -offset
	}

	if offset > qScaleOffMax || !s.status.synced || s.freq.num < 3 {
		s.scale = time.Duration(1)
	} else if offset < qScaleOffMin {
		s.scale = qScaleOffMax / qScaleOffMin
	} else {
		s.scale = qScaleOffMax / offset
	}
}

func (p *peer) String() string {
	return fmt.Sprintf("%s [%d]", p.addr, p.state)
}

func (p *peer) setNext(d time.Duration) {
	p.next = d + time.Duration(rand.Int63n(int64(d)/10))
}

func (s *Service) scaleInterval(d time.Duration) (sd time.Duration) {
	sd = s.scale * d
	r := maxDuration(5*time.Second, sd/10)
	return sd + r
}

func errorInterval() time.Duration {
	return time.Duration(rand.Int63n(int64(intervalQueryPathetic*qScaleOffMax/qScaleOffMin)) / 10)
}

type byOffset []*ntpOffset

func (ol byOffset) Len() int {
	return len(ol)
}

func (ol byOffset) Swap(i, j int) {
	ol[i], ol[j] = ol[j], ol[i]
}

func (ol byOffset) Less(i, j int) bool {
	return ol[i].offset < ol[j].offset
}
