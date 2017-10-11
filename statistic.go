package gontpd

import (
	"net"
	"net/http"

	geoip2 "github.com/oschwald/geoip2-golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type statistic struct {
	//stats
	reqCounter  *prometheus.CounterVec
	fastCounter prometheus.Counter
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
			Error.Fatal(err)
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

	driftGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "ntp",
		Subsystem: "stat",
		Name:      "drift_ppm",
		Help:      "The drift of ntp",
	})
	prometheus.MustRegister(driftGauge)

	http.Handle("/metrics", promhttp.Handler())
	Info.Printf("Listen metric: %s", cfg.ExpoMetric)
	go http.ListenAndServe(cfg.ExpoMetric, nil)

	return &statistic{
		reqCounter:  reqCounter,
		fastCounter: fastCounter,
		offsetGauge: offsetGauge,
		dispGauge:   dispGauge,
		delayGauge:  delayGauge,
		pollGauge:   pollGauge,
		driftGauge:  driftGauge,
		geoDB:       geoDB,
	}
}

func (s *statistic) logIP(raddr *net.UDPAddr) {

	s.fastCounter.Inc()

	if s.geoDB == nil {
		return
	}

	country, err := s.geoDB.Country(raddr.IP)
	if err != nil {
		Error.Print("stat ip err=", err)
		return
	}
	s.reqCounter.WithLabelValues(country.Country.IsoCode).Inc()
}
