package gontpd

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func New(cfg *Config) (svr *Server, err error) {
	addr, err := net.ResolveUDPAddr("udp", cfg.Listen)
	if err != nil {
		return nil, err
	}
	svr = &Server{cfg: cfg, state: make([]byte, originTimeStamp)}
	svr.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return
	}
	err = svr.followUpState()
	return
}

func (svr *Server) followUpState() (err error) {

	addr, err := net.ResolveUDPAddr("udp", svr.cfg.UpState)
	if err != nil {
		return
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return
	}
	err = conn.SetDeadline(time.Now().Add(3 * time.Second))
	if err != nil {
		return
	}

	msg := makeDummyRequest()
	cookie := make([]byte, 8)
	_, err = rand.Read(cookie)
	if err != nil {
		return
	}
	copy(msg[transmitTimeStamp:], cookie)
	_, err = conn.Write(msg)

	n, _, err := conn.ReadFrom(msg)
	if n < 48 || err != nil {
		err = fmt.Errorf("invalid upstream state:%s", err)
		return
	}
	if !bytes.Equal(msg[originTimeStamp:originTimeStamp+8], cookie) {
		err = fmt.Errorf("mismatch %x vs %x", cookie,
			msg[originTimeStamp:originTimeStamp+8])
		return
	}
	if msg[2] == 0 {
		msg[2] = 0x8
	}
	copy(svr.state, msg)
	return
}

type Server struct {
	worker []*Worker
	state  []byte
	cfg    *Config
	conn   *net.UDPConn
}

func (svr *Server) updateWorker() {
	metaHdr := binary.BigEndian.Uint64(svr.state)
	rootRefHdr := binary.BigEndian.Uint64(svr.state[rootRefOffset:])
	refTimeHdr := binary.BigEndian.Uint64(svr.state[referenceTimeStamp:])
	for i := range svr.worker {
		svr.worker[i].metaHdr = metaHdr
		svr.worker[i].rootRefHdr = rootRefHdr
		svr.worker[i].refTimeHdr = refTimeHdr
	}
}

func (s *Server) Run() {
	for i := uint(0); i < s.cfg.Workernum; i++ {
		worker := &Worker{
			conn: s.conn,
			cfg:  s.cfg}

		if s.cfg.Metric != "" {
			s := &counter{}
			s.total = prometheus.NewCounter(prometheus.CounterOpts{
				Namespace:   "ntp",
				Subsystem:   "requests",
				Name:        "total",
				Help:        "The total number of ntp request",
				ConstLabels: prometheus.Labels{"id": fmt.Sprintf("%d", i)}})
			prometheus.MustRegister(s.total)

			s.drop = prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace:   "ntp",
				Subsystem:   "requests",
				Name:        "drop",
				Help:        "The total dropped ntp request",
				ConstLabels: prometheus.Labels{"id": fmt.Sprintf("%d", i)},
			}, []string{"reason"})
			prometheus.MustRegister(s.drop)
			worker.counter = s
		}
		s.worker = append(s.worker, worker)
	}
	s.updateWorker()

	for i := range s.worker {
		go s.worker[i].run(i)
	}

	if s.cfg.Metric != "" {
		http.Handle("/metrics", promhttp.Handler())
		log.Printf("Listen metric: %s", s.cfg.Metric)
		go http.ListenAndServe(s.cfg.Metric, nil)
	}
	time.Sleep(256 * time.Second)

	for {
		err := s.followUpState()
		if err != nil {
			log.Println(err)
			time.Sleep(16 * time.Second)
			continue
		}
		s.updateWorker()
		time.Sleep(256 * time.Second)
	}
}

func makeDummyRequest() (p []byte) {
	p = make([]byte, 48)
	p[0] = 0xe3 // ntpv4/client/no leap
	p[originTimeStamp] = 0xd
	p[referenceTimeStamp] = 0xe
	p[receiveTimeStamp] = 0xa
	p[transmitTimeStamp] = 0xe
	return
}
