package cache

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/f1nniboy/lrcmux/internal/logging"
)

type Redis struct {
	client *redis.Client
	log    *slog.Logger
}

func NewRedis(client *redis.Client) *Redis {
	return &Redis{client: client, log: logging.New("cache")}
}

func (r *Redis) GetBytes(ctx context.Context, key string) ([]byte, bool, error) {
	raw, err := r.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("redis get: %w", err)
	}
	return raw, true, nil
}

func (r *Redis) SetBytes(ctx context.Context, key string, data []byte, ttl time.Duration) error {
	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}
	return nil
}

func (r *Redis) GetManyBytes(ctx context.Context, keys ...string) ([][]byte, error) {
	vals, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("redis mget: %w", err)
	}
	out := make([][]byte, len(vals))
	for i, v := range vals {
		if v == nil {
			continue
		}
		if s, ok := v.(string); ok {
			out[i] = []byte(s)
		}
	}
	return out, nil
}

func (r *Redis) Delete(ctx context.Context, keys ...string) error {
	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("redis del: %w", err)
	}
	return nil
}

func (r *Redis) TTL(ctx context.Context, key string) (time.Duration, error) {
	d, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("redis ttl: %w", err)
	}
	return d, nil
}

func (r *Redis) TTLMany(ctx context.Context, keys ...string) ([]time.Duration, error) {
	pipe := r.client.Pipeline()
	cmds := make([]*redis.DurationCmd, len(keys))
	for i, k := range keys {
		cmds[i] = pipe.TTL(ctx, k)
	}
	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("redis ttl pipeline: %w", err)
	}
	out := make([]time.Duration, len(keys))
	for i, cmd := range cmds {
		out[i] = cmd.Val()
	}
	return out, nil
}

func (r *Redis) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if err := r.client.Expire(ctx, key, ttl).Err(); err != nil {
		return fmt.Errorf("redis expire: %w", err)
	}
	return nil
}

var incrExpireScript = redis.NewScript(`
local n = redis.call('INCR', KEYS[1])
redis.call('EXPIRE', KEYS[1], tonumber(ARGV[1]))
return n
`)

func (r *Redis) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	if ttl == 0 {
		n, err := r.client.Incr(ctx, key).Result()
		if err != nil {
			return 0, fmt.Errorf("redis incr: %w", err)
		}
		return n, nil
	}
	n, err := incrExpireScript.Run(ctx, r.client, []string{key}, int64(ttl.Seconds())).Int64()
	if err != nil {
		return 0, fmt.Errorf("redis incr: %w", err)
	}
	return n, nil
}
