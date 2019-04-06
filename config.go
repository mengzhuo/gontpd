package gontpd

import (
	"fmt"
	"io/ioutil"
	"net"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Listen    string   `yaml:"listen"`
	Workernum uint     `yaml:"workernum"`
	Connnum   uint     `yaml:"connnum"`
	Metric    string   `yaml:"metric"`
	UpState   string   `yaml:"up_state"`
	ACL       []string `yaml:"acl"`
	rACL      []net.IPMask
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
	return
}

func (c *Config) InACL(ip net.IP) bool {
	return false
}
