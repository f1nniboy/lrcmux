package orchestrator

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/f1nniboy/lrcmux/internal/cache"
	"github.com/f1nniboy/lrcmux/internal/isrc"
	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

const testISRC = "XX0000000000"

// fails every request, so the resolver can only use the cache
type offline struct{}

func (offline) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("offline")
}

// answers every deezer request with the test track, counting them
type fakeDeezer struct{ calls atomic.Int32 }

func (f *fakeDeezer) RoundTrip(*http.Request) (*http.Response, error) {
	f.calls.Add(1)
	body := `{"id":1,"isrc":"` + testISRC + `","title":"Song","artist":{"name":"Artist"},"duration":200}`
	return &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
}

func buildOrch(t *testing.T, deezer http.RoundTripper, provs ...*stubProvider) (*Orchestrator, cache.Cache) {
	t.Helper()
	c := cache.NewMemory()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	ps := make([]providers.Provider, len(provs))
	for i, p := range provs {
		ps[i] = p
	}
	o := New(ps, c, NewBreaker(c, log), isrc.New(&http.Client{Transport: deezer}, c, time.Hour, log), nil, Options{Timeout: 2 * time.Second})
	o.log = log
	return o, c
}

// the track is cached up front, so Get never reaches deezer
func newOrch(t *testing.T, provs ...*stubProvider) (*Orchestrator, cache.Cache) {
	t.Helper()
	o, c := buildOrch(t, offline{}, provs...)
	track := lyrics.Track{ISRC: testISRC, Title: "Song", Artist: "Artist", Duration: 200}
	if err := cache.Set(t.Context(), c, "isrc2track:"+testISRC, track, 0); err != nil {
		t.Fatal(err)
	}
	return o, c
}

func lineResult(context.Context) (*lyrics.Result, error) {
	return &lyrics.Result{SyncLevel: lyrics.SyncLine, Lines: []lyrics.Line{{Text: "x", StartMs: 1}}}, nil
}

func TestGetEscalatesPastCleanPlain(t *testing.T) {
	word := &stubProvider{id: "word", maxLevel: lyrics.SyncWord}
	line := &stubProvider{id: "line", maxLevel: lyrics.SyncLine, search: lineResult}
	none := &stubProvider{id: "none", maxLevel: lyrics.SyncNone}
	o, c := newOrch(t, word, line, none)
	plain := lyrics.Result{Source: lyrics.Source{ID: "none"}, SyncLevel: lyrics.SyncNone, Lines: []lyrics.Line{{Text: "x"}}}
	if err := cache.Set(t.Context(), c, cacheKey(testISRC, "none"), plain, 0); err != nil {
		t.Fatal(err)
	}

	resp, err := o.Get(t.Context(), Request{ISRC: testISRC, Level: lyrics.SyncWord})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Result.Source.ID != "line" || line.calls.Load() != 1 {
		t.Errorf("got %s after %d line calls, want line after 1", resp.Result.Source.ID, line.calls.Load())
	}
}

func TestGetCancelledLeaderDoesNotFailWaiters(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	slow := &stubProvider{id: "line", maxLevel: lyrics.SyncLine, search: func(ctx context.Context) (*lyrics.Result, error) {
		close(started)
		select {
		case <-release:
			return lineResult(ctx)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}}
	o, _ := newOrch(t, slow)
	req := Request{ISRC: testISRC, Level: lyrics.SyncLine}

	leaderCtx, cancel := context.WithCancel(t.Context())
	go func() { _, _ = o.Get(leaderCtx, req) }()
	<-started

	type result struct {
		resp *Response
		err  error
	}
	waiter := make(chan result, 1)
	go func() {
		resp, err := o.Get(t.Context(), req)
		waiter <- result{resp, err}
	}()
	// give the waiter time to join the in-flight fanout
	time.Sleep(50 * time.Millisecond)
	cancel()
	close(release)

	got := <-waiter
	if got.err != nil || got.resp.Result.Source.ID != "line" {
		t.Fatalf("waiter got %v, %v, want the line result", got.resp, got.err)
	}
	if n := slow.calls.Load(); n != 1 {
		t.Errorf("provider calls = %d, want 1", n)
	}
}

func TestGetCharging(t *testing.T) {
	errOver := errors.New("over the limit")
	req := func(charges *atomic.Int32, err error) Request {
		return Request{ISRC: testISRC, Level: lyrics.SyncLine, Charge: func(context.Context) error {
			charges.Add(1)
			return err
		}}
	}

	t.Run("a cold resolve is refused before deezer is asked", func(t *testing.T) {
		deezer := &fakeDeezer{}
		line := &stubProvider{id: "line", maxLevel: lyrics.SyncLine, search: lineResult}
		o, _ := buildOrch(t, deezer, line)
		var charges atomic.Int32
		if _, err := o.Get(t.Context(), req(&charges, errOver)); !errors.Is(err, errOver) {
			t.Fatalf("err = %v, want the charge's error", err)
		}
		if n := deezer.calls.Load(); n != 0 {
			t.Errorf("deezer calls = %d, want 0", n)
		}
		if n := line.calls.Load(); n != 0 {
			t.Errorf("provider calls = %d, want 0", n)
		}
	})

	t.Run("a cold resolve and its fanout pay once", func(t *testing.T) {
		line := &stubProvider{id: "line", maxLevel: lyrics.SyncLine, search: lineResult}
		o, _ := buildOrch(t, &fakeDeezer{}, line)
		var charges atomic.Int32
		if _, err := o.Get(t.Context(), req(&charges, nil)); err != nil {
			t.Fatal(err)
		}
		if n := charges.Load(); n != 1 {
			t.Errorf("charges = %d, want 1", n)
		}
	})

	t.Run("a request the cache answers alone is free", func(t *testing.T) {
		line := &stubProvider{id: "line", maxLevel: lyrics.SyncLine}
		o, c := newOrch(t, line)
		cached := lyrics.Result{Source: lyrics.Source{ID: "line"}, SyncLevel: lyrics.SyncLine, Lines: []lyrics.Line{{Text: "x", StartMs: 1}}}
		if err := cache.Set(t.Context(), c, cacheKey(testISRC, "line"), cached, 0); err != nil {
			t.Fatal(err)
		}
		var charges atomic.Int32
		if _, err := o.Get(t.Context(), req(&charges, nil)); err != nil {
			t.Fatal(err)
		}
		if n := charges.Load(); n != 0 {
			t.Errorf("charges = %d, want 0", n)
		}
	})

	t.Run("a known miss and a malformed isrc are free", func(t *testing.T) {
		deezer := &fakeDeezer{}
		o, c := buildOrch(t, deezer, &stubProvider{id: "line", maxLevel: lyrics.SyncLine})
		if err := cache.SetMiss(t.Context(), c, "isrc2track:"+testISRC, time.Hour); err != nil {
			t.Fatal(err)
		}
		var charges atomic.Int32
		for _, id := range []string{testISRC, "XX/0000"} {
			r := req(&charges, nil)
			r.ISRC = id
			if _, err := o.Get(t.Context(), r); !errors.Is(err, ErrNotFound) {
				t.Fatalf("%s: err = %v, want ErrNotFound", id, err)
			}
		}
		if n := deezer.calls.Load(); n != 0 || charges.Load() != 0 {
			t.Errorf("deezer calls = %d, charges = %d, want 0 and 0", n, charges.Load())
		}
	})

	t.Run("a cancelled request does no upstream work", func(t *testing.T) {
		deezer := &fakeDeezer{}
		line := &stubProvider{id: "line", maxLevel: lyrics.SyncLine, search: lineResult}
		o, _ := buildOrch(t, deezer, line)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		var charges atomic.Int32
		if _, err := o.Get(ctx, req(&charges, nil)); !errors.Is(err, context.Canceled) {
			t.Fatalf("err = %v, want context.Canceled", err)
		}
		if n := deezer.calls.Load(); n != 0 || line.calls.Load() != 0 {
			t.Errorf("deezer calls = %d, provider calls = %d, want 0 and 0", n, line.calls.Load())
		}
	})

	t.Run("no provider left by strict is free", func(t *testing.T) {
		deezer := &fakeDeezer{}
		o, _ := buildOrch(t, deezer, &stubProvider{id: "plain", maxLevel: lyrics.SyncNone})
		var charges atomic.Int32
		r := req(&charges, nil)
		r.Strict = true
		if _, err := o.Get(t.Context(), r); !errors.Is(err, ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
		if n := deezer.calls.Load(); n != 0 || charges.Load() != 0 {
			t.Errorf("deezer calls = %d, charges = %d, want 0 and 0", n, charges.Load())
		}
	})

	t.Run("cache mode never asks deezer", func(t *testing.T) {
		deezer := &fakeDeezer{}
		o, _ := buildOrch(t, deezer, &stubProvider{id: "line", maxLevel: lyrics.SyncLine})
		var charges atomic.Int32
		r := req(&charges, nil)
		r.FetchMode = FetchCache
		if _, err := o.Get(t.Context(), r); !errors.Is(err, ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
		if n := deezer.calls.Load(); n != 0 || charges.Load() != 0 {
			t.Errorf("deezer calls = %d, charges = %d, want 0 and 0", n, charges.Load())
		}
	})
}

func TestGetInstrumentalFlag(t *testing.T) {
	tests := []struct {
		name   string
		want   string
		lyrics bool
	}{
		{name: "lyrics from another provider win", lyrics: true, want: "lyrics"},
		{name: "kept when nobody has lyrics", want: "flag"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flagged := make(chan struct{})
			flag := &stubProvider{id: "flag", maxLevel: lyrics.SyncLine, search: func(context.Context) (*lyrics.Result, error) {
				defer close(flagged)
				return &lyrics.Result{Instrumental: true}, nil
			}}
			// answers after the flag, unless the fanout was cancelled by then
			other := &stubProvider{id: "lyrics", maxLevel: lyrics.SyncLine, search: func(ctx context.Context) (*lyrics.Result, error) {
				<-flagged
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(50 * time.Millisecond):
				}
				if !tc.lyrics {
					return nil, lyrics.ErrNotFound
				}
				return lineResult(ctx)
			}}
			o, _ := newOrch(t, flag, other)

			resp, err := o.Get(t.Context(), Request{ISRC: testISRC, Level: lyrics.SyncLine})
			if err != nil {
				t.Fatal(err)
			}
			if resp.Result.Source.ID != tc.want {
				t.Errorf("got %s, want %s", resp.Result.Source.ID, tc.want)
			}
		})
	}
}
