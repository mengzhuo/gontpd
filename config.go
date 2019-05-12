package gontpd

import (
	"fmt"
	"io/ioutil"
	"net"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Listen    string   `yaml:"listen"`
	Workernum uint     `yaml:"worker_num"`
	Metric    string   `yaml:"metric"`
	UpState   string   `yaml:"up_state"`
	ACL       []string `yaml:"acl"`
	rACL      []*net.IPNet

	UpStateIntervalSec uint `yaml:"up_state_sec"`
}

func NewConfig(p string) (cfg *Config, err error) {
	var data []byte
	cfg = &Config{}
	data, err = ioutil.ReadFile(p)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return
	}

	if cfg.Listen == "" {
		err = fmt.Errorf("listen is empty")
		return
	}

	if cfg.UpState == "" {
		err = fmt.Errorf("peer list is empty")
	}
	for _, c := range cfg.ACL {
		var n *net.IPNet
		_, n, err = net.ParseCIDR(c)
		if err != nil {
			return
		}
		cfg.rACL = append(cfg.rACL, n)
	}

	if cfg.UpStateIntervalSec < 3 {
		cfg.UpStateIntervalSec = 3
	}
	return
}

func (c *Config) InACL(ip net.IP) bool {
	for _, n := range c.rACL {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}
