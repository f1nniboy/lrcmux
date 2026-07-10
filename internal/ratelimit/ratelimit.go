package ratelimit

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

type LimitError struct {
	RetryAfter time.Duration
}

func (e *LimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded, retry after %ds", int(e.RetryAfter.Seconds()))
}

func (e *LimitError) Unwrap() error { return ErrRateLimited }

var ErrRateLimited = errors.New("rate limit exceeded")

type Limiter struct {
	limiter *redis_rate.Limiter
	log     *slog.Logger
	rate    redis_rate.Limit
	limit   int64
	window  time.Duration
}

func (l *Limiter) Limit() int64          { return l.limit }
func (l *Limiter) Window() time.Duration { return l.window }

func New(rdb *redis.Client, limit int64, window time.Duration, log *slog.Logger) *Limiter {
	return &Limiter{
		limiter: redis_rate.NewLimiter(rdb),
		rate:    redis_rate.Limit{Rate: int(limit), Burst: int(limit), Period: window},
		limit:   limit,
		window:  window,
		log:     log,
	}
}

func (l *Limiter) Charge(ctx context.Context, ip string) error {
	res, err := l.limiter.Allow(ctx, "rl:window:"+hashIP(ip), l.rate)
	if err != nil {
		l.log.Warn("ratelimit charge failed, allowing request", "err", err)
		return nil
	}
	if res.Allowed == 0 {
		return &LimitError{RetryAfter: res.RetryAfter}
	}
	return nil
}

func hashIP(ip string) string {
	sum := sha256.Sum256([]byte(ip))
	return fmt.Sprintf("%x", sum[:16])
}
