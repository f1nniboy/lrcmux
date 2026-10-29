package orchestrator

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
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

// the track is cached up front, so Get never reaches deezer
func newOrch(t *testing.T, provs ...*stubProvider) (*Orchestrator, cache.Cache) {
	t.Helper()
	c := cache.NewMemory()
	track := lyrics.Track{ISRC: testISRC, Title: "Song", Artist: "Artist", Duration: 200}
	if err := cache.Set(t.Context(), c, "isrc2track:"+testISRC, track, 0); err != nil {
		t.Fatal(err)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	ps := make([]providers.Provider, len(provs))
	for i, p := range provs {
		ps[i] = p
	}
	o := New(ps, c, NewBreaker(c, log), isrc.New(&http.Client{Transport: offline{}}, c, time.Hour, log), nil, Options{Timeout: 2 * time.Second})
	o.log = log
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
