package gontpd

import (
	"testing"
	"time"
)

func TestAbsDuration(t *testing.T) {
	a := -time.Second
	c := absDuration(a)
	if c.Seconds() < 1.0 {
		t.Error("abs(a) < 1")
	}
}
