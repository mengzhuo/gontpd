package gontpd

import (
	"encoding/binary"
	"log"
	"net"
	"time"
)

type Worker struct {
	conn  *net.UDPConn
	cfg   *Config
	state []byte
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
		if err != nil {
			continue
		}
		if n < 48 {
			continue
		}
		if !isValidNTPRequest(buf) {
			continue
		}
		if w.cfg.InACL(remote.IP) {
			continue
		}
		referTime = binary.BigEndian.Uint64(buf[transmitTimeStamp:])
		copy(buf, w.state)
		binary.BigEndian.PutUint64(buf[originTimeStamp:], referTime)
		binary.BigEndian.PutUint64(buf[receiveTimeStamp:], toNtpTime(rcvTime))
		binary.BigEndian.PutUint64(buf[transmitTimeStamp:], toNtpTime(time.Now()))
		_, err = w.conn.WriteToUDP(buf, remote)
		if err != nil {
			log.Println(err)
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

var ntpEpoch = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)

func toNtpTime(t time.Time) uint64 {
	nsec := uint64(t.Sub(ntpEpoch))
	sec := nsec / nanoPerSec
	frac := (nsec - sec*nanoPerSec) << 32 / nanoPerSec
	return uint64(sec<<32 | frac)
}
