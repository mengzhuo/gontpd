package gontpd

import "time"

const (
	metaOffset = iota * 8
	rootRefOffset
	referenceTimeStamp
	originTimeStamp
	receiveTimeStamp
	transmitTimeStamp
)

const (
	ntpEpochNanosecond = -2208988800000000000
	nanoPerSec         = 1e9
)

func toNTPTime(t time.Time) uint64 {
	nsec := t.UnixNano() - ntpEpochNanosecond
	sec := nsec / nanoPerSec
	return uint64(sec<<32 | (nsec-sec*nanoPerSec)<<32/nanoPerSec)
}
