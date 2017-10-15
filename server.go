package gontpd

import (
	"net"
	"time"
)

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

func (s *Service) setTemplate(no *ntpOffset) {

	SetLi(s.template, s.status.leap)
	SetVersion(s.template, 4)
	SetMode(s.template, ModeServer)

	SetUint8(s.template, Stratum, s.status.stratum)
	SetInt8(s.template, ClockPrecision, systemPrecision())
	SetUint32(s.template, RootDelayPos, toNtpShortTime(s.status.rootDelay))

	if s.stats != nil {
		s.stats.delayGauge.Set(no.delay.Seconds())
		s.stats.offsetGauge.Set(no.offset.Seconds())
		s.stats.dispGauge.Set(no.err.Seconds())
	}

	SetUint32(s.template, RootDispersionPos, toNtpShortTime(s.status.rootDispersion))
	SetUint64(s.template, ReferenceTimeStamp, toNtpTime(s.status.refTime))
	SetUint32(s.template, ReferIDPos, s.status.sendRefId)

	SetInt8(s.template, Poll, int8(s.status.poll))
}

func (s *Service) workerDo(i int) {
	var (
		n           int
		remoteAddr  *net.UDPAddr
		err         error
		receiveTime time.Time
	)

	p := make([]byte, 48)

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

		// GetMode
		switch p[LiVnMode] &^ 0xf8 {
		case ModeSymmetricActive:
			// return
			errBuf := make([]byte, 48)
			copy(errBuf, s.template)
			SetUint8(errBuf, Stratum, 0)
			SetUint32(errBuf, ReferIDPos, 0x41435354)
			s.conn.WriteToUDP(errBuf, remoteAddr)

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
			if s.stats != nil {
				s.stats.logIP(remoteAddr)
			}
		default:
			Warn.Printf("%s not support client request mode:%x",
				remoteAddr.String(), p[LiVnMode]&^0xf8)
		}
	}
}
