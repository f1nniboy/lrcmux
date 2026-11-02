package orchestrator

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/f1nniboy/lrcmux/internal/cache"
	"github.com/f1nniboy/lrcmux/internal/config"
	"github.com/f1nniboy/lrcmux/internal/isrc"
	"github.com/f1nniboy/lrcmux/internal/logging"
	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/metrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
	"github.com/f1nniboy/lrcmux/internal/proxy"
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

type FetchMode string

const (
	FetchDefault FetchMode = "default"
	FetchCache   FetchMode = "cache"
	FetchForce   FetchMode = "force"
)

type Request struct {
	Charge    func(ctx context.Context) error
	Artist    string
	Title     string
	Album     string
	ISRC      string
	FetchMode FetchMode
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

func (o *Orchestrator) OpenCircuits(ctx context.Context) map[string]string {
	return o.breaker.Tripped(ctx, o.providers)
}

func (o *Orchestrator) Get(ctx context.Context, req Request) (*Response, error) {
	if len(o.providers) == 0 {
		return nil, ErrNoProviders
	}

	active, err := filterBySources(o.providers, req.Sources)
	if err != nil {
		return nil, err
	}
	// strict never falls back, so providers below the level are useless
	if req.Strict {
		active = slices.DeleteFunc(slices.Clone(active), func(p providers.Provider) bool {
			return p.MaxLevel() < req.Level
		})
	}
	if len(active) == 0 {
		return nil, ErrNotFound
	}

	// charged once, before the first upstream request
	charge := sync.OnceValue(func() error {
		// don't start upstream work for a request that's already cancelled
		if err := ctx.Err(); err != nil {
			return err
		}
		if req.Charge == nil {
			return nil
		}
		return req.Charge(ctx)
	})

	in := isrc.ResolveInput{
		Artist:   req.Artist,
		Title:    req.Title,
		Album:    req.Album,
		Duration: req.Duration,
		ISRC:     req.ISRC,
	}
	track, hit, err := o.resolver.Cached(ctx, in)
	if !hit {
		if req.FetchMode == FetchCache {
			o.recordOutcome("cache")
			return nil, ErrNotFound
		}
		if err := charge(); err != nil {
			return nil, err
		}
		track, err = o.resolver.Resolve(ctx, in)
	}
	if err != nil {
		// only a miss is a cacheable 404
		outcome := "isrc_error"
		if errors.Is(err, lyrics.ErrNotFound) {
			outcome, err = "isrc_not_found", ErrNotFound
		}
		o.recordOutcome(outcome)
		return nil, err
	}

	q := lyrics.Query{Track: track}

	respond := func(r *lyrics.Result, cached bool) *Response {
		// a copy, since the fanout may still be caching r in the background
		out := *r.Downgrade(req.Level)
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

		return &Response{Result: &out, Cached: cached, TTL: ttl}
	}

	cached, unknowns := o.getCacheAndProviders(ctx, q, req, active)

	// in strict mode we must meet the level, otherwise we pick
	// the best across levels
	pickLevel := lyrics.SyncNone
	if req.Strict {
		pickLevel = req.Level
	}

	// serve directly from cache when nothing worth querying remains
	//
	// worthQuerying already dropped providers that can't improve on cache,
	// so if unknowns is empty, the pick here is as good as we can get
	if req.FetchMode == FetchCache || len(unknowns) == 0 {
		o.recordOutcome("cache")
		if picked := o.pick(cached, pickLevel); picked != nil {
			o.log.Debug("serving from cache", "provider", picked.Source.ID, "sync", picked.SyncLevel.String())
			return respond(picked, true), nil
		}
		return nil, ErrNotFound
	}

	if err := charge(); err != nil {
		return nil, err
	}

	key := q.Track.ISRC + ":" + req.Level.String()
	if req.Strict {
		key += ":strict"
	}
	// a forced fetch can't take a pick seeded from the cache
	if req.FetchMode == FetchForce {
		key += ":force"
	}
	if len(req.Sources) > 0 {
		// order of sources shouldn't matter
		sorted := slices.Clone(req.Sources)
		slices.Sort(sorted)
		key += ":" + strings.Join(sorted, ",")
	}

	// counted per request, since followers of the flight below needed a fanout too
	o.recordOutcome("fanout")
	v, err, _ := o.sf.Do(key, func() (any, error) {
		// so the leader leaving can't cancel it for the others
		ctx, end := proxy.WithSession(context.WithoutCancel(ctx))
		defer end()

		// seed with cached so pick weighs them against fresh results each tier
		results := slices.Clone(cached)

		tiers := buildTiers(unknowns, req.Level)
		for i, tier := range tiers {
			o.log.Debug("fanout tier", "tier", i, "providers", providers.IDs(tier), "target_level", req.Level.String())
			results = append(results, o.fanOut(ctx, tier, q, req.Level)...)

			last := i == len(tiers)-1
			if picked := o.pick(results, pickLevel); picked != nil && (last || satisfies(picked, maxLevel(tiers[i+1]))) {
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
// after applying force/cache mode, breaker filter, and the
// "must beat what's already cached" filter
func (o *Orchestrator) getCacheAndProviders(ctx context.Context, q lyrics.Query, req Request, provs []providers.Provider) (cached []*lyrics.Result, unknowns []providers.Provider) {
	if req.FetchMode == FetchForce {
		return nil, slices.Clone(provs)
	}

	cached, unknowns = o.checkCache(ctx, q, provs)
	if req.FetchMode == FetchCache {
		return cached, nil
	}

	unknowns = o.breaker.Filter(ctx, unknowns)
	return cached, worthQuerying(unknowns, cached)
}

// drops providers that can't improve on what's already cached
func worthQuerying(unknowns []providers.Provider, cached []*lyrics.Result) []providers.Provider {
	// rankResult prefers clean over level, so best captures both dimensions:
	// a clean result at any level beats a censored result at a higher level
	var best *lyrics.Result
	for _, c := range cached {
		if best == nil || rankResult(c, best) > 0 {
			best = c
		}
	}
	return slices.DeleteFunc(unknowns, func(p providers.Provider) bool {
		// once the cached result is satisfying, only a higher sync level can
		// improve on it
		return best != nil && satisfies(best, best.SyncLevel) && p.MaxLevel() <= best.SyncLevel
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
		case cache.Hit:
			// entries can predate a cleanup rule, and word-level ones never expire
			if !results[i].Clean() {
				o.log.Debug("cached result empty after cleanup", "provider", p.ID())
				continue
			}
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
	if r.Instrumental {
		return o.opts.TTL.Instrumental.Duration
	}
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
