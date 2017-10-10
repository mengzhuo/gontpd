package gontpd

import (
	"fmt"
	"math"
	"strings"
	"syscall"
	"time"
)

func (s *Service) setOffset(p *peer) (err error) {

	Info.Printf("set offset from :%s offset=%s", p.addr, p.offset)

	tmx := &syscall.Timex{}
	offsetNsec := p.offset.Nanoseconds()

	if absDuration(p.offset) < maxAdjust {
		tmx.Modes = ADJ_STATUS | ADJ_NANO | ADJ_OFFSET | ADJ_TIMECONST | ADJ_MAXERROR | ADJ_ESTERROR
		tmx.Status = STA_PLL
		tmx.Offset = offsetNsec
		tmx.Constant = int64(s.poll)
		tmx.Maxerror = 0
		tmx.Esterror = 0
	} else {

		Warn.Printf("settimeofday from %s = %s", p.addr, p.offset)
		tv := syscall.NsecToTimeval(time.Now().Add(p.offset).UnixNano())
		return syscall.Settimeofday(tv)
	}

	switch p.leap {
	case LeapIns:
		tmx.Status |= STA_INS
	case LeapDel:
		tmx.Status |= STA_DEL
	}

	var rc int
	rc, err = syscall.Adjtimex(tmx)
	if rc != 0 {
		return fmt.Errorf("rc=%d status=%s", rc, statusToString(tmx.Status))
	}
	return
}

/*
 * Mode codes (timex.mode)
 */

const (
	ADJ_OFFSET            = 0x0001 /* time offset */
	ADJ_FREQUENCY         = 0x0002 /* frequency offset */
	ADJ_MAXERROR          = 0x0004 /* maximum time error */
	ADJ_ESTERROR          = 0x0008 /* estimated time error */
	ADJ_STATUS            = 0x0010 /* clock status */
	ADJ_TIMECONST         = 0x0020 /* pll time constant */
	ADJ_TAI               = 0x0080 /* set TAI offset */
	ADJ_SETOFFSET         = 0x0100 /* add 'time' to current time */
	ADJ_MICRO             = 0x1000 /* select microsecond resolution */
	ADJ_NANO              = 0x2000 /* select nanosecond resolution */
	ADJ_TICK              = 0x4000 /* tick value */
	ADJ_OFFSET_SINGLESHOT = 0x8001 /* old-fashioned adjtime */
	ADJ_OFFSET_SS_READ    = 0xa001 /* read-only adjtime */

	MOD_OFFSET    = ADJ_OFFSET
	MOD_FREQUENCY = ADJ_FREQUENCY
	MOD_MAXERROR  = ADJ_MAXERROR
	MOD_ESTERROR  = ADJ_ESTERROR
	MOD_STATUS    = ADJ_STATUS
	MOD_TIMECONST = ADJ_TIMECONST
	MOD_TAI       = ADJ_TAI
	MOD_MICRO     = ADJ_MICRO
	MOD_NANO      = ADJ_NANO

	STA_PLL       = 0x0001 /* enable PLL updates (rw) */
	STA_PPSFREQ   = 0x0002 /* enable PPS freq discipline (rw) */
	STA_PPSTIME   = 0x0004 /* enable PPS time discipline (rw) */
	STA_FLL       = 0x0008 /* select frequency-lock mode (rw) */
	STA_INS       = 0x0010 /* insert leap (rw) */
	STA_DEL       = 0x0020 /* delete leap (rw) */
	STA_UNSYNC    = 0x0040 /* clock unsynchronized (rw) */
	STA_FREQHOLD  = 0x0080 /* hold frequency (rw) */
	STA_PPSSIGNAL = 0x0100 /* PPS signal present (ro) */
	STA_PPSJITTER = 0x0200 /* PPS signal jitter exceeded (ro) */
	STA_PPSWANDER = 0x0400 /* PPS signal wander exceeded (ro) */
	STA_PPSERROR  = 0x0800 /* PPS signal calibration error (ro) */
	STA_CLOCKERR  = 0x1000 /* clock hardware fault (ro) */
	STA_NANO      = 0x2000 /* resolution (0 = us, 1 = ns) (ro) */
	STA_MODE      = 0x4000 /* mode (0 = PLL, 1 = FLL) (ro) */
	STA_CLK       = 0x8000 /* clock source (0 = A, 1 = B) (ro) */

	TIME_INS   = 1          /* insert leap second */
	TIME_DEL   = 2          /* delete leap second */
	TIME_OOP   = 3          /* leap second in progress */
	TIME_WAIT  = 4          /* leap second has occurred */
	TIME_ERROR = 5          /* clock not synchronized */
	TIME_BAD   = TIME_ERROR /* bw compat */

	STA_RONLY = (STA_PPSSIGNAL | STA_PPSJITTER | STA_PPSWANDER |
		STA_PPSERROR | STA_CLOCKERR | STA_NANO | STA_MODE | STA_CLK)
)

func statusToString(s int32) (status string) {

	buf := []string{}

	if STA_PLL&s != 0 {
		buf = append(buf, "STA_PLL")
	}
	if STA_PPSFREQ&s != 0 {
		buf = append(buf, "STA_PPSFREQ")
	}
	if STA_PPSTIME&s != 0 {
		buf = append(buf, "STA_PPSTIME")
	}
	if STA_FLL&s != 0 {
		buf = append(buf, "STA_FLL")
	}
	if STA_INS&s != 0 {
		buf = append(buf, "STA_INS")
	}
	if STA_DEL&s != 0 {
		buf = append(buf, "STA_DEL")
	}
	if STA_UNSYNC&s != 0 {
		buf = append(buf, "STA_UNSYNC")
	}
	if STA_FREQHOLD&s != 0 {
		buf = append(buf, "STA_FREQHOLD")
	}
	if STA_PPSSIGNAL&s != 0 {
		buf = append(buf, "STA_PPSSIGNAL")
	}
	if STA_PPSJITTER&s != 0 {
		buf = append(buf, "STA_PPSJITTER")
	}
	if STA_PPSWANDER&s != 0 {
		buf = append(buf, "STA_PPSWANDER")
	}
	if STA_PPSERROR&s != 0 {
		buf = append(buf, "STA_PPSERROR")
	}
	if STA_CLOCKERR&s != 0 {
		buf = append(buf, "STA_CLOCKERR")
	}
	if STA_NANO&s != 0 {
		buf = append(buf, "STA_NANO")
	}
	if STA_MODE&s != 0 {
		buf = append(buf, "STA_MODE")
	}
	if STA_CLK&s != 0 {
		buf = append(buf, "STA_CLK")
	}
	return strings.Join(buf, ", ")
}

func systemPrecision() int8 {
	tmx := &syscall.Timex{}
	syscall.Adjtimex(tmx)
	// linux 1 for usec
	return int8(math.Log2(float64(tmx.Precision) * 1e-6))
}
