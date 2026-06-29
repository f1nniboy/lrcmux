package orchestrator

import (
	"context"
	"errors"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"golang.org/x/sync/singleflight"

	"github.com/f1nniboy/lrcmux/internal/cache"
	"github.com/f1nniboy/lrcmux/internal/isrc"
	"github.com/f1nniboy/lrcmux/internal/logging"
	"github.com/f1nniboy/lrcmux/internal/lyrics"
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
}

type Request struct {
	Query  lyrics.Query
	Level  lyrics.SyncLevel
	Strict bool
	Force  bool
	Charge func(ctx context.Context) error
}

type Response struct {
	Result *lyrics.Result
	Cached bool
	Track  lyrics.Track
}

type ProviderHealth struct {
	Available bool   `json:"available"`
	TTL       int64  `json:"ttl,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

// ProviderInfo describes a provider and its current circuit breaker state.
type ProviderInfo struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Health ProviderHealth `json:"health"`
}

type Options struct {
	Timeout      time.Duration
	CacheTTL     time.Duration
	CacheMissTTL time.Duration
}

func New(provs []providers.Provider, c cache.Cache, breaker *Breaker, resolver *isrc.Resolver, opts Options) *Orchestrator {
	return &Orchestrator{
		providers: provs,
		cache:     c,
		breaker:   breaker,
		resolver:  resolver,
		opts:      opts,
		log:       logging.New("orchestrator"),
	}
}

func (o *Orchestrator) Providers() []providers.Provider { return o.providers }

func (o *Orchestrator) ProviderInfos(ctx context.Context) []ProviderInfo {
	disabled := o.breaker.states(ctx, o.providers)
	infos := make([]ProviderInfo, len(o.providers))
	for i, p := range o.providers {
		health := ProviderHealth{Available: true}
		if open, ok := disabled[p.ID()]; ok {
			health.Available = false
			health.TTL = int64(open.TTL.Seconds())
			health.Reason = open.State.Reason
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

	if req.Query.ISRC == "" {
		isrc, err := o.resolver.Resolve(ctx, req.Query)
		if err != nil {
			return nil, ErrNotFound
		}
		req.Query.ISRC = isrc
	}
	if meta, ok := o.resolver.LookupMeta(ctx, req.Query.ISRC); ok {
		req.Query.Artist = meta.Artist
		req.Query.Title = meta.Title
		req.Query.Album = meta.Album
		req.Query.Duration = meta.Duration
	}

	best, unknowns := o.checkCache(ctx, req.Query, req.Force)

	if best != nil && best.SyncLevel >= req.Level {
		o.log.Debug("serving from cache", "provider", best.Source.ID, "sync", best.SyncLevel.String())
		return respond(best, true, req.Level, req.Query), nil
	}

	unknowns = o.breaker.Filter(ctx, unknowns)

	if len(unknowns) == 0 {
		if !req.Strict && best != nil {
			o.log.Debug("all providers explored, serving best available from cache", "provider", best.Source.ID, "sync", best.SyncLevel.String())
			return respond(best, true, req.Level, req.Query), nil
		}
		return nil, ErrNotFound
	}

	sfKey := queryKey(req.Query)
	type fanResult struct{ resp *Response }
	v, err, _ := o.sf.Do(sfKey, func() (any, error) {
		if req.Charge != nil {
			if err := req.Charge(ctx); err != nil {
				return nil, err
			}
		}

		results := o.fanOut(ctx, unknowns, req.Query, req.Level)

		picked := o.pick(results, req.Level)
		if picked == nil && !req.Strict {
			all := results
			if best != nil {
				all = append(all, best)
			}
			picked = o.pick(all, lyrics.SyncNone)
			if picked != nil {
				o.log.Debug("pick: falling back to best available", "provider", picked.Source.ID, "sync", picked.SyncLevel.String())
			}
		}
		if picked == nil {
			return nil, ErrNotFound
		}

		return &fanResult{respond(picked, false, req.Level, req.Query)}, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(*fanResult).resp, nil
}

func (o *Orchestrator) checkCache(ctx context.Context, q lyrics.Query, force bool) (best *lyrics.Result, unknowns []providers.Provider) {
	if force {
		return nil, append([]providers.Provider(nil), o.providers...)
	}

	keys := make([]string, len(o.providers))
	for i, p := range o.providers {
		keys[i] = cacheKey(q.ISRC, p.ID())
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

func respond(r *lyrics.Result, cached bool, level lyrics.SyncLevel, q lyrics.Query) *Response {
	return &Response{
		Result: lyrics.Downgrade(r, level),
		Cached: cached,
		Track: lyrics.Track(q),
	}
}

func (o *Orchestrator) fanOut(ctx context.Context, active []providers.Provider, q lyrics.Query, level lyrics.SyncLevel) []*lyrics.Result {
	fanCtx, cancel := context.WithTimeout(ctx, o.opts.Timeout)
	defer cancel()

	ids := make([]string, len(active))
	for i, p := range active {
		ids[i] = p.ID()
	}
	o.log.Debug("fanning out", "providers", ids, "target_level", level.String(), "timeout_ms", o.opts.Timeout.Milliseconds())

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
	var successes []string
	collect := func(out providerOutcome) {
		if out.err == nil && out.result != nil {
			out.result.Source = lyrics.Source{ID: out.id, Name: out.name}
			out.result.Lines = lyrics.CleanLines(out.result.Lines)
			results = append(results, out.result)
			successes = append(successes, out.id)
		}
		if errors.Is(out.err, lyrics.ErrNotFound) {
			misses = append(misses, out.id)
		}
		o.logOutcome(out, q)
		o.breaker.Record(out.id, out.err)
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
		for _, r := range results {
			if err := cache.Set(bg, o.cache, cacheKey(q.ISRC, r.Source.ID), *r, o.opts.CacheTTL); err != nil {
				o.log.Warn("cache set failed", "err", err, "provider", r.Source.ID)
			}
		}
		for _, provider := range misses {
			if err := cache.SetMiss(bg, o.cache, cacheKey(q.ISRC, provider), o.opts.CacheMissTTL); err != nil {
				o.log.Warn("miss cache set failed", "provider", provider, "err", err)
			}
		}
		if len(successes) > 0 {
			streakKeys := make([]string, len(successes))
			for i, id := range successes {
				streakKeys[i] = "cb:" + id + ":streak"
			}
			o.cache.Delete(bg, streakKeys...)
		}
	}()

	return results
}

func (o *Orchestrator) logOutcome(out providerOutcome, q lyrics.Query) {
	switch {
	case out.err == nil && out.result != nil:
		o.log.Debug("provider ok", "provider", out.id, "sync", out.result.SyncLevel.String(), "latency_ms", out.latency.Milliseconds())
	case errors.Is(out.err, lyrics.ErrNotFound):
		o.log.Debug("provider not found", "provider", out.id, "latency_ms", out.latency.Milliseconds())
	case errors.Is(out.err, providers.ErrRateLimited):
		o.log.Info("provider rate limited", "provider", out.id, "latency_ms", out.latency.Milliseconds())
	case errors.Is(out.err, context.DeadlineExceeded) || errors.Is(out.err, lyrics.ErrTimeout):
		o.log.Info("provider timeout", "provider", out.id, "latency_ms", out.latency.Milliseconds())
	case errors.Is(out.err, context.Canceled):
		o.log.Debug("provider cancelled", "provider", out.id, "latency_ms", out.latency.Milliseconds())
	default:
		o.log.Warn("provider error", "provider", out.id, "err", out.err, "latency_ms", out.latency.Milliseconds())
		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetTag("provider", out.id)
			scope.SetContext("query", sentry.Context{
				"isrc":   q.ISRC,
				"artist": q.Artist,
				"title":  q.Title,
			})
			sentry.CaptureException(out.err)
		})
	}
}

// rankResult reports whether a is a better result than b.
// Primary criterion: sync level. Secondary: line count (fewer lines may indicate truncated lyrics).
func rankResult(a, b *lyrics.Result) bool {
	if a.SyncLevel != b.SyncLevel {
		return a.SyncLevel > b.SyncLevel
	}
	return len(a.Lines) > len(b.Lines)
}

func (o *Orchestrator) pick(results []*lyrics.Result, level lyrics.SyncLevel) *lyrics.Result {
	if len(results) == 0 {
		return nil
	}
	sort.SliceStable(results, func(i, j int) bool {
		return rankResult(results[i], results[j])
	})

	for i, r := range results {
		o.log.Debug("pick candidate", "rank", i+1, "provider", r.Source.ID, "sync", r.SyncLevel.String())
	}

	for _, r := range results {
		if r.SyncLevel >= level {
			o.log.Debug("pick selected", "provider", r.Source.ID, "sync", r.SyncLevel.String())
			return r
		}
	}
	o.log.Debug("pick: no result meets target level")
	return nil
}
