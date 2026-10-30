package ratelimit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/netip"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"

	"github.com/f1nniboy/lrcmux/internal/config"
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
}

func (l *Limiter) Limit() int            { return l.rate.Rate }
func (l *Limiter) Burst() int            { return l.rate.Burst }
func (l *Limiter) Window() time.Duration { return l.rate.Period }

func New(rdb *redis.Client, rl config.RateLimit, log *slog.Logger) *Limiter {
	return &Limiter{
		limiter: redis_rate.NewLimiter(rdb),
		rate:    redis_rate.Limit{Rate: int(rl.Limit), Burst: int(rl.Burst), Period: rl.Window.Duration},
		log:     log,
	}
}

func (l *Limiter) Charge(ctx context.Context, ip string) error {
	// a charge that started finishes, since a cancelled one would fail open below
	res, err := l.limiter.Allow(context.WithoutCancel(ctx), "rl:window:"+clientKey(ip), l.rate)
	if err != nil {
		l.log.Warn("ratelimit charge failed, allowing request", "err", err)
		return nil
	}
	if res.Allowed == 0 {
		return &LimitError{RetryAfter: res.RetryAfter}
	}
	return nil
}

// an ipv6 client holds a whole /64, so it counts once
func clientKey(ip string) string {
	if addr, err := netip.ParseAddr(ip); err == nil {
		bits := 32
		if addr = addr.Unmap(); addr.Is6() {
			bits = 64
		}
		prefix, _ := addr.Prefix(bits)
		ip = prefix.String()
	}
	sum := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(sum[:16])
}
