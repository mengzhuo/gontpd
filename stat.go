package gontpd

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rainycape/geoip"
)

type statistic struct {
	//stats
	reqCounter      *prometheus.CounterVec
	fastCounter     prometheus.Counter
	fastDropCounter *prometheus.CounterVec
	offsetGauge     prometheus.Gauge
	dispGauge       prometheus.Gauge
	delayGauge      prometheus.Gauge
	pollGauge       prometheus.Gauge
	driftGauge      prometheus.Gauge
	geoDB           *geoip.GeoIP
}

func newStatistic(cfg *Config) *statistic {

	var (
		geoDB *geoip.GeoIP
		err   error
	)

	if cfg.GeoDB != "" {
		geoDB, err = geoip.Open(cfg.GeoDB)
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

	fastCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "ntp",
		Subsystem: "requests",
		Name:      "fasttotal",
		Help:      "The total number of ntp request",
	})
	prometheus.MustRegister(fastCounter)

	fastDropCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "ntp",
		Subsystem: "requests",
		Name:      "fastdrop",
		Help:      "The total dropped ntp request",
	}, []string{"reason"})
	prometheus.MustRegister(fastDropCounter)

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
		reqCounter:      reqCounter,
		fastCounter:     fastCounter,
		fastDropCounter: fastDropCounter,
		offsetGauge:     offsetGauge,
		dispGauge:       dispGauge,
		delayGauge:      delayGauge,
		pollGauge:       pollGauge,
		geoDB:           geoDB,
	}
}
