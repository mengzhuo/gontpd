package gontpd

type Config struct {
	PeerList    []string
	Listen      string
	WorkerNum   int
	ForceUpdate bool
}
