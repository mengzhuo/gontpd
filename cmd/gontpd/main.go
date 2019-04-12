package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/mengzhuo/gontpd"
	yaml "gopkg.in/yaml.v2"
)

var (
	fp = flag.String("c", "gontpd.yaml", "yaml config file")
	ff = flag.Int("f", 16, "log flag")
	fv = flag.Bool("v", false, "print version")

	fpprof = flag.String("pprof", "", "pprof listen")

	Version = "dev"
)

func main() {
	flag.Parse()

	if *fv {
		flag.PrintDefaults()
		return
	}

	log.SetFlags(*ff)

	if *ff != 0 {
		log.SetPrefix("[GoNTPd] ")
	}

	if *fpprof != "" {
		go http.ListenAndServe(*fpprof, nil)
	}

	p, err := ioutil.ReadFile(*fp)
	if err != nil {
		log.Fatal(err)
	}
	cfg := &gontpd.Config{}
	err = yaml.Unmarshal(p, cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v", cfg)
	d, err := gontpd.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(d.Run())
}
