package gontpd

type Config struct {
	PeerList    []string
	GeoDB       string
	Metric      string
	Listen      string
	WorkerNum   int
	ConnNum     int
	CacheSize   int
	LanDrop     bool
	ForceUpdate bool

	MaxPoll uint8
	MinPoll uint8
}
