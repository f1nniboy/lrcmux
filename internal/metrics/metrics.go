package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// goes past the provider timeout, since one lookup can run several tiers
var latencyBuckets = []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 15, 30, 60}

type Collector struct {
	HTTPRequests    *prometheus.CounterVec
	HTTPLatency     *prometheus.HistogramVec
	ProviderOps     *prometheus.CounterVec
	ProviderLatency *prometheus.HistogramVec
	Served          *prometheus.CounterVec
	RequestOutcomes *prometheus.CounterVec
	registry        *prometheus.Registry
	Listen          string
}

func New(listen string) *Collector {
	reg := prometheus.NewRegistry()
	c := &Collector{Listen: listen, registry: reg}

	c.HTTPRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lrcmux_http_requests_total",
		Help: "Total /get requests by format, level, and status",
	}, []string{"format", "level", "status"})

	c.HTTPLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "lrcmux_http_request_duration_seconds",
		Help:    "/get request latency by format, level, and status",
		Buckets: latencyBuckets,
	}, []string{"format", "level", "status"})

	c.ProviderOps = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lrcmux_provider_requests_total",
		Help: "Provider fan-out requests by provider and outcome",
	}, []string{"prov", "outcome"})

	c.ProviderLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "lrcmux_provider_request_duration_seconds",
		Help:    "Provider request latency",
		Buckets: latencyBuckets,
	}, []string{"prov"})

	c.Served = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lrcmux_served_total",
		Help: "Lyrics responses by the provider whose result was served",
	}, []string{"prov"})

	c.RequestOutcomes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lrcmux_request_outcomes_total",
		Help: "Lookups by how they were answered: cache, fanout, or a failed track match",
	}, []string{"stage"})

	reg.MustRegister(
		c.HTTPRequests,
		c.HTTPLatency,
		c.ProviderOps,
		c.ProviderLatency,
		c.Served,
		c.RequestOutcomes,
	)
	return c
}

func (c *Collector) Register(col prometheus.Collector) {
	c.registry.MustRegister(col)
}

func (c *Collector) Handler() http.Handler {
	return promhttp.HandlerFor(c.registry, promhttp.HandlerOpts{})
}
