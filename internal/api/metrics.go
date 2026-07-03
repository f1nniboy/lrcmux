package api

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/f1nniboy/lrcmux/internal/orchestrator"
)

type breakerCollector struct {
	orch      *orchestrator.Orchestrator
	stateDesc *prometheus.Desc
	infoDesc  *prometheus.Desc
}

func newBreakerCollector(orch *orchestrator.Orchestrator) *breakerCollector {
	return &breakerCollector{
		orch: orch,
		stateDesc: prometheus.NewDesc(
			"lrcmux_provider_breaker_open",
			"1 if the provider circuit breaker is open, 0 if closed",
			[]string{"prov"},
			nil,
		),
		infoDesc: prometheus.NewDesc(
			"lrcmux_provider_breaker_info",
			"Reason the provider circuit breaker is open, only present when open",
			[]string{"prov", "reason"},
			nil,
		),
	}
}

func (c *breakerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.stateDesc
	ch <- c.infoDesc
}

func (c *breakerCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	for _, info := range c.orch.ProviderInfos(ctx) {
		var v float64
		if !info.Health.Ok {
			v = 1
			ch <- prometheus.MustNewConstMetric(c.infoDesc, prometheus.GaugeValue, 1, info.ID, info.Health.Reason)
		}
		ch <- prometheus.MustNewConstMetric(c.stateDesc, prometheus.GaugeValue, v, info.ID)
	}
}
