package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/mengzhuo/gontpd"
	yaml "gopkg.in/yaml.v2"
)

var (
	fp = flag.String("c", "gontpd.yaml", "yaml config file")
	ff = flag.Int("f", 16, "log flag")
)

func main() {
	flag.Parse()

	log.SetFlags(*ff)

	p, err := ioutil.ReadFile(*fp)
	if err != nil {
		log.Fatal(err)
	}
	cfg := &gontpd.Config{}
	err = yaml.Unmarshal(p, cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%#v", cfg)
	d := gontpd.New(cfg)
	log.Fatal(d.Run())
}
