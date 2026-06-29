package orchestrator

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/f1nniboy/lrcmux/internal/cache"
	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

const (
	breakerStreakThreshold int64         = 3
	breakerStreakTTL       time.Duration = 3 * time.Minute
	breakerRateLimitTTL    time.Duration = 3 * time.Minute
	breakerErrorTTL        time.Duration = 3 * time.Minute
)

type breakerState struct {
	Reason string
}

type breakerOpen struct {
	TTL   time.Duration
	State breakerState
}

type Breaker struct {
	cache cache.Cache
	log   *slog.Logger
}

func NewBreaker(c cache.Cache, log *slog.Logger) *Breaker {
	return &Breaker{cache: c, log: log}
}

// returns a map of provider ID to open circuit info for providers whose
// circuit is currently open
func (b *Breaker) states(ctx context.Context, provs []providers.Provider) map[string]breakerOpen {
	keys := make([]string, len(provs))
	for i, p := range provs {
		keys[i] = "cb:" + p.ID()
	}
	ttls, err := b.cache.TTLMany(ctx, keys...)
	if err != nil {
		return nil
	}
	out := make(map[string]breakerOpen, len(provs))
	for i, p := range provs {
		if ttls[i] <= 0 {
			continue
		}
		open := breakerOpen{TTL: ttls[i]}
		if st, status, _ := cache.Get[breakerState](ctx, b.cache, keys[i]); status == cache.Found {
			open.State = st
		}
		out[p.ID()] = open
	}
	return out
}

// returns the subset of providers whose circuit is closed, skipping any
// provider whose circuit is currently open
func (b *Breaker) Filter(ctx context.Context, provs []providers.Provider) []providers.Provider {
	if len(provs) == 0 {
		return provs
	}
	keys := make([]string, len(provs))
	for i, p := range provs {
		keys[i] = "cb:" + p.ID()
	}
	raws, err := b.cache.GetManyBytes(ctx, keys...)
	if err != nil {
		return provs
	}
	var out []providers.Provider
	for i, p := range provs {
		if raws[i] != nil {
			b.log.Debug("circuit open, skipping provider", "provider", p.ID())
		} else {
			out = append(out, p)
		}
	}
	return out
}

func (b *Breaker) Record(provider string, err error) {
	ctx := context.Background()
	c := b.cache
	switch {
	case err == nil:
		// streak reset is done by the caller

	case errors.Is(err, providers.ErrRateLimited):
		cache.Set(ctx, c, "cb:"+provider, breakerState{Reason: "rate_limited"}, breakerRateLimitTTL)
		b.log.Debug("circuit opened (rate limited)", "provider", provider, "ttl", breakerRateLimitTTL)

	case errors.Is(err, lyrics.ErrNotFound), errors.Is(err, context.Canceled):
		// not a provider health issue

	default:
		streakKey := "cb:" + provider + ":streak"
		n, _ := c.Incr(ctx, streakKey, breakerStreakTTL)
		if n >= breakerStreakThreshold {
			cache.Set(ctx, c, "cb:"+provider, breakerState{Reason: "error_streak"}, breakerErrorTTL)
			c.Delete(ctx, streakKey)
			b.log.Debug("circuit opened (error streak)", "provider", provider, "streak", n, "ttl", breakerErrorTTL)
		}
	}
}
