package cache

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"time"
)

type Status uint8

const (
	Miss      Status = iota // key absent, never queried
	Hit                     // key present
	KnownMiss               // key present, but marked as miss
)

var missValue = []byte{0}

func Get[T any](ctx context.Context, c Cache, key string) (T, Status, error) {
	var zero T
	raw, ok, err := c.GetBytes(ctx, key)
	if err != nil {
		return zero, Miss, err
	}
	if !ok {
		return zero, Miss, nil
	}
	val, status := decode[T](raw)
	return val, status, nil
}

func GetMany[T any](ctx context.Context, c Cache, keys []string) ([]T, []Status, error) {
	raws, err := c.GetManyBytes(ctx, keys...)
	if err != nil {
		return nil, nil, err
	}
	vals := make([]T, len(keys))
	statuses := make([]Status, len(keys))
	for i, raw := range raws {
		if raw == nil {
			statuses[i] = Miss
			continue
		}
		vals[i], statuses[i] = decode[T](raw)
	}
	return vals, statuses, nil
}

func SetMiss(ctx context.Context, c Cache, key string, ttl time.Duration) error {
	return c.SetBytes(ctx, key, missValue, ttl)
}

// if T is string the value is stored as raw bytes, otherwise encoded
func Set[T any](ctx context.Context, c Cache, key string, val T, ttl time.Duration) error {
	var data []byte
	if s, ok := any(val).(string); ok {
		data = []byte(s)
	} else {
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(val); err != nil {
			return fmt.Errorf("gob encode: %w", err)
		}
		data = buf.Bytes()
	}
	return c.SetBytes(ctx, key, data, ttl)
}

func decode[T any](raw []byte) (T, Status) {
	var zero T
	if bytes.Equal(raw, missValue) {
		return zero, KnownMiss
	}
	var out T
	if sp, isStr := any(&out).(*string); isStr {
		*sp = string(raw)
		return out, Hit
	}
	if gob.NewDecoder(bytes.NewReader(raw)).Decode(&out) != nil {
		return zero, Miss //nolint:nilerr
	}
	return out, Hit
}
