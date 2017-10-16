package gontpd

import (
	"net"
	"sync"
	"time"
)

const (
	minStep   = 30 * time.Second
	minPoll   = 5
	maxPoll   = 8
	maxAdjust = 128 * time.Millisecond
	initRefer = 0x494e4954 // ascii for INIT
)

type ntpFreq struct {
	overallOffset time.Duration
	x, y          float64
	xx, xy        float64
	samples       int
	num           uint
}

func NewService(cfg *Config) (s *Service, err error) {
	cfg.log()

	s = &Service{
		cfg:      cfg,
		scale:    time.Duration(1),
		status:   &ntpStatus{},
		freq:     &ntpFreq{},
		updateAt: time.Now(),
	}

	if cfg.Listen != "" {
		addr, err := net.ResolveUDPAddr("udp", cfg.Listen)
		if err != nil {
			return nil, err
		}
		s.conn, err = net.ListenUDP("udp", addr)
		if err != nil {
			return nil, err
		}
	}

	for _, host := range cfg.ServerList {
		addrList, err := net.LookupHost(host)
		if err != nil {
			Warn.Printf("peer:%s err:%s", host, err)
			continue
		}
		for _, paddr := range addrList {
			p := newPeer(paddr)
			s.peerList = append(s.peerList, p)
		}
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
	status   *ntpStatus
	freq     *ntpFreq
	scale    time.Duration
	updateAt time.Time
	filters  uint8
}

func (s *Service) Serve() {

	if s.cfg.ExpoMetric != "" {
		s.stats = newStatistic(s.cfg)
	}

	resetClock()

	var wg sync.WaitGroup
	for _, p := range s.peerList {
		wg.Add(1)
		go s.run(p, &wg)
	}

	if s.cfg.Listen != "" {
		for i := 0; i < s.cfg.WorkerNum; i++ {
			go s.workerDo(i)
		}
	}
	wg.Wait()
}
