package gontpd

import "time"

type Config struct {
	MaxStd time.Duration `yaml:"max_std"`

	PeerList  []string `yaml:"peer_list"`
	GeoDB     string   `yaml:"geo_db"`
	Metric    string   `yaml:"metric"`
	Listen    string   `yaml:"listen"`
	WorkerNum int      `yaml:"worker_num"`
	ConnNum   int      `yaml:"conn_num"`
	RateSize  int      `yaml:"rate_size"`

	RateDrop    bool `yaml:"rate_drop"`
	LanDrop     bool `yaml:"lan_drop"`
	ForceUpdate bool `yaml:"force_update"`

	MaxPoll uint8 `yaml:"max_poll"`
	MinPoll uint8 `yaml:"min_poll"`
}
