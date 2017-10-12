package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/mengzhuo/gontpd"
)

var (
	fp    = flag.String("c", "config.example.yml", "Go NTP config file")
	level = flag.String("l", "info", "Log level, debug/info/warn/error")
)

func main() {
	flag.Parse()

	nilLogger := log.New(ioutil.Discard, "", log.Ldate)

	switch *level {
	case "info":
	case "warn":
		gontpd.Info = nilLogger
	case "error":
		gontpd.Info = nilLogger
		gontpd.Warn = nilLogger
	}
	gontpd.Info.Print("starting gontpd")

	cfg, err := gontpd.NewConfigFromFile(*fp)
	if err != nil {
		log.Fatal(err)
	}
	service, err := gontpd.NewService(cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(service.Serve())
}
