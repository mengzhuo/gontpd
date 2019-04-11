package gontpd

import (
	"encoding/binary"
	"log"
	"net"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Worker struct {
	conn    *net.UDPConn
	cfg     *Config
	state   []byte
	counter *counter
}

type counter struct {
	total prometheus.Counter
	drop  *prometheus.CounterVec
}

func (w *Worker) run(i uint) {
	log.Printf("worker:%d online", i)
	var (
		buf    []byte
		n      int
		remote *net.UDPAddr
		err    error

		rcvTime   time.Time
		referTime uint64
	)

	buf = make([]byte, 48, 48)
	// BCE
	_ = buf[47]
	_ = w.state[originTimeStamp-1]

	for {
		n, remote, err = w.conn.ReadFromUDP(buf)
		rcvTime = time.Now()
		if err != nil || remote.Port == 0 {
			continue
		}
		if n < 48 {
			if w.counter != nil {
				w.counter.drop.WithLabelValues("small").Inc()
			}
			continue
		}
		if !isValidNTPRequest(buf) {
			if w.counter != nil {
				w.counter.drop.WithLabelValues("invalid").Inc()
			}
			continue
		}
		if w.cfg.InACL(remote.IP) {
			if w.counter != nil {
				w.counter.drop.WithLabelValues("acl").Inc()
			}
			continue
		}
		referTime = binary.BigEndian.Uint64(buf[transmitTimeStamp:])
		copy(buf, w.state)
		binary.BigEndian.PutUint64(buf[originTimeStamp:], referTime)
		binary.BigEndian.PutUint64(buf[receiveTimeStamp:], toNTPTime(rcvTime))
		binary.BigEndian.PutUint64(buf[transmitTimeStamp:], toNTPTime(time.Now()))
		_, err = w.conn.WriteToUDP(buf, remote)
		if err != nil {
			log.Println(err)
		}
		if w.counter != nil {
			w.counter.total.Inc()
		}
	}
}

func isValidNTPRequest(p []byte) (r bool) {
	switch p[0] &^ 0xf8 {
	case 3: // modeClient
	default:
		return
	}

	return true
}
