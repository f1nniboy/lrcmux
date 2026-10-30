package providers

import (
	"context"
	"errors"
	"sync"

	"golang.org/x/sync/singleflight"
)

var ErrRejected = errors.New("secret rejected")

// a key or token every request shares, fetched again only once upstream rejects it
type Secret struct {
	fetch func(context.Context) (string, error)
	sf    singleflight.Group

	val string
	mu  sync.Mutex
}

func NewSecret(fetch func(context.Context) (string, error)) *Secret {
	return &Secret{fetch: fetch}
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
	if v, err = s.get(ctx, v); err != nil {
		return err
	}
	return fn(v)
}

// stale is the rejected secret, so a caller that lost the race takes the new one
func (s *Secret) get(ctx context.Context, stale string) (string, error) {
	if v := s.load(); v != "" && v != stale {
		return v, nil
	}
	// keyed by stale, so a renewal never joins a flight that hands back the stale one,
	// and each caller still leaves at its own deadline
	ch := s.sf.DoChan(stale, func() (any, error) {
		if v := s.load(); v != "" && v != stale {
			return v, nil
		}
		v, err := s.fetch(context.WithoutCancel(ctx))
		if err != nil {
			return "", err
		}
		s.mu.Lock()
		s.val = v
		s.mu.Unlock()
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
