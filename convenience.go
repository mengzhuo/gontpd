package gontpd

import (
	"math"
	"math/rand"
	"time"
)

const (
	nanoSecPerSec = int64(time.Second)
)

func maxDuration(a, b time.Duration) time.Duration {
	if a < b {
		return b
	}
	return a
}

func minDuration(a, b time.Duration) time.Duration {
	if a > b {
		return b
	}
	return a
}

func absDurationLess(a, b time.Duration) bool {
	if a < 0 {
		a = -a
	}
	return a < b
}

func absDuration(a time.Duration) time.Duration {
	if a < 0 {
		return -a
	}
	return a
}

func secondToDuration(a float64) time.Duration {
	return time.Duration(a * float64(time.Second))
}

func randDuration() time.Duration {
	return secondToDuration(float64(rand.Intn(3000)) / 1000)
}
func durationToPoll(t time.Duration) int8 {
	return int8(math.Log2(absDuration(t).Seconds()))
}

func reverseToInterval(d time.Duration) int8 {
	return int8(math.Log2(d.Seconds()))
}
