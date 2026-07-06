package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Collector struct {
	Listen          string
	HTTPRequests    *prometheus.CounterVec
	HTTPLatency     *prometheus.HistogramVec
	CacheOps        *prometheus.CounterVec
	ProviderOps     *prometheus.CounterVec
	ProviderLatency *prometheus.HistogramVec
	RequestOutcomes *prometheus.CounterVec
	registry        *prometheus.Registry
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
		Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"format", "level", "status"})

	c.CacheOps = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lrcmux_cache_ops_total",
		Help: "Request cache results (hit/miss)",
	}, []string{"result"})

	c.ProviderOps = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lrcmux_provider_requests_total",
		Help: "Provider fan-out requests by provider and outcome",
	}, []string{"prov", "outcome"})

	c.ProviderLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "lrcmux_provider_request_duration_seconds",
		Help:    "Provider request latency",
		Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"prov"})

	c.RequestOutcomes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lrcmux_request_outcomes_total",
		Help: "Request funnel stages: isrc_not_found, cache_hit, fanout, breakers_open",
	}, []string{"stage"})

	reg.MustRegister(
		c.HTTPRequests,
		c.HTTPLatency,
		c.CacheOps,
		c.ProviderOps,
		c.ProviderLatency,
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
