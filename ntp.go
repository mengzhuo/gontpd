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
	modeReserved uint8 = iota
	modeSymmetricActive
	modeSymmetricPassive
	modeClient
	modeServer
	modeBroadcast
	modeControlMessage
	modeReservedPrivate
)

const (
	liVnModePos = iota
	stratumPos
	pollPos
	clockPrecisionPos
)

const (
	rootDelayPos = iota*4 + 4
	rootDispersionPos
	referIDPos
)

const (
	referenceTimeStamp = iota*8 + 16
	originTimeStamp
	receiveTimeStamp
	transmitTimeStamp
)

var ntpEpoch = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)

func toNtpTime(t time.Time) uint64 {
	nsec := uint64(t.Sub(ntpEpoch))
	sec := nsec / nanoPerSec
	frac := (nsec - sec*nanoPerSec) << 32 / nanoPerSec
	return uint64(sec<<32 | frac)
}

func setLi(m []byte, li uint8) {
	m[0] = (m[0] & 0x3f) | li<<6
}

func setMode(m []byte, mode uint8) {
	m[0] = (m[0] & 0xf8) | mode
}

func getMode(m []byte) uint8 {
	return m[0] &^ 0xf8
}

func setVersion(m []byte, v uint8) {
	m[0] = (m[0] & 0xc7) | v<<3
}

func setUint64(m []byte, index int, value uint64) {
	binary.BigEndian.PutUint64(m[index:], value)
}

func setUint8(m []byte, index int, value uint8) {
	m[index] = value
}

func setInt8(m []byte, index int, value int8) {
	// bigEndian
	m[index] = byte(value)
}

func setUint32(m []byte, index int, value uint32) {
	binary.BigEndian.PutUint32(m[index:], value)
}

func toNtpShortTime(t time.Duration) uint32 {
	sec := t / nanoPerSec
	frac := (t - sec*nanoPerSec) << 16 / nanoPerSec
	return uint32(sec<<16 | frac)
}

func newTemplate() (t []byte) {
	t = make([]byte, 48)
	setLi(t, noLeap)
	setVersion(t, 4)
	setMode(t, modeServer)
	setUint32(t, referIDPos, initRefer)
	setInt8(t, pollPos, minPoll)
	setUint8(t, stratumPos, 0xff)
	setUint64(t, referenceTimeStamp, toNtpTime(time.Now()))
	return
}

func (d *NTPd) setTemplate(op *offsetPeer) {

	setLi(d.template, uint8(op.resp.Leap))
	setMode(d.template, modeServer)

	setUint8(d.template, stratumPos, op.resp.Stratum+1)
	setInt8(d.template, clockPrecisionPos, systemPrecision())

	d.delay = op.resp.RootDelay + op.resp.RTT/2
	setUint32(d.template, rootDelayPos, toNtpShortTime(d.delay))

	d.disp = op.resp.RootDelay/2 + op.resp.RootDispersion
	setUint32(d.template, rootDispersionPos,
		toNtpShortTime(d.disp))
	setUint64(d.template, referenceTimeStamp, toNtpTime(op.resp.Time))
	setUint32(d.template, referIDPos, op.peer.refId)

	setInt8(d.template, pollPos, int8(op.peer.trustLevel))
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
