package gontpd

import (
	"context"
	"log"
	"net"
	"syscall"
	"time"
)

func (d *NTPd) listen() {

	var operr error
	lc := net.ListenConfig{func(network, address string, conn syscall.RawConn) (err error) {
		fn := func(fd uintptr) {
			operr = syscall.SetsockoptInt(int(fd),
				syscall.SOL_SOCKET,
				syscall.SO_REUSEADDR, 1)
		}

		if err = conn.Control(fn); err != nil {
			return err
		}

		err = operr
		return
	}}
	lp, err := lc.ListenPacket(context.Background(), "udp", d.cfg.Listen)
	if err != nil {
		log.Print(err)
		return
	}
	d.conn = lp.(*net.UDPConn)
	d.template = newTemplate()
	log.Printf("start listen:%s", d.cfg.Listen)

	for i := 0; i < d.cfg.WorkerNum; i++ {
		go d.worker(i)
	}

}

func (d *NTPd) worker(id int) {
	var (
		n           int
		remoteAddr  *net.UDPAddr
		err         error
		receiveTime time.Time
	)

	p := make([]byte, 48)

	defer func(id int) {
		if r := recover(); r != nil {
			log.Printf("Worker: %d fatal, reason:%s, read:%d", id, r, n)
		} else {
			log.Printf("Worker: %d exited, reason:%s, read:%d", id, err, n)
		}
	}(id)
	log.Printf("worker %d start", id)

	for {
		n, remoteAddr, err = d.conn.ReadFromUDP(p)
		if err != nil {
			return
		}

		receiveTime = time.Now()
		if n < 48 {
			log.Printf("worker: %s get small packet %d",
				remoteAddr.String(), n)
			continue
		}

		// GetMode
		switch p[LiVnModePos] &^ 0xf8 {
		case ModeSymmetricActive:
			// return
			errBuf := make([]byte, 48)
			copy(errBuf, d.template)
			SetUint8(errBuf, StratumPos, 0)
			SetUint32(errBuf, ReferIDPos, 0x41435354)
			d.conn.WriteToUDP(errBuf, remoteAddr)

		case ModeReserved:
			fallthrough
		case ModeClient:
			copy(p[0:OriginTimeStamp], d.template)
			copy(p[OriginTimeStamp:OriginTimeStamp+8],
				p[TransmitTimeStamp:TransmitTimeStamp+8])
			SetUint64(p, ReceiveTimeStamp, toNtpTime(receiveTime))
			SetUint64(p, TransmitTimeStamp, toNtpTime(time.Now()))
			_, err = d.conn.WriteToUDP(p, remoteAddr)
			if err != nil {
				log.Printf("worker: %s write failed. %s", remoteAddr.String(), err)
				continue
			}
			if d.stat != nil {
				d.stat.fastCounter.Add(1)
				d.stat.logIP(remoteAddr)
			}
		default:
			log.Printf("%s not support client request mode:%x",
				remoteAddr.String(), p[LiVnModePos]&^0xf8)
		}
	}

}
