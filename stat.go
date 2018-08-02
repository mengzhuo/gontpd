package gontpd

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rainycape/geoip"
)

type workerStat struct {
	CCReq   *prometheus.CounterVec
	Req     prometheus.Counter
	ACL     prometheus.Counter
	Rate    prometheus.Counter
	Malform prometheus.Counter
	Unknown prometheus.Counter
	GeoDB   *geoip.GeoIP
}

func newWorkerStat(id string) (s *workerStat) {

	s = &workerStat{}
	s.CCReq = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   "ntp",
		Subsystem:   "requests",
		Name:        "total",
		Help:        "The total number of ntp request",
		ConstLabels: prometheus.Labels{"id": id},
	}, []string{"cc"})
	prometheus.MustRegister(s.CCReq)

	s.Req = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace:   "ntp",
		Subsystem:   "requests",
		Name:        "fasttotal",
		Help:        "The total number of ntp request",
		ConstLabels: prometheus.Labels{"id": id},
	})
	prometheus.MustRegister(s.Req)

	s.ACL = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace:   "ntp",
		Subsystem:   "requests",
		Name:        "drop",
		Help:        "The total dropped ntp request",
		ConstLabels: prometheus.Labels{"id": id, "reason": "acl"},
	})
	prometheus.MustRegister(s.ACL)

	s.Rate = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace:   "ntp",
		Subsystem:   "requests",
		Name:        "drop",
		Help:        "The total dropped ntp request",
		ConstLabels: prometheus.Labels{"id": id, "reason": "rate"},
	})
	prometheus.MustRegister(s.Rate)

	s.Malform = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace:   "ntp",
		Subsystem:   "requests",
		Name:        "drop",
		Help:        "The total dropped ntp request",
		ConstLabels: prometheus.Labels{"id": id, "reason": "malform"},
	})
	prometheus.MustRegister(s.Malform)

	s.Unknown = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace:   "ntp",
		Subsystem:   "requests",
		Name:        "drop",
		Help:        "The total dropped ntp request",
		ConstLabels: prometheus.Labels{"id": id, "reason": "unknown_method"},
	})
	prometheus.MustRegister(s.Unknown)
	return
}

type ntpStat struct {
	offsetGauge prometheus.Gauge
	dispGauge   prometheus.Gauge
	delayGauge  prometheus.Gauge
	pollGauge   prometheus.Gauge
	driftGauge  prometheus.Gauge
}

func newNTPStat(listen string) *ntpStat {

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
	log.Printf("Listen metric: %s", listen)
	go http.ListenAndServe(listen, nil)

	return &ntpStat{
		offsetGauge: offsetGauge,
		dispGauge:   dispGauge,
		delayGauge:  delayGauge,
		pollGauge:   pollGauge,
	}
}
