package cache

import (
	"context"
	"strconv"
	"sync"
	"time"
)

type memEntry struct {
	expires time.Time
	data    []byte
}

type Memory struct {
	entries map[string]memEntry
	mu      sync.RWMutex
}

func NewMemory() *Memory {
	m := &Memory{entries: make(map[string]memEntry)}
	go m.cleanup()
	return m
}

func (m *Memory) alive(e memEntry) bool {
	return e.expires.IsZero() || time.Now().Before(e.expires)
}

func (m *Memory) cleanup() {
	t := time.NewTicker(time.Minute)
	for range t.C {
		m.mu.Lock()
		for k, e := range m.entries {
			if !m.alive(e) {
				delete(m.entries, k)
			}
		}
		m.mu.Unlock()
	}
}

func (m *Memory) GetBytes(_ context.Context, key string) ([]byte, bool, error) {
	m.mu.RLock()
	e, ok := m.entries[key]
	m.mu.RUnlock()
	if !ok || !m.alive(e) {
		return nil, false, nil
	}
	return e.data, true, nil
}

func (m *Memory) GetManyBytes(_ context.Context, keys ...string) ([][]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([][]byte, len(keys))
	for i, k := range keys {
		if e, ok := m.entries[k]; ok && m.alive(e) {
			out[i] = e.data
		}
	}
	return out, nil
}

func (m *Memory) SetBytes(_ context.Context, key string, data []byte, ttl time.Duration) error {
	var expires time.Time
	if ttl > 0 {
		expires = time.Now().Add(ttl)
	}
	m.mu.Lock()
	m.entries[key] = memEntry{data: data, expires: expires}
	m.mu.Unlock()
	return nil
}

func (m *Memory) Delete(_ context.Context, keys ...string) error {
	m.mu.Lock()
	for _, k := range keys {
		delete(m.entries, k)
	}
	m.mu.Unlock()
	return nil
}

func (m *Memory) TTL(_ context.Context, key string) (time.Duration, error) {
	m.mu.RLock()
	e, ok := m.entries[key]
	m.mu.RUnlock()
	if !ok || e.expires.IsZero() || !m.alive(e) {
		return -1, nil
	}
	return time.Until(e.expires), nil
}

func (m *Memory) TTLMany(_ context.Context, keys ...string) ([]time.Duration, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]time.Duration, len(keys))
	for i, k := range keys {
		e, ok := m.entries[k]
		if !ok || e.expires.IsZero() || !m.alive(e) {
			out[i] = -1
			continue
		}
		out[i] = time.Until(e.expires)
	}
	return out, nil
}

func (m *Memory) Expire(_ context.Context, key string, ttl time.Duration) error {
	m.mu.Lock()
	if e, ok := m.entries[key]; ok {
		e.expires = time.Now().Add(ttl)
		m.entries[key] = e
	}
	m.mu.Unlock()
	return nil
}

func (m *Memory) Incr(_ context.Context, key string, ttl time.Duration) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var n int64 = 1
	if e, ok := m.entries[key]; ok && m.alive(e) {
		if parsed, err := strconv.ParseInt(string(e.data), 10, 64); err == nil {
			n = parsed + 1
		}
	}
	var expires time.Time
	if ttl > 0 {
		expires = time.Now().Add(ttl)
	}
	m.entries[key] = memEntry{data: []byte(strconv.FormatInt(n, 10)), expires: expires}
	return n, nil
}
