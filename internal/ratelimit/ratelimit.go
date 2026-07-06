package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/prometheus/client_golang/prometheus"
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

// returns {wait_ms, count}:
// - wait_ms=0 on allow, positive on deny
// - count is current window count
var allowScript = redis.NewScript(`
local now = tonumber(ARGV[1])
local window_ms = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])

redis.call("ZREMRANGEBYSCORE", KEYS[1], "-inf", now - window_ms)

local count = redis.call("ZCARD", KEYS[1])
if count >= limit then
    local oldest = tonumber(redis.call("ZRANGE", KEYS[1], 0, 0, "WITHSCORES")[2])
    return {oldest + window_ms - now, count}
end

redis.call("ZADD", KEYS[1], now, now .. "-" .. math.random())
redis.call("PEXPIRE", KEYS[1], window_ms)

return {0, count + 1}
`)

const (
	penaltyThreshold = 5
	maxStrikes       = 8 // basePenalty<<maxStrikes-1, so max 256m penalty
	basePenalty      = time.Minute
	maxPenalty       = basePenalty * (1 << (maxStrikes - 1))
)

type Limiter struct {
	rdb       *redis.Client
	limit     int64
	window    time.Duration
	log       *slog.Logger
	denied    *prometheus.CounterVec
	penalized prometheus.Counter
}

func (l *Limiter) Limit() int64          { return l.limit }
func (l *Limiter) Window() time.Duration { return l.window }

func New(rdb *redis.Client, limit int64, window time.Duration, log *slog.Logger, denied *prometheus.CounterVec, penalized prometheus.Counter) *Limiter {
	return &Limiter{rdb: rdb, limit: limit, window: window, log: log, denied: denied, penalized: penalized}
}

// returns remaining requests in current window on success,
// or a LimitError if the request should be denied
func (l *Limiter) Allow(ctx context.Context, ip string) (int64, error) {
	penaltyKey := "rl:penalty:" + ip
	ttl, err := l.rdb.TTL(ctx, penaltyKey).Result()
	if err == nil && ttl > 0 {
		l.log.Debug("ip in penalty box", "ip", ip, "retry_after", ttl)
		if l.denied != nil {
			l.denied.WithLabelValues("penalty").Inc()
		}

		// this may or may not be a good idea
		delay := time.Duration(rand.Float64() * float64(3*time.Second))
		select {
		case <-time.After(delay):
		case <-ctx.Done():
		}

		return 0, &LimitError{RetryAfter: ttl}
	}

	key := "rl:window:" + ip
	raw, err := allowScript.Run(ctx, l.rdb, []string{key},
		time.Now().UnixMilli(), l.window.Milliseconds(), l.limit).Result()
	if err != nil {
		l.log.Warn("ratelimit check failed, allowing request", "err", err)
		return l.limit, nil
	}
	arr := raw.([]any)
	waitMs := arr[0].(int64)
	count := arr[1].(int64)
	remaining := l.limit - count

	if waitMs > 0 {
		retryAfter := time.Duration(waitMs) * time.Millisecond
		l.log.Debug("sliding window exceeded", "ip", ip, "retry_after", waitMs)
		if l.denied != nil {
			l.denied.WithLabelValues("window").Inc()
		}
		l.recordOffense(ctx, ip, retryAfter)
		return 0, &LimitError{RetryAfter: retryAfter}
	}
	return remaining, nil
}

func (l *Limiter) recordOffense(ctx context.Context, ip string, retryAfter time.Duration) {
	cooloffKey := "rl:cooloff:" + ip
	n, err := l.rdb.Incr(ctx, cooloffKey).Result()
	if err != nil {
		return
	}
	if n == 1 {
		l.rdb.PExpire(ctx, cooloffKey, retryAfter)
		l.log.Debug("cooloff started", "ip", ip, "retry_after", retryAfter)
		return
	}
	if n <= penaltyThreshold {
		l.log.Debug("offense in cooloff window", "ip", ip, "offense", n)
		return
	}

	strikeKey := "rl:strikes:" + ip
	strikes, err := l.rdb.Incr(ctx, strikeKey).Result()
	if err != nil {
		return
	}
	l.rdb.Expire(ctx, strikeKey, maxPenalty)

	penalty := basePenalty * (1 << (min(strikes, maxStrikes) - 1))
	l.rdb.Set(ctx, "rl:penalty:"+ip, 1, penalty)
	if l.penalized != nil {
		l.penalized.Inc()
	}
	l.log.Info("ip penalized", "ip", ip, "strikes", strikes, "penalty", penalty)
}
