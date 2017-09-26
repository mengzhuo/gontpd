package gontpd

import (
	"io/ioutil"
	"runtime"

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

func NewConfigFromFile(s string) (cfg *Config, err error) {

	var data []byte
	data, err = ioutil.ReadFile(s)
	if err != nil {
		return nil, err
	}
	cfg = &Config{}
	err = yaml.Unmarshal(data, cfg)
	if cfg.WorkerNum == 0 {
		cfg.WorkerNum = runtime.NumCPU()
	}
	Info.Printf("%#v", cfg)
	return
}
