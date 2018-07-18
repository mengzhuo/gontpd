package gontpd

import (
	"context"
	"fmt"
	"log"
	"net"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

func (d *NTPd) listen() {

	for j := 0; j < d.cfg.ConnNum; j++ {
		conn, err := d.makeConn()
		if err != nil {
			log.Fatal(err)
		}
		for i := 0; i < d.cfg.WorkerNum; i++ {
			go d.worker(fmt.Sprintf("%d:%d", j, i), conn)
		}
	}
}

func (d *NTPd) makeConn() (conn *net.UDPConn, err error) {

	var operr error

	cfgFn := func(network, address string, conn syscall.RawConn) (err error) {

		fn := func(fd uintptr) {
			operr = syscall.SetsockoptInt(int(fd),
				syscall.SOL_SOCKET,
				unix.SO_REUSEPORT, 1)
		}

		if err = conn.Control(fn); err != nil {
			return err
		}

		err = operr
		return
	}
	lc := net.ListenConfig{Control: cfgFn}
	lp, err := lc.ListenPacket(context.Background(), "udp", d.cfg.Listen)
	if err != nil {
		return
	}
	conn = lp.(*net.UDPConn)
	return
}

func (d *NTPd) worker(id string, conn *net.UDPConn) {
	var (
		n           int
		remoteAddr  *net.UDPAddr
		err         error
		receiveTime time.Time
	)

	p := make([]byte, 48)
	oob := make([]byte, 1)

	defer func(id string) {
		if r := recover(); r != nil {
			log.Printf("Worker: %s fatal, reason:%s, read:%d", id, r, n)
		} else {
			log.Printf("Worker: %s exited, reason:%s, read:%d", id, err, n)
		}
	}(id)

	log.Printf("worker %s started", id)

	for {
		n, _, _, remoteAddr, err = conn.ReadMsgUDP(p, oob)
		if err != nil {
			return
		}

		receiveTime = time.Now()
		if n < 48 {
			if debug {
				log.Printf("worker: %s get small packet %d",
					remoteAddr.String(), n)
			}
			continue
		}
		// BCE
		_ = p[47]

		// GetMode
		switch p[LiVnModePos] &^ 0xf8 {
		case ModeSymmetricActive:
			// return
			errBuf := make([]byte, 48)
			copy(errBuf, d.template)
			SetUint8(errBuf, StratumPos, 0)
			SetUint32(errBuf, ReferIDPos, acstKoD)
			conn.WriteToUDP(errBuf, remoteAddr)

		case ModeReserved:
			fallthrough
		case ModeClient:
			copy(p[0:OriginTimeStamp], d.template)
			copy(p[OriginTimeStamp:OriginTimeStamp+8],
				p[TransmitTimeStamp:TransmitTimeStamp+8])
			SetUint64(p, ReceiveTimeStamp, toNtpTime(receiveTime))
			SetUint64(p, TransmitTimeStamp, toNtpTime(time.Now()))
			_, err = conn.WriteToUDP(p, remoteAddr)
			if err != nil && debug {
				log.Printf("worker: %s write failed. %s", remoteAddr.String(), err)
			}
			if d.stat != nil {
				d.stat.fastCounter.WithLabelValues(id).Inc()
				d.stat.logIP(remoteAddr)
			}
		default:
			if debug {
				log.Printf("%s not support client request mode:%x",
					remoteAddr.String(), p[LiVnModePos]&^0xf8)
			}
		}
	}
}
