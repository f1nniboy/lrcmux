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
	"golang.org/x/sync/singleflight"

	"github.com/f1nniboy/lrcmux/internal/cache"
	"github.com/f1nniboy/lrcmux/internal/config"
	"github.com/f1nniboy/lrcmux/internal/isrc"
	"github.com/f1nniboy/lrcmux/internal/logging"
	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/metrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

var (
	ErrNoProviders = errors.New("no providers available")
	ErrNotFound    = errors.New("no lyrics found")
)

type Orchestrator struct {
	providers []providers.Provider
	cache     cache.Cache
	breaker   *Breaker
	resolver  *isrc.Resolver
	opts      Options
	log       *slog.Logger
	sf        singleflight.Group
	metrics   *metrics.Collector
}

type Request struct {
	Artist    string
	Title     string
	Album     string
	Duration  int64
	ISRC      string
	Level     lyrics.SyncLevel
	Strict    bool
	FetchMode string // "default", "cache", "force"
	Charge    func(ctx context.Context) error
}

type Response struct {
	Result *lyrics.Result
	Cached bool
	TTL    time.Duration
}

type ProviderHealth struct {
	Ok     bool   `json:"ok"`
	TTL    int64  `json:"ttl,omitempty"`
	Reason string `json:"reason,omitempty"`
}

type ProviderInfo struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Health ProviderHealth `json:"health"`
}

type Options struct {
	Timeout time.Duration
	TTL     config.CacheTTL
}

func New(provs []providers.Provider, c cache.Cache, breaker *Breaker, resolver *isrc.Resolver, coll *metrics.Collector, opts Options) *Orchestrator {
	return &Orchestrator{
		providers: provs,
		cache:     c,
		breaker:   breaker,
		resolver:  resolver,
		metrics:   coll,
		opts:      opts,
		log:       logging.New("orchestrator"),
	}
}

func (o *Orchestrator) Providers() []providers.Provider { return o.providers }

func (o *Orchestrator) recordOutcome(stage string) {
	if o.metrics != nil {
		o.metrics.RequestOutcomes.WithLabelValues(stage).Inc()
	}
}

func (o *Orchestrator) Health(ctx context.Context) []ProviderInfo {
	disabled := o.breaker.states(ctx, o.providers)
	infos := make([]ProviderInfo, len(o.providers))
	for i, p := range o.providers {
		health := ProviderHealth{Ok: true}
		if open, ok := disabled[p.ID()]; ok {
			health.Ok = false
			health.TTL = int64(open.TTL.Seconds())
			health.Reason = open.Reason
		}
		infos[i] = ProviderInfo{ID: p.ID(), Name: p.Name(), Health: health}
	}
	return infos
}

type providerOutcome struct {
	id      string
	name    string
	result  *lyrics.Result
	err     error
	latency time.Duration
}

func (o *Orchestrator) Get(ctx context.Context, req Request) (*Response, error) {
	if len(o.providers) == 0 {
		return nil, ErrNoProviders
	}

	track, err := o.resolver.Resolve(ctx, isrc.ResolveInput{
		Artist:    req.Artist,
		Title:     req.Title,
		Album:     req.Album,
		Duration:  req.Duration,
		ISRC:      req.ISRC,
		CacheOnly: req.FetchMode == "cache",
	})
	if err != nil {
		o.recordOutcome("isrc_not_found")
		if req.Charge != nil {
			if err := req.Charge(ctx); err != nil {
				return nil, err
			}
		}
		return nil, ErrNotFound
	}

	q := lyrics.Query{Track: track}

	best, unknowns := o.checkCache(ctx, q, req.FetchMode == "force")

	if best != nil && best.SyncLevel >= req.Level {
		o.log.Debug("serving from cache", "provider", best.Source.ID, "sync", best.SyncLevel.String())
		o.recordOutcome("cache_hit")
		return o.respond(ctx, best, true, req.Level, q), nil
	}

	// don't hit providers in cache-only mode
	if req.FetchMode == "cache" {
		if !req.Strict && best != nil {
			o.recordOutcome("cache_hit")
			return o.respond(ctx, best, true, req.Level, q), nil
		}
		return nil, ErrNotFound
	}

	// filter out providers that are currently disabled
	unknowns = o.breaker.Filter(ctx, unknowns)

	// drop providers that can never satisfy the requested level
	if req.Strict {
		var capable []providers.Provider
		for _, p := range unknowns {
			if p.MaxLevel() >= req.Level {
				capable = append(capable, p)
			}
		}
		unknowns = capable
	}

	// no providers left to query: all were cached misses, breaker-open, or filtered by strict
	if len(unknowns) == 0 {
		if !req.Strict && best != nil {
			o.log.Debug("no providers to query, serving best available from cache", "provider", best.Source.ID, "sync", best.SyncLevel.String())
			return o.respond(ctx, best, true, req.Level, q), nil
		}
		return nil, ErrNotFound
	}

	if req.Charge != nil {
		if err := req.Charge(ctx); err != nil {
			return nil, err
		}
	}

	v, err, _ := o.sf.Do(queryKey(q, req.Level), func() (any, error) {
		o.recordOutcome("fanout")

		var allResults []*lyrics.Result
		for i, tier := range buildTiers(unknowns, req.Level) {
			if len(tier) == 0 {
				continue
			}
			o.log.Debug("fanout tier", "tier", i, "providers", providerIDs(tier), "target_level", req.Level.String())
			allResults = append(allResults, o.fanOut(ctx, tier, q, req.Level)...)

			if picked := o.pick(allResults, req.Level); picked != nil {
				o.log.Debug("tier satisfied", "tier", i, "provider", picked.Source.ID, "sync", picked.SyncLevel.String())
				return o.respond(ctx, picked, false, req.Level, q), nil
			}
			o.log.Debug("tier exhausted", "tier", i, "collected", len(allResults))
			if len(allResults) > 0 {
				// we have results but can't meet the level
				// remaining tiers are worse, so they won't improve on what we have
				break
			}
		}
		if !req.Strict {
			if best != nil {
				allResults = append(allResults, best)
			}
			if len(allResults) > 0 {
				if picked := o.pick(allResults, lyrics.SyncNone); picked != nil {
					o.log.Debug("pick: falling back to best available", "provider", picked.Source.ID, "sync", picked.SyncLevel.String())
					return o.respond(ctx, picked, false, req.Level, q), nil
				}
			}
		}
		return nil, ErrNotFound
	})
	if err != nil {
		return nil, err
	}
	return v.(*Response), nil
}

func (o *Orchestrator) checkCache(ctx context.Context, q lyrics.Query, force bool) (best *lyrics.Result, unknowns []providers.Provider) {
	if force {
		return nil, append([]providers.Provider(nil), o.providers...)
	}

	keys := make([]string, len(o.providers))
	for i, p := range o.providers {
		keys[i] = cacheKey(q.Track.ISRC, p.ID())
	}
	results, statuses, err := cache.GetMany[lyrics.Result](ctx, o.cache, keys)
	if err != nil {
		o.log.Warn("cache get failed", "err", err)
		return nil, append([]providers.Provider(nil), o.providers...)
	}
	for i, p := range o.providers {
		switch statuses[i] {
		case cache.Found:
			r := results[i]
			o.log.Debug("cache hit", "provider", p.ID(), "sync", r.SyncLevel.String())
			if best == nil || rankResult(&r, best) {
				best = &results[i]
			}
		case cache.KnownMiss:
			o.log.Debug("cached miss", "provider", p.ID())
		default:
			o.log.Debug("cache miss", "provider", p.ID())
			unknowns = append(unknowns, p)
		}
	}
	return
}

func (o *Orchestrator) respond(ctx context.Context, r *lyrics.Result, cached bool, level lyrics.SyncLevel, q lyrics.Query) *Response {
	out := lyrics.Downgrade(r, level)
	out.Track = q.Track

	ttl := o.ttlFor(r)
	if cached && ttl > 0 {
		if cacheTTL, err := o.cache.TTL(ctx, cacheKey(q.Track.ISRC, r.Source.ID)); err == nil {
			ttl = cacheTTL
		}
	}
	if ttl == 0 {
		// pretty arbitrary, but we always want some sane value for
		// Cache-Control header
		ttl = 365 * 24 * time.Hour
	}

	return &Response{Result: out, Cached: cached, TTL: ttl}
}

func (o *Orchestrator) fanOut(ctx context.Context, active []providers.Provider, q lyrics.Query, level lyrics.SyncLevel) []*lyrics.Result {
	fanCtx, cancel := context.WithTimeout(ctx, o.opts.Timeout)
	defer cancel()

	o.log.Debug("fanning out", "providers", providerIDs(active), "target_level", level.String(), "timeout", o.opts.Timeout.Milliseconds())

	ch := make(chan providerOutcome, len(active))
	var wg sync.WaitGroup
	for _, p := range active {
		wg.Add(1)
		go func(p providers.Provider) {
			defer wg.Done()
			o.log.Debug("querying provider", "provider", p.ID())
			start := time.Now()
			r, err := p.Search(fanCtx, q)
			ch <- providerOutcome{id: p.ID(), name: p.Name(), result: r, err: err, latency: time.Since(start)}
		}(p)
	}
	go func() { wg.Wait(); close(ch) }()

	var results []*lyrics.Result
	var misses []string

	collect := func(out providerOutcome) {
		if out.err == nil && out.result != nil {
			out.result.Source = lyrics.Source{ID: out.id, Name: out.name}
			out.result.Lines = lyrics.CleanLines(out.result.Lines)
			results = append(results, out.result)
		}
		if errors.Is(out.err, lyrics.ErrNotFound) {
			misses = append(misses, out.id)
		}
		outcome := o.logOutcome(out, q)
		o.breaker.Record(out.id, outcome)
	}

	for out := range ch {
		collect(out)
		if out.result != nil && out.result.SyncLevel >= level {
			o.log.Debug("target satisfied, cancelling remaining", "provider", out.id, "sync", out.result.SyncLevel.String())
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
		lvl     slog.Level = slog.LevelDebug
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
			scope.SetTag("provider", out.id)
			scope.SetContext("query", sentry.Context{
				"isrc":   q.Track.ISRC,
				"artist": q.Track.Artist,
				"title":  q.Track.Title,
			})
			sentry.CaptureException(out.err)
		})
	}

	args := append([]any{"provider", out.id, "outcome", outcome, "latency", out.latency.Milliseconds()}, extra...)
	o.log.Log(context.Background(), lvl, "provider result", args...)

	if o.metrics != nil {
		o.metrics.ProviderOps.WithLabelValues(out.id, outcome).Inc()
		o.metrics.ProviderLatency.WithLabelValues(out.id).Observe(out.latency.Seconds())
	}

	return outcome
}

func (o *Orchestrator) pick(results []*lyrics.Result, level lyrics.SyncLevel) *lyrics.Result {
	var best *lyrics.Result
	for _, r := range results {
		if r.SyncLevel >= level && (best == nil || rankResult(r, best)) {
			best = r
		}
	}
	if best != nil {
		o.log.Debug("pick selected", "provider", best.Source.ID, "sync", best.SyncLevel.String())
	} else {
		o.log.Debug("pick: no result meets target level")
	}
	return best
}

func (o *Orchestrator) ttlFor(r *lyrics.Result) time.Duration {
	switch r.SyncLevel {
	case lyrics.SyncWord:
		return o.opts.TTL.Word.Duration
	case lyrics.SyncLine:
		return o.opts.TTL.Line.Duration
	default:
		return o.opts.TTL.None.Duration
	}
}

func rankResult(a, b *lyrics.Result) bool {
	if a.SyncLevel != b.SyncLevel {
		return a.SyncLevel > b.SyncLevel
	}
	return len(a.Lines) > len(b.Lines)
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
