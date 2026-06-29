package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type LimitError struct {
	RetryAfter time.Duration
}

func (e *LimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded, retry after %s", e.RetryAfter)
}

func (e *LimitError) Unwrap() error { return ErrRateLimited }

var ErrRateLimited = errors.New("rate limit exceeded")

// checks the current count before incrementing so no rollback is ever needed
// 0 if denied, 1 if allowed
var allowScript = redis.NewScript(`
local count = tonumber(redis.call("GET", KEYS[1]) or "0")
if count >= tonumber(ARGV[2]) then return 0 end
local n = redis.call("INCR", KEYS[1])
if n == 1 then redis.call("PEXPIRE", KEYS[1], tonumber(ARGV[1])) end
return 1
`)

type Option func(*Limiter)

func WithLogger(log *slog.Logger) Option {
	return func(l *Limiter) { l.log = log }
}

type Limiter struct {
	rdb    *redis.Client
	limit  int64
	window time.Duration
	log    *slog.Logger
}

func (l *Limiter) Limit() int64          { return l.limit }
func (l *Limiter) Window() time.Duration { return l.window }

func New(rdb *redis.Client, limit int64, window time.Duration, opts ...Option) *Limiter {
	l := &Limiter{rdb: rdb, limit: limit, window: window, log: slog.Default()}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

func (l *Limiter) Allow(ctx context.Context, ip string) error {
	bucket := time.Now().Unix() / int64(l.window.Seconds())
	key := "rl:" + ip + ":" + strconv.FormatInt(bucket, 10)
	res, err := allowScript.Run(ctx, l.rdb, []string{key}, l.window.Milliseconds(), l.limit).Int64()
	if err != nil {
		l.log.Warn("ratelimit check failed, allowing request", "err", err)
		return nil
	}
	if res == 0 {
		retryAfter := time.Until(time.Unix((bucket+1)*int64(l.window.Seconds()), 0))
		return &LimitError{RetryAfter: retryAfter}
	}
	return nil
}
