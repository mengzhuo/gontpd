package gontpd

import (
	"context"
	"fmt"
	"math/bits"
	"time"

	"github.com/beevik/ntp"
)

const (
	ReplyCount           = 8
	MaxRootDelay         = 10 * time.Second
	MaxStratum     uint8 = 15
	minFilterCount       = 3
	maxInterval          = 1024 * time.Second
	minInterval          = 4 * time.Second
	reachStable    uint8 = 1<<8 - 1
	phi                  = 15 * time.Microsecond
)

const (
	stateInit = iota
	stateInvalid
	stateTemporaryDown
	stateSyncing
	stateStable
)

const (
	NoLeap uint8 = iota
	LeapIns
	LeapDel
	NotSync
)

const (
	queryMinInterval = 5 * time.Second
	queryMaxInterval = (1 << 12) * time.Minute
)

var (
	peerTransTable []peerTransitionFunc
	epoch          = time.Unix(0, 0)
)

type peer struct {
	addr       string
	filter     [ReplyCount]*filter
	referTime  time.Time
	updateAt   time.Time
	interval   time.Duration
	offset     time.Duration
	delay      time.Duration
	rootDelay  time.Duration
	rootDisp   time.Duration
	dispersion time.Duration

	filterIndex int
	best        int
	queryCount  int
	referId     uint32
	poll        int8
	leap        uint8
	reach       uint8
	stratum     uint8
	state       uint8
}

type filter struct {
	updateAt   time.Time
	offset     time.Duration
	delay      time.Duration
	dispersion time.Duration
	resp       *ntp.Response
}

func peerInit(p *peer, resp *ntp.Response, err error) {
	if p.queryCount < minFilterCount {
		// keep init
		Debug.Printf("query Count:%d keep initing", p.queryCount)
		return
	}

	if p.reach > 1 {
		p.state = stateSyncing
		Info.Printf("%s init -> syncing", p.addr)
		return
	}

	Warn.Printf("peer:%s has no good query while initing", p.addr)
	p.state = stateInvalid
	Info.Printf("%s init -> invalid", p.addr)
}

func peerInvalid(p *peer, resp *ntp.Response, err error) {
	Error.Fatal("we should not go into here")
}

func peerTemporaryDown(p *peer, resp *ntp.Response, err error) {
	Info.Printf("%s temporary down", p.addr)
	if p.reach&3 != 0 {
		p.state = stateSyncing
		Info.Printf("%s down -> syncing", p.addr)
		return
	}

	p.interval = time.Duration(8-bits.OnesCount8(p.reach)) * time.Minute
}

func peerSyncing(p *peer, resp *ntp.Response, err error) {
	if p.reach&3 == 0 {
		p.state = stateTemporaryDown
		Info.Printf("%s syncing -> temporary down", p.addr)
		p.interval = minInterval
		return
	}

	if p.reach == reachStable {
		Info.Printf("%s syncing -> stable", p.addr)
		p.state = stateStable
	}

	if p.checkFilter() {
		syncClock(p)
	}
	p.interval = time.Duration(bits.OnesCount8(p.reach)) * (10 * time.Second)
}
func peerStable(p *peer, resp *ntp.Response, err error) {

	if err != nil || p.reach != reachStable {
		p.state = stateSyncing
		Info.Printf("%s stable -> syncing", p.addr)
		return
	}

	if p.checkFilter() {
		syncClock(p)
	}
	// offset     interval
	// 0.128s  -> 1024s
	// 0.005s  -> 4s
	p.interval = secondToDuration(absDuration(p.offset).Seconds()*-8992.68 + 1065.46)
}

func syncClock(p *peer) {
	if DebugFlag&NoSyncClock != 0 {
		Debug.Printf("syncClock to:%s, offset:%s", p.addr, p.offset)
		return
	}
	// find median offset from all peers
	select {
	case syncLock <- struct{}{}:
	default:
	}
}

func init() {
	peerTransTable = []peerTransitionFunc{
		stateInit:          peerInit,
		stateInvalid:       peerInvalid,
		stateTemporaryDown: peerTemporaryDown,
		stateSyncing:       peerSyncing,
		stateStable:        peerStable,
	}
}

type peerTransitionFunc func(*peer, *ntp.Response, error)

func newPeer(addr string) *peer {
	p := &peer{
		addr:       addr,
		interval:   queryMinInterval,
		best:       -1,
		offset:     time.Duration(0),
		delay:      time.Minute,
		dispersion: time.Minute,
	}
	for i := 0; i < ReplyCount; i++ {
		p.filter[i] = &filter{
			updateAt:   epoch,
			dispersion: MaxRootDelay,
			delay:      MaxRootDelay,
			offset:     MaxRootDelay,
		}
	}
	return p
}

func (p *peer) insertFilter(resp *ntp.Response) {
	// the fisrt filter is the newest
	for i := ReplyCount - 1; i > 0; i-- {
		if p.filter[i].updateAt.Equal(epoch) {
			continue
		}
		interval := time.Now().Sub(p.filter[i].updateAt)
		p.filter[i] = p.filter[i-1]
		p.filter[i].updateAt = time.Now()
		p.filter[i].dispersion += phi * interval // add up dispersion
	}
	p.filter[0].delay = resp.RTT + resp.RootDelay
	p.filter[0].offset = resp.ClockOffset
	p.filter[0].updateAt = time.Now()
	p.filter[0].dispersion = resp.RootDispersion + p.filter[0].delay/2
	p.filter[0].resp = resp
}

func (p *peer) run(pctx context.Context) {
	ctx, cancel := context.WithCancel(pctx)
	defer cancel()

	Info.Printf("%s started %s", p.addr, p.interval)
	defer Info.Printf("%s stopped", p.addr)

	var timer *time.Timer

	for {
		resp, err := ntp.Query(p.addr)
		p.reach <<= 1
		p.queryCount++
		if err == nil && resp != nil && resp.Validate() == nil {
			p.reach |= 1
			p.insertFilter(resp)
		} else {
			Warn.Print(err)
			if resp != nil {
				Warn.Print(resp.Validate(), resp)
			}
		}

		peerTransTable[p.state](p, resp, err)

		if p.state == stateInvalid {
			Warn.Printf("peer %s invalid, exited looping", p.addr)
			return
		}
		if p.interval > maxInterval {
			p.interval = maxInterval
		}
		if p.interval < minInterval {
			p.interval = minInterval
		}
		p.interval += randDuration()
		Debug.Printf("%s will sleep %s", p.addr, p.interval)
		timer = time.NewTimer(p.interval)
		select {
		case <-timer.C:
		case <-ctx.Done():
			return
		}
		timer.Stop()
	}
}

func (p *peer) checkFilter() (valid bool) {

	bestIndex := -1
	bestDisp := time.Minute
	validCount := 0

	for i, r := range p.filter {
		if r == nil {
			continue
		}
		if r.dispersion >= MaxRootDelay {
			continue
		}
		validCount += 1
		if r.dispersion < bestDisp {
			bestIndex = i
		}
	}

	if validCount == 0 {
		valid = false
		return
	}

	p.best = bestIndex
	r := p.filter[bestIndex]
	p.dispersion = r.dispersion
	p.rootDelay = r.resp.RootDelay + r.resp.RTT
	p.delay = r.delay
	p.rootDisp = r.resp.RootDispersion
	p.offset = r.resp.ClockOffset
	p.interval = r.resp.Poll
	p.referTime = r.resp.ReferenceTime
	p.referId = r.resp.ReferenceID
	p.stratum = r.resp.Stratum
	p.updateAt = time.Now()
	Debug.Printf("%s choice new filter %d offset=%s", p.addr, bestIndex, p.offset)
	valid = true
	return
}

func (p *peer) String() string {
	return fmt.Sprintf("%s[%s]%s", p.addr, stateToString(p.state), p.offset)
}

func stateToString(u uint8) string {
	switch u {
	case stateInit:
		return "Init"
	case stateStable:
		return "stable"
	case stateInvalid:
		return "invalid"
	case stateTemporaryDown:
		return "down"
	case stateSyncing:
		return "syncing"
	}
	return "unknown"
}
