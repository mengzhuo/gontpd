package gontpd

const (
	msgAdjTime = iota + 1
	msgAdjFreq
	msgSetTime
)

type ctrlMsg struct {
	id    int
	delta *ntpOffset
	freq  float64
}
