package providers

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

func counting() (func(context.Context) (string, error), *atomic.Int32) {
	var n atomic.Int32
	return func(context.Context) (string, error) {
		return fmt.Sprintf("secret-%d", n.Add(1)), nil
	}, &n
}

func TestSecretShared(t *testing.T) {
	fetch, n := counting()
	s := NewSecret(fetch)
	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			if err := s.Do(t.Context(), func(string) error { return nil }); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if n.Load() != 1 {
		t.Errorf("fetches = %d, want 1", n.Load())
	}
}

func TestSecretRejected(t *testing.T) {
	fetch, n := counting()
	s := NewSecret(fetch)
	var seen []string
	err := s.Do(t.Context(), func(v string) error {
		seen = append(seen, v)
		if v == "secret-1" {
			return ErrRejected
		}
		return nil
	})
	if err != nil || len(seen) != 2 || seen[1] != "secret-2" {
		t.Fatalf("err %v, saw %q, want a retry with secret-2", err, seen)
	}
	// a caller that was rejected with the same old secret takes the new one
	if v, _ := s.get(t.Context(), "secret-1"); v != "secret-2" || n.Load() != 2 {
		t.Errorf("got %q after %d fetches, want secret-2 after 2", v, n.Load())
	}
}

func TestSecretCallerLeaves(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	s := NewSecret(func(context.Context) (string, error) {
		close(started)
		<-release
		return "secret", nil
	})

	leaving, cancel := context.WithCancel(t.Context())
	left := make(chan error, 1)
	go func() {
		_, err := s.get(leaving, "")
		left <- err
	}()
	<-started
	staying := make(chan error, 1)
	go func() {
		_, err := s.get(t.Context(), "")
		staying <- err
	}()

	cancel()
	if err := <-left; !errors.Is(err, context.Canceled) {
		t.Errorf("leaving caller got %v, want context.Canceled", err)
	}
	close(release)
	if err := <-staying; err != nil {
		t.Errorf("staying caller got %v", err)
	}
}
