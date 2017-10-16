package gontpd

import (
	"log"
	"math"
	"strings"
	"syscall"
	"time"
)

var firstAdj bool = true

func resetClock() {

	mode := uint32(adjSTATUS | adjNANO | adjOFFSET | adjFREQUENCY | adjMAXERROR | adjESTERROR | adjTIMECONST)

	tmx := &syscall.Timex{
		Modes:  mode,
		Status: staPLL,
	}

	rc, err := syscall.Adjtimex(tmx)
	if debug {
		log.Print("reset", rc, err)
	}
	if rc == -1 {
		Warn.Print("reset failed", rc, err)
	}
}

func gettimeCorrected() float64 {
	return float64(time.Now().UnixNano())*1e9 + getOffset().Seconds()
}

func getOffset() (offset time.Duration) {
	tmx := &syscall.Timex{
		Status: staPLL,
	}
	rc, err := syscall.Adjtimex(tmx)
	if rc == -1 {
		Error.Printf("get offset failed:%d, %s", rc, err)
		return
	}
	// 1us = 1000 ns
	offset = time.Duration(tmx.Offset * 1000)
	return
}

func (s *Service) setOffset(no *ntpOffset) (synced bool) {

	if debug {
		log.Printf("set offset %v", no.offset)
	}

	d := no.offset
	old := getOffset()
	d += old

	tmx := &syscall.Timex{}
	offsetNsec := d.Nanoseconds()

	if absDuration(d) < maxAdjust {
		Info.Printf("set offset slew offset=%s", no.offset)
		tmx.Modes = adjSTATUS | adjNANO | adjOFFSET | adjMAXERROR | adjESTERROR
		tmx.Status = staPLL
		tmx.Offset = offsetNsec
		tmx.Maxerror = 0
		tmx.Esterror = 0
	} else {
		Warn.Printf("settimeofday from %s", no.offset)
		firstAdj = true
		tv := syscall.NsecToTimeval(time.Now().Add(no.offset).UnixNano())
		Warn.Print(syscall.Settimeofday(&tv))
		return false
	}

	switch no.status.leap {
	case LeapIns:
		tmx.Status |= staINS
	case LeapDel:
		tmx.Status |= staDEL
	}

	rc, _ := syscall.Adjtimex(tmx)
	if rc != 0 {
		Error.Printf("rc=%d status=%s", rc, statusToString(tmx.Status))
	}
	if !firstAdj && old.Nanoseconds() == 0 {
		synced = true
	}
	firstAdj = false
	return
}

/*
 * Mode codes (timex.mode)
 */

const (
	adjOFFSET            = 0x0001 /* time offset */
	adjFREQUENCY         = 0x0002 /* frequency offset */
	adjMAXERROR          = 0x0004 /* maximum time error */
	adjESTERROR          = 0x0008 /* estimated time error */
	adjSTATUS            = 0x0010 /* clock status */
	adjTIMECONST         = 0x0020 /* pll time constant */
	adjTAI               = 0x0080 /* set TAI offset */
	adjSETOFFSET         = 0x0100 /* add 'time' to current time */
	adjMICRO             = 0x1000 /* select microsecond resolution */
	adjNANO              = 0x2000 /* select nanosecond resolution */
	adjTICK              = 0x4000 /* tick value */
	adjOFFSET_SINGLESHOT = 0x8001 /* old-fashioned adjtime */
	adjOFFSET_SS_READ    = 0xa001 /* read-only adjtime */

	modOFFSET    = adjOFFSET
	modFREQUENCY = adjFREQUENCY
	modMAXERROR  = adjMAXERROR
	modESTERROR  = adjESTERROR
	modSTATUS    = adjSTATUS
	modTIMECONST = adjTIMECONST
	modTAI       = adjTAI
	modMICRO     = adjMICRO
	modNANO      = adjNANO

	staPLL       = 0x0001 /* enable PLL updates (rw) */
	staPPSFREQ   = 0x0002 /* enable PPS freq discipline (rw) */
	staPPSTIME   = 0x0004 /* enable PPS time discipline (rw) */
	staFLL       = 0x0008 /* select frequency-lock mode (rw) */
	staINS       = 0x0010 /* insert leap (rw) */
	staDEL       = 0x0020 /* delete leap (rw) */
	staUNSYNC    = 0x0040 /* clock unsynchronized (rw) */
	staFREQHOLD  = 0x0080 /* hold frequency (rw) */
	staPPSSIGNAL = 0x0100 /* PPS signal present (ro) */
	staPPSJITTER = 0x0200 /* PPS signal jitter exceeded (ro) */
	staPPSWANDER = 0x0400 /* PPS signal wander exceeded (ro) */
	staPPSERROR  = 0x0800 /* PPS signal calibration error (ro) */
	staCLOCKERR  = 0x1000 /* clock hardware fault (ro) */
	staNANO      = 0x2000 /* resolution (0 = us, 1 = ns) (ro) */
	staMODE      = 0x4000 /* mode (0 = PLL, 1 = FLL) (ro) */
	staCLK       = 0x8000 /* clock source (0 = A, 1 = B) (ro) */

	timeINS   = 1         /* insert leap second */
	timeDEL   = 2         /* delete leap second */
	timeOOP   = 3         /* leap second in progress */
	timeWAIT  = 4         /* leap second has occurred */
	timeERROR = 5         /* clock not synchronized */
	timeBAD   = timeERROR /* bw compat */

	staRONLY = (staPPSSIGNAL | staPPSJITTER | staPPSWANDER |
		staPPSERROR | staCLOCKERR | staNANO | staMODE | staCLK)
)

func statusToString(s int32) (status string) {

	buf := []string{}

	if staPLL&s != 0 {
		buf = append(buf, "staPLL")
	}
	if staPPSFREQ&s != 0 {
		buf = append(buf, "staPPSFREQ")
	}
	if staPPSTIME&s != 0 {
		buf = append(buf, "staPPSTIME")
	}
	if staFLL&s != 0 {
		buf = append(buf, "staFLL")
	}
	if staINS&s != 0 {
		buf = append(buf, "staINS")
	}
	if staDEL&s != 0 {
		buf = append(buf, "staDEL")
	}
	if staUNSYNC&s != 0 {
		buf = append(buf, "staUNSYNC")
	}
	if staFREQHOLD&s != 0 {
		buf = append(buf, "staFREQHOLD")
	}
	if staPPSSIGNAL&s != 0 {
		buf = append(buf, "staPPSSIGNAL")
	}
	if staPPSJITTER&s != 0 {
		buf = append(buf, "staPPSJITTER")
	}
	if staPPSWANDER&s != 0 {
		buf = append(buf, "staPPSWANDER")
	}
	if staPPSERROR&s != 0 {
		buf = append(buf, "staPPSERROR")
	}
	if staCLOCKERR&s != 0 {
		buf = append(buf, "staCLOCKERR")
	}
	if staNANO&s != 0 {
		buf = append(buf, "staNANO")
	}
	if staMODE&s != 0 {
		buf = append(buf, "staMODE")
	}
	if staCLK&s != 0 {
		buf = append(buf, "staCLK")
	}
	return strings.Join(buf, ", ")
}

func systemPrecision() int8 {
	tmx := &syscall.Timex{}
	syscall.Adjtimex(tmx)
	// linux 1 for usec
	return int8(math.Log2(float64(tmx.Precision) * 1e-6))
}
