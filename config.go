package gontpd

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	ServerList []string `yaml:"server"`
	Listen     string   `yaml:"listen"`
	ExpoMetric string   `yaml:"metric"`
	GeoDB      string   `yaml:"geodb"`
	WorkerNum  int      `yaml:"worker"`
	ReqRateSec int      `yaml:"rate"`
}

func (c *Config) log() {
	Info.Printf("listen: %s", c.Listen)
	Info.Printf("export: %s", c.ExpoMetric)
	Info.Printf("geoDB : %s", c.GeoDB)
	Info.Printf("worker: %d", c.WorkerNum)
	Info.Printf("rate  : %d", c.ReqRateSec)
	Info.Printf("srv   : %v", c.ServerList)
}

func NewConfigFromFile(s string) (cfg *Config, err error) {

	var data []byte
	data, err = ioutil.ReadFile(s)
	if err != nil {
		return nil, err
	}
	cfg = &Config{}
	err = yaml.Unmarshal(data, cfg)
	if cfg.WorkerNum == 0 {
		cfg.Listen = ""
	}
	return
}
