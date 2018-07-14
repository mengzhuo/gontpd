package gontpd

import (
	"log"
	"net"
	"net/http"

	geoip2 "github.com/oschwald/geoip2-golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type statistic struct {
	//stats
	reqCounter  *prometheus.CounterVec
	fastCounter *prometheus.CounterVec
	offsetGauge prometheus.Gauge
	dispGauge   prometheus.Gauge
	delayGauge  prometheus.Gauge
	pollGauge   prometheus.Gauge
	driftGauge  prometheus.Gauge
	geoDB       *geoip2.Reader
}

func newStatistic(cfg *Config) *statistic {

	var (
		geoDB *geoip2.Reader
		err   error
	)

	if cfg.GeoDB != "" {
		geoDB, err = geoip2.Open(cfg.GeoDB)
		if err != nil {
			log.Fatal(err)
			return nil
		}
	}

	reqCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "ntp",
		Subsystem: "requests",
		Name:      "total",
		Help:      "The total number of ntp request",
	}, []string{"cc"})

	prometheus.MustRegister(reqCounter)
	fastCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "ntp",
		Subsystem: "requests",
		Name:      "fasttotal",
		Help:      "The total number of ntp request",
	}, []string{"worker"})
	prometheus.MustRegister(fastCounter)

	offsetGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "ntp",
		Subsystem: "stat",
		Name:      "offset_sec",
		Help:      "The offset to upper peer",
	})
	prometheus.MustRegister(offsetGauge)

	dispGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "ntp",
		Subsystem: "stat",
		Name:      "dispersion_sec",
		Help:      "The dispersion of service",
	})
	prometheus.MustRegister(dispGauge)

	delayGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "ntp",
		Subsystem: "stat",
		Name:      "delay_sec",
		Help:      "The root delay of service",
	})
	prometheus.MustRegister(delayGauge)

	pollGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "ntp",
		Subsystem: "stat",
		Name:      "poll_interval_sec",
		Help:      "The poll interval of ntp",
	})
	prometheus.MustRegister(pollGauge)

	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Listen metric: %s", cfg.Metric)
	go http.ListenAndServe(cfg.Metric, nil)

	return &statistic{
		reqCounter:  reqCounter,
		fastCounter: fastCounter,
		offsetGauge: offsetGauge,
		dispGauge:   dispGauge,
		delayGauge:  delayGauge,
		pollGauge:   pollGauge,
		geoDB:       geoDB,
	}
}

func (s *statistic) logIP(raddr *net.UDPAddr) {

	if s.geoDB == nil {
		return
	}

	country, err := s.geoDB.Country(raddr.IP)
	if err != nil {
		log.Print("stat ip err=", err)
		return
	}
	s.reqCounter.WithLabelValues(country.Country.IsoCode).Inc()
}
