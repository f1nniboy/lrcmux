package orchestrator

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"time"

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
	cache     cache.Cache
	sf        singleflight.Group
	breaker   *Breaker
	resolver  *isrc.Resolver
	log       *slog.Logger
	metrics   *metrics.Collector
	providers []providers.Provider
	opts      Options
}

type Request struct {
	Charge func(ctx context.Context) error
	Artist string
	Title  string
	Album  string
	ISRC   string
	// "default", "cache", "force"
	FetchMode string
	Sources   []string
	Duration  int64
	Level     lyrics.SyncLevel
	Strict    bool
}

type Response struct {
	Result *lyrics.Result
	Cached bool
	TTL    time.Duration
}

type ProviderHealth struct {
	Reason string `json:"reason,omitempty"`
	TTL    int64  `json:"ttl,omitempty"`
	Ok     bool   `json:"ok"`
}

type ProviderInfo struct {
	lyrics.Source
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
		infos[i] = ProviderInfo{Source: providers.Source(p), Health: health}
	}
	return infos
}

func (o *Orchestrator) Get(ctx context.Context, req Request) (*Response, error) {
	if len(o.providers) == 0 {
		return nil, ErrNoProviders
	}

	active, err := filterBySources(o.providers, req.Sources)
	if err != nil {
		return nil, err
	}
	if len(active) == 0 {
		return nil, ErrNoProviders
	}

	track, err := o.resolver.Resolve(ctx, isrc.ResolveInput{
		Artist:   req.Artist,
		Title:    req.Title,
		Album:    req.Album,
		Duration: req.Duration,
		ISRC:     req.ISRC,
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

	respond := func(r *lyrics.Result, cached bool) *Response {
		out := r.Downgrade(req.Level)
		out.Track = q.Track

		ttl := o.ttlFor(r)
		if cached && ttl > 0 {
			if cacheTTL, err := o.cache.TTL(ctx, cacheKey(q.Track.ISRC, r.Source.ID)); err == nil {
				ttl = cacheTTL
			}
		}
		if ttl == 0 {
			// a bit arbitrary, but we always want some TTL for
			// the Cache-Control header
			ttl = 365 * 24 * time.Hour
		}

		return &Response{Result: out, Cached: cached, TTL: ttl}
	}

	cached, unknowns := o.getCacheAndProviders(ctx, q, req, active)

	// in strict mode we must meet the level,
	// otherwise we pick the best across levels
	pickLevel := lyrics.SyncNone
	if req.Strict {
		pickLevel = req.Level
	}

	// serve directly from cache when nothing worth querying remains
	//
	// worthQuerying already dropped providers that can't improve on cache,
	// so if unknowns is empty, the pick here is as good as we can get
	if req.FetchMode == "cache" || len(unknowns) == 0 {
		if picked := o.pick(cached, pickLevel); picked != nil {
			o.log.Debug("serving from cache", "provider", picked.Source.ID, "sync", picked.SyncLevel.String())
			o.recordOutcome("cache_hit")
			return respond(picked, true), nil
		}
		return nil, ErrNotFound
	}

	if req.Charge != nil {
		if err := req.Charge(ctx); err != nil {
			return nil, err
		}
	}

	key := q.Track.ISRC + ":" + req.Level.String()
	if len(req.Sources) > 0 {
		// order of sources shouldn't matter
		sorted := slices.Clone(req.Sources)
		slices.Sort(sorted)
		key += ":" + strings.Join(sorted, ",")
	}

	v, err, _ := o.sf.Do(key, func() (any, error) {
		o.recordOutcome("fanout")

		// seed with cached so pick considers them alongside fresh results for
		// each tier, skipped in strict since it will never fall back
		var results []*lyrics.Result
		if !req.Strict {
			results = append(results, cached...)
		}

		for i, tier := range buildTiers(unknowns, req.Level) {
			o.log.Debug("fanout tier", "tier", i, "providers", providers.IDs(tier), "target_level", req.Level.String())
			results = append(results, o.fanOut(ctx, tier, q, req.Level)...)

			if picked := o.pick(results, pickLevel); picked != nil {
				o.log.Debug("tier satisfied", "tier", i, "provider", picked.Source.ID, "level", picked.SyncLevel.String())
				return respond(picked, slices.Contains(cached, picked)), nil
			}
			o.log.Debug("tier exhausted", "tier", i, "collected", len(results))
		}
		return nil, ErrNotFound
	})
	if err != nil {
		return nil, err
	}
	return v.(*Response), nil
}

// returns cached results and the providers still worth querying,
// after applying force/cache mode, breaker filter, strict, and the
// "must beat what's already cached" filter
func (o *Orchestrator) getCacheAndProviders(ctx context.Context, q lyrics.Query, req Request, provs []providers.Provider) (cached []*lyrics.Result, unknowns []providers.Provider) {
	if req.FetchMode == "force" {
		return nil, slices.Clone(provs)
	}

	cached, unknowns = o.checkCache(ctx, q, provs)
	if req.FetchMode == "cache" {
		return cached, nil
	}

	unknowns = o.breaker.Filter(ctx, unknowns)
	return cached, worthQuerying(unknowns, cached, req)
}

// drops providers that can't improve on what's already cached
// and, in strict mode, those that can't satisfy the requested level
func worthQuerying(unknowns []providers.Provider, cached []*lyrics.Result, req Request) []providers.Provider {
	var bestCachedLevel lyrics.SyncLevel
	for _, c := range cached {
		if c.SyncLevel > bestCachedLevel {
			bestCachedLevel = c.SyncLevel
		}
	}
	return slices.DeleteFunc(unknowns, func(p providers.Provider) bool {
		if len(cached) > 0 && p.MaxLevel() <= bestCachedLevel {
			return true
		}
		return req.Strict && p.MaxLevel() < req.Level
	})
}

func (o *Orchestrator) checkCache(ctx context.Context, q lyrics.Query, provs []providers.Provider) (hits []*lyrics.Result, unknowns []providers.Provider) {
	keys := make([]string, len(provs))
	for i, p := range provs {
		keys[i] = cacheKey(q.Track.ISRC, p.ID())
	}
	results, statuses, err := cache.GetMany[lyrics.Result](ctx, o.cache, keys)
	if err != nil {
		o.log.Warn("cache get failed", "err", err)
		return nil, slices.Clone(provs)
	}
	for i, p := range provs {
		switch statuses[i] {
		case cache.Found:
			o.log.Debug("cache hit", "provider", p.ID(), "sync", results[i].SyncLevel.String())
			hits = append(hits, &results[i])
		case cache.KnownMiss:
			o.log.Debug("known miss", "provider", p.ID())
		default:
			o.log.Debug("cache miss", "provider", p.ID())
			unknowns = append(unknowns, p)
		}
	}
	return hits, unknowns
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

func (o *Orchestrator) recordOutcome(stage string) {
	if o.metrics != nil {
		o.metrics.RequestOutcomes.WithLabelValues(stage).Inc()
	}
}
