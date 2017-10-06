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
	if cfg.ReqRateSec != 0 {
		t.Error(cfg)
		data, err := ioutil.ReadFile("config.example.yml")
		t.Error(string(data), err)
	}
	if cfg.WorkerNum != 2 {
		t.Error(cfg)
	}

	if cfg.GeoDB != "GeoLite2-Country.mmdb" {
		t.Error(cfg)
	}
}
