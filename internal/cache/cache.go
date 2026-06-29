package cache

import (
	"context"
	"time"
)

type Cache interface {
	GetBytes(ctx context.Context, key string) ([]byte, bool, error)
	GetManyBytes(ctx context.Context, keys ...string) ([][]byte, error)
	SetBytes(ctx context.Context, key string, data []byte, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	TTLMany(ctx context.Context, keys ...string) ([]time.Duration, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	Incr(ctx context.Context, key string, ttl time.Duration) (int64, error)
}
