package orchestrator

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/f1nniboy/lrcmux/internal/cache"
	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

type providerOutcome struct {
	err     error
	result  *lyrics.Result
	source  lyrics.Source
	latency time.Duration
}

func (o *Orchestrator) fanOut(ctx context.Context, active []providers.Provider, q lyrics.Query, level lyrics.SyncLevel) []*lyrics.Result {
	fanCtx, cancel := context.WithTimeout(ctx, o.opts.Timeout)
	defer cancel()

	o.log.Debug("fanning out", "providers", providers.IDs(active), "target_level", level.String(), "timeout", o.opts.Timeout.Milliseconds())

	ch := make(chan providerOutcome, len(active))
	var wg sync.WaitGroup
	for _, p := range active {
		wg.Go(func() {
			o.log.Debug("querying provider", "provider", p.ID())
			start := time.Now()
			r, err := p.Search(fanCtx, q)
			ch <- providerOutcome{source: providers.Source(p), result: r, err: err, latency: time.Since(start)}
		})
	}
	go func() { wg.Wait(); close(ch) }()

	var results []*lyrics.Result
	var misses []string

	collect := func(out providerOutcome) {
		if out.err == nil && out.result != nil {
			out.result.Source = out.source
			out.result.Lines = lyrics.CleanLines(out.result.Lines)
			results = append(results, out.result)
		}
		if errors.Is(out.err, lyrics.ErrNotFound) {
			misses = append(misses, out.source.ID)
		}
		outcome := o.logOutcome(out, q)
		o.breaker.Record(out.source.ID, outcome)
	}

	for out := range ch {
		collect(out)
		if out.result != nil && satisfies(out.result, level) {
			o.log.Debug("target satisfied, cancelling remaining", "provider", out.source.ID, "sync", out.result.SyncLevel.String())
			cancel()
			break
		}
	}
	for out := range ch {
		collect(out)
	}

	o.log.Debug("fanout done", "collected", len(results), "of", len(active))

	go func() {
		bg := context.Background()
		ids := make([]string, len(results))

		for i, r := range results {
			ids[i] = r.Source.ID
			if err := cache.Set(bg, o.cache, cacheKey(q.Track.ISRC, r.Source.ID), *r, o.ttlFor(r)); err != nil {
				o.log.Warn("cache set failed", "err", err, "provider", r.Source.ID)
			}
		}
		for _, provider := range misses {
			if err := cache.SetMiss(bg, o.cache, cacheKey(q.Track.ISRC, provider), o.opts.TTL.Miss.Duration); err != nil {
				o.log.Warn("miss cache set failed", "provider", provider, "err", err)
			}
		}

		o.breaker.ResetStreak(bg, ids)
	}()

	return results
}

func (o *Orchestrator) logOutcome(out providerOutcome, q lyrics.Query) string {
	var (
		lvl     = slog.LevelDebug
		outcome string
		extra   []any
	)

	switch {
	case out.err == nil && out.result != nil:
		outcome = "ok"
		extra = []any{"level", out.result.SyncLevel.String()}
	case errors.Is(out.err, lyrics.ErrNotFound):
		outcome = "not_found"
	case errors.Is(out.err, providers.ErrRateLimited):
		lvl, outcome = slog.LevelInfo, "rate_limited"
	case errors.Is(out.err, context.DeadlineExceeded):
		lvl, outcome = slog.LevelInfo, "timeout"
	case errors.Is(out.err, context.Canceled):
		outcome = "canceled"
	case isNetworkNoise(out.err):
		lvl, outcome = slog.LevelInfo, "network_error"
		extra = []any{"err", out.err}
	default:
		lvl, outcome = slog.LevelWarn, "error"
		extra = []any{"err", out.err}

		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetTag("provider", out.source.ID)
			scope.SetContext("query", sentry.Context{
				"isrc":   q.Track.ISRC,
				"artist": q.Track.Artist,
				"title":  q.Track.Title,
			})
			sentry.CaptureException(out.err)
		})
	}

	args := append([]any{"provider", out.source.ID, "outcome", outcome, "latency", out.latency.Milliseconds()}, extra...)
	o.log.Log(context.Background(), lvl, "provider result", args...)

	if o.metrics != nil {
		o.metrics.ProviderOps.WithLabelValues(out.source.ID, outcome).Inc()
		o.metrics.ProviderLatency.WithLabelValues(out.source.ID).Observe(out.latency.Seconds())
	}

	return outcome
}

func isNetworkNoise(err error) bool {
	if _, ok := errors.AsType[*net.OpError](err); ok {
		return true
	}
	if ne, ok := errors.AsType[net.Error](err); ok && ne.Timeout() {
		return true
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	if _, ok := errors.AsType[*tls.CertificateVerificationError](err); ok {
		return true
	}
	return false
}
