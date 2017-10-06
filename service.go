package gontpd

import (
	"context"
	"net"
	"sort"
	"time"
)

const (
	minStep   = 120 * time.Second
	minPoll   = 5
	maxPoll   = 8
	maxAdjust = 128 * time.Millisecond
	initRefer = 0x494e4954 // ascii for INIT
)

var (
	// the only data global variable
	syncLock = make(chan struct{})
)

func NewService(cfg *Config) (s *Service, err error) {
	addr, err := net.ResolveUDPAddr("udp", cfg.Listen)
	if err != nil {
		return nil, err
	}
	s = &Service{
		cfg: cfg,
	}
	s.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return
	}

	for _, paddr := range cfg.ServerList {
		p := newPeer(paddr)
		s.peerList = append(s.peerList, p)
	}
	s.template = newTemplate()

	return
}

type Service struct {
	peerList []*peer
	conn     *net.UDPConn
	stats    *statistic
	cfg      *Config
	template []byte
	interval time.Duration
	poll     int8
}

func (s *Service) serveSyncLock() (err error) {

	var (
		lock  uint8
		timer *time.Timer
	)
	s.interval = 30 * time.Second
	timer = time.NewTimer(s.interval)

	for {
		select {
		case <-syncLock:
			lock |= 1
		case <-timer.C:
			lock |= 2
		}
		if lock != 3 {
			continue
		}

		err = s.syncClock()
		if err != nil {
			return
		}
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(s.interval)
		lock = 0
	}
	return
}

func (s *Service) syncClock() (err error) {
	availablePeers := []*peer{}
	for _, p := range s.peerList {
		if p.state < stateSyncing {
			Debug.Printf("%s state not Syncing", p.addr)
			continue
		}
		if p.stratum >= 15 {
			Debug.Printf("%s stratum is bad", p.addr)
			continue
		}
		if p.delay >= MaxRootDelay {
			Debug.Printf("%s delay %s >= MaxRootDelay %s", p.addr, p.delay, MaxRootDelay)
			continue
		}
		availablePeers = append(availablePeers, p)
	}

	switch len(availablePeers) {
	case 0:
		Warn.Print("no availablePeers")
		return nil
	case 1:
		s.setParams(availablePeers[0])
		return s.setOffset(availablePeers[0])
	default:
		sort.Sort(byOffset(availablePeers))
		Debug.Printf("availablePeers %d", len(availablePeers))
		for _, p := range availablePeers {
			Debug.Printf(" |-- %s", p)
		}
		bestPeer := availablePeers[len(availablePeers)/2]
		s.setParams(bestPeer)
		return s.setOffset(bestPeer)
	}
}

func (s *Service) setParams(p *peer) {

	SetLi(s.template, p.leap)
	SetVersion(s.template, 4)
	SetMode(s.template, ModeServer)

	SetUint8(s.template, Stratum, p.stratum+1)

	SetInt8(s.template, ClockPrecision, systemPrecision())

	SetUint32(s.template, RootDelayPos, toNtpShortTime(p.delay))
	s.stats.delayGauge.Set(p.delay.Seconds())

	s.stats.offsetGauge.Set(p.offset.Seconds())
	s.stats.dispGauge.Set(p.dispersion.Seconds())

	SetUint32(s.template, RootDispersionPos, toNtpShortTime(p.dispersion))
	SetUint64(s.template, ReferenceTimeStamp, toNtpTime(p.updateAt))
	SetUint32(s.template, ReferIDPos, p.referId)

	switch p.state {
	case stateStable:
		s.interval = maxInterval
	case stateSyncing:
		s.interval = p.interval * 2
	default:
		s.interval = 30 * time.Second
	}
	s.poll = int8(durationToPoll(s.interval))
	Debug.Printf("try to set poll %d, by %s", s.poll, s.interval)
	if s.poll > maxPoll {
		s.poll = maxPoll
	}
	if s.poll < minPoll {
		s.poll = minPoll
	}
	SetInt8(s.template, Poll, s.poll)
}

type byOffset []*peer

func (b byOffset) Len() int {
	return len(b)
}

func (b byOffset) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b byOffset) Less(i, j int) bool {
	return b[i].offset < b[j].offset
}

func (s *Service) workerDo(i int) {
	var (
		n           int
		remoteAddr  *net.UDPAddr
		err         error
		receiveTime time.Time
		limiter     *secondLimitter
	)

	p := make([]byte, 48)
	if s.cfg.ReqRateSec > 0 {
		Info.Printf("limiter %d", s.cfg.ReqRateSec)
		limiter = newLimitter(s.cfg.ReqRateSec)
		go limiter.run()
	}

	defer func(i int) {
		if r := recover(); r != nil {
			Error.Printf("Worker: %d fatal, reason:%s, read:%d", i, r, n)
		} else {
			Info.Printf("Worker: %d exited, reason:%s, read:%d", i, err, n)
		}
	}(i)

	for {
		n, remoteAddr, err = s.conn.ReadFromUDP(p)
		if err != nil {
			return
		}

		receiveTime = time.Now()
		if n < 48 {
			Warn.Printf("worker: %s get small packet %d",
				remoteAddr.String(), n)
			continue
		}

		if limiter != nil {
			if !limiter.allow(remoteAddr.IP, receiveTime) {
				Warn.Printf("worker[%d]: get limitted ip %s",
					i, remoteAddr.String())
				continue
			}
		}

		// GetMode
		switch p[LiVnMode] &^ 0xf8 {
		case ModeReserved:
			fallthrough
		case ModeClient:
			copy(p[0:OriginTimeStamp], s.template)
			copy(p[OriginTimeStamp:OriginTimeStamp+8],
				p[TransmitTimeStamp:TransmitTimeStamp+8])
			SetUint64(p, ReceiveTimeStamp, toNtpTime(receiveTime))
			SetUint64(p, TransmitTimeStamp, toNtpTime(time.Now()))
			_, err = s.conn.WriteToUDP(p, remoteAddr)
			if err != nil {
				Error.Printf("worker: %s write failed. %s", remoteAddr.String(), err)
				continue
			}
			s.stats.logIP(remoteAddr)
		default:
			Warn.Printf("%s not client request mode:%x",
				remoteAddr.String(), p[LiVnMode]&^0xf8)
		}
	}
}

func (s *Service) Serve() error {
	if s.cfg.ExpoMetric != "" {
		s.stats = newStatistic(s.cfg)
	}
	ctx := context.TODO()
	for _, p := range s.peerList {
		go p.run(ctx)
	}

	for i := 0; i < s.cfg.WorkerNum; i++ {
		go s.workerDo(i)
	}

	return s.serveSyncLock()
}

func newTemplate() (t []byte) {
	t = make([]byte, 48)
	SetLi(t, 0)
	SetVersion(t, 4)
	SetMode(t, ModeServer)
	SetUint32(t, ReferIDPos, initRefer)
	SetInt8(t, Poll, 4)
	SetUint64(t, ReferenceTimeStamp, toNtpTime(time.Now()))
	return
}
