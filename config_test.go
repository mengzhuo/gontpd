package gontpd

import (
	"io/ioutil"
	"testing"
)

func TestConfigLoad(t *testing.T) {
	cfg, err := NewConfigFromFile("config.example.yml")
	if err != nil {
		t.Error(err)
	}
	if cfg.ReqRateSec != 4 {
		t.Error(cfg)
		data, err := ioutil.ReadFile("config.example.yml")
		t.Error(string(data), err)
	}
	if cfg.WorkerNum != 7 {
		t.Error(cfg)
	}

	if cfg.GeoDB != "helloWorld.geo" {
		t.Error(cfg)
	}
}
