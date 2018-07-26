package gontpd

import (
	"testing"
	"time"
)

func TestToNtpShortTime(t *testing.T) {
	a := toNtpShortTime(time.Second)
	if a != 65536 {
		t.Fatal(a)
	}
}
