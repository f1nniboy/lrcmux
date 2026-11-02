package providers

import (
	"context"
	"errors"
	"sync"

	"golang.org/x/sync/singleflight"

	"github.com/f1nniboy/lrcmux/internal/cache"
)

var ErrRejected = errors.New("secret rejected")

// a key or token every request shares, fetched again only once upstream rejects it.
// it's kept in the cache too, so a restart or another instance can reuse it
type Secret struct {
	cache cache.Cache
	fetch func(context.Context) (string, error)
	sf    singleflight.Group
	key   string

	val string
	mu  sync.Mutex
}

func NewSecret(c cache.Cache, provider string, fetch func(context.Context) (string, error)) *Secret {
	return &Secret{cache: c, key: "secret:" + provider, fetch: fetch}
}

// runs fn with the secret, and once more with a new one if fn fails with ErrRejected
func (s *Secret) Do(ctx context.Context, fn func(secret string) error) error {
	v, err := s.get(ctx, "")
	if err != nil {
		return err
	}
	if err = fn(v); !errors.Is(err, ErrRejected) {
		return err
	}
	fresh, ferr := s.get(ctx, v)
	if ferr != nil {
		return ferr
	}
	// upstream can still hand out the secret it rejects, a retry would only fail again
	if fresh == v {
		return err
	}
	return fn(fresh)
}

// stale is the rejected secret, so a caller that lost the race takes the new one
func (s *Secret) get(ctx context.Context, stale string) (string, error) {
	if v := s.load(); v != "" && v != stale {
		return v, nil
	}
	// keyed by stale, so a renewal never joins a flight that hands back the stale one,
	// and each caller still returns at its own deadline
	ch := s.sf.DoChan(stale, func() (any, error) {
		if v := s.load(); v != "" && v != stale {
			return v, nil
		}
		ctx := context.WithoutCancel(ctx)
		// another instance may have renewed it already. on a cache error we just fetch
		if b, ok, err := s.cache.GetBytes(ctx, s.key); err == nil && ok && string(b) != stale {
			s.store(string(b))
			return string(b), nil
		}
		v, err := s.fetch(ctx)
		if err != nil {
			return "", err
		}
		s.store(v)
		_ = s.cache.SetBytes(ctx, s.key, []byte(v), 0)
		return v, nil
	})
	select {
	case res := <-ch:
		if res.Err != nil {
			return "", res.Err
		}
		return res.Val.(string), nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func (s *Secret) load() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.val
}

func (s *Secret) store(v string) {
	s.mu.Lock()
	s.val = v
	s.mu.Unlock()
}
