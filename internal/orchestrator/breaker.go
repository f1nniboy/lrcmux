package orchestrator

import (
	"context"
	"log/slog"
	"time"

	"github.com/f1nniboy/lrcmux/internal/cache"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

const (
	breakerThreshold int64         = 5
	breakerTTL       time.Duration = 3 * time.Minute
)

type breakerState struct {
	TTL    time.Duration
	Reason string
}

type Breaker struct {
	cache cache.Cache
	log   *slog.Logger
}

func NewBreaker(c cache.Cache, log *slog.Logger) *Breaker {
	return &Breaker{cache: c, log: log}
}

func (b *Breaker) states(ctx context.Context, provs []providers.Provider) map[string]breakerState {
	keys := make([]string, len(provs))
	for i, p := range provs {
		keys[i] = "cb:" + p.ID()
	}
	ttls, err := b.cache.TTLMany(ctx, keys...)
	if err != nil {
		return nil
	}
	vals, _, err := cache.GetMany[breakerState](ctx, b.cache, keys)
	if err != nil {
		return nil
	}
	out := make(map[string]breakerState, len(provs))
	for i, p := range provs {
		if ttls[i] <= 0 {
			continue
		}
		out[p.ID()] = breakerState{TTL: ttls[i], Reason: vals[i].Reason}
	}
	return out
}

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

func (b *Breaker) ResetStreak(ctx context.Context, providers []string) {
	if len(providers) == 0 {
		return
	}
	keys := make([]string, len(providers))
	for i, id := range providers {
		keys[i] = "cb:" + id + ":streak"
	}
	b.cache.Delete(ctx, keys...)
}

func (b *Breaker) Record(provider, outcome string) {
	switch outcome {
	case "ok", "not_found", "canceled":
		// caller handles streak reset via ResetStreak

	default:
		ctx := context.Background()
		streakKey := "cb:" + provider + ":streak"
		n, err := b.cache.Incr(ctx, streakKey, breakerTTL)
		if err != nil {
			b.log.Warn("breaker increment failed", "provider", provider, "err", err)
			return
		}
		if n >= breakerThreshold {
			cache.Set(ctx, b.cache, "cb:"+provider, breakerState{Reason: outcome}, breakerTTL)
			b.cache.Delete(ctx, streakKey)
			b.log.Debug("circuit opened", "provider", provider, "streak", n, "reason", outcome)
		}
	}
}
