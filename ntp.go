package gontpd

import (
	"encoding/binary"
	"time"
)

var (
	epoch     = time.Unix(0, 0)
	pollTable = [...]time.Duration{
		1 << (minPoll + 0) * time.Second,
		1 << (minPoll + 1) * time.Second,
		1 << (minPoll + 2) * time.Second,
		1 << (minPoll + 3) * time.Second,
		1 << (minPoll + 4) * time.Second,
		1 << (minPoll + 5) * time.Second,
		1 << (minPoll + 6) * time.Second,
		1 << (minPoll + 7) * time.Second,
		1 << (minPoll + 8) * time.Second,
		1 << (minPoll + 9) * time.Second,
		1 << (minPoll + 10) * time.Second,
		1 << (minPoll + 11) * time.Second,
	}
)

const (
	nanoPerSec = 1e9

	// INIT
	initRefer = 0x494e4954

	// ACST | The association belongs to a unicast server.
	acstKoD = 0x41435354

	// RATE | Rate exceeded
	rateKoD = 0x52415445
)

const (
	ModeReserved uint8 = iota
	ModeSymmetricActive
	ModeSymmetricPassive
	ModeClient
	ModeServer
	ModeBroadcast
	ModeControlMessage
	ModeReservedPrivate
)

const (
	LiVnModePos = iota
	StratumPos
	PollPos
	ClockPrecisionPos
)

const (
	RootDelayPos = iota*4 + 4
	RootDispersionPos
	ReferIDPos
)

const (
	ReferenceTimeStamp = iota*8 + 16
	OriginTimeStamp
	ReceiveTimeStamp
	TransmitTimeStamp
)

var ntpEpoch = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)

func toNtpTime(t time.Time) uint64 {
	nsec := uint64(t.Sub(ntpEpoch))
	sec := nsec / nanoPerSec
	frac := (nsec - sec*nanoPerSec) << 32 / nanoPerSec
	return uint64(sec<<32 | frac)
}

func SetLi(m []byte, li uint8) {
	m[0] = (m[0] & 0x3f) | li<<6
}

func SetMode(m []byte, mode uint8) {
	m[0] = (m[0] & 0xf8) | mode
}

func GetMode(m []byte) uint8 {
	return m[0] &^ 0xf8
}

func SetVersion(m []byte, v uint8) {
	m[0] = (m[0] & 0xc7) | v<<3
}

func SetUint64(m []byte, index int, value uint64) {
	binary.BigEndian.PutUint64(m[index:], value)
}

func SetUint8(m []byte, index int, value uint8) {
	m[index] = value
}

func SetInt8(m []byte, index int, value int8) {
	// bigEndian
	m[index] = byte(value)
}

func SetUint32(m []byte, index int, value uint32) {
	binary.BigEndian.PutUint32(m[index:], value)
}

func toNtpShortTime(t time.Duration) uint32 {
	sec := t / nanoPerSec
	frac := (t - sec*nanoPerSec) << 16 / nanoPerSec
	return uint32(sec<<16 | frac)
}

func newTemplate() (t []byte) {
	t = make([]byte, 48)
	SetLi(t, NoLeap)
	SetVersion(t, 4)
	SetMode(t, ModeServer)
	SetUint32(t, ReferIDPos, initRefer)
	SetInt8(t, PollPos, minPoll)
	SetUint8(t, StratumPos, 0xff)
	SetUint64(t, ReferenceTimeStamp, toNtpTime(time.Now()))
	return
}

func (d *NTPd) setTemplate(op *offsetPeer) {

	SetLi(d.template, uint8(op.resp.Leap))
	SetMode(d.template, ModeServer)

	SetUint8(d.template, StratumPos, op.resp.Stratum+1)
	SetInt8(d.template, ClockPrecisionPos, systemPrecision())

	d.delay = op.resp.RootDelay + op.resp.RTT/2
	SetUint32(d.template, RootDelayPos, toNtpShortTime(d.delay))

	d.disp = op.resp.RootDelay/2 + op.resp.RootDispersion
	SetUint32(d.template, RootDispersionPos,
		toNtpShortTime(d.disp))
	SetUint64(d.template, ReferenceTimeStamp, toNtpTime(op.resp.Time))
	SetUint32(d.template, ReferIDPos, op.peer.refId)

	SetInt8(d.template, PollPos, int8(op.peer.trustLevel))
}

func stddev(pl []time.Duration) time.Duration {
	var sum time.Duration
	for _, p := range pl {
		sum += absDuration(p)
	}
	avg := sum / time.Duration(len(pl))
	sum = 0
	al := make([]time.Duration, len(pl))
	for i := 0; i < len(pl); i++ {
		off := pl[i] - avg
		al[i] = off * off
		sum += al[i]
	}
	sum = sum / time.Duration(len(pl))
	return time.Duration(uintSqrt(uint64(sum)))
}

// https://en.wikipedia.org/wiki/Integer_square_root
func uintSqrt(n uint64) uint64 {
	if n < 2 {
		return n
	}
	smallCandidate := uintSqrt(n>>2) << 1
	largeCandidate := smallCandidate + 1
	if largeCandidate*largeCandidate > n {
		return smallCandidate
	}
	return largeCandidate
}
