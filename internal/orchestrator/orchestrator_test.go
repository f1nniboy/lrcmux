package orchestrator

import (
	"context"
	"io"
	"log/slog"
	"slices"
	"testing"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

func testOrch() *Orchestrator {
	return &Orchestrator{log: slog.New(slog.NewTextHandler(io.Discard, nil))}
}

func result(source string, level lyrics.SyncLevel, lines ...string) *lyrics.Result {
	r := &lyrics.Result{Source: lyrics.Source{ID: source}, SyncLevel: level}
	for _, t := range lines {
		r.Lines = append(r.Lines, lyrics.Line{Text: t})
	}
	return r
}

type stubProvider struct {
	id string
	providers.Common
	maxLevel lyrics.SyncLevel
}

func (s *stubProvider) ID() string                 { return s.id }
func (s *stubProvider) Name() string               { return s.id }
func (s *stubProvider) URL() string                { return "" }
func (s *stubProvider) Desc() string               { return "" }
func (s *stubProvider) MaxLevel() lyrics.SyncLevel { return s.maxLevel }
func (s *stubProvider) Search(_ context.Context, _ lyrics.Query) (*lyrics.Result, error) {
	return nil, nil
}

func prov(id string, level lyrics.SyncLevel) providers.Provider {
	return &stubProvider{id: id, maxLevel: level}
}

func TestWorthQuerying(t *testing.T) {
	word := prov("word", lyrics.SyncWord)
	line := prov("line", lyrics.SyncLine)
	none := prov("none", lyrics.SyncNone)
	all := []providers.Provider{word, line, none}

	t.Run("no cache: keep all providers (non-strict, level word)", func(t *testing.T) {
		out := worthQuerying(slices.Clone(all), nil, Request{Level: lyrics.SyncWord})
		if len(out) != 3 {
			t.Errorf("expected all 3 providers, got %v", providers.IDs(out))
		}
	})

	t.Run("cache has line: drop line-and-below, keep word", func(t *testing.T) {
		cached := []*lyrics.Result{result("cached", lyrics.SyncLine, "x")}
		out := worthQuerying(slices.Clone(all), cached, Request{Level: lyrics.SyncWord})
		if len(out) != 1 || out[0].ID() != "word" {
			t.Errorf("expected [word], got %v", providers.IDs(out))
		}
	})

	t.Run("cache has word: drop everything", func(t *testing.T) {
		cached := []*lyrics.Result{result("cached", lyrics.SyncWord, "x")}
		out := worthQuerying(slices.Clone(all), cached, Request{Level: lyrics.SyncWord})
		if len(out) != 0 {
			t.Errorf("expected empty, got %v", providers.IDs(out))
		}
	})

	t.Run("strict word: drop below-word providers", func(t *testing.T) {
		out := worthQuerying(slices.Clone(all), nil, Request{Level: lyrics.SyncWord, Strict: true})
		if len(out) != 1 || out[0].ID() != "word" {
			t.Errorf("expected [word], got %v", providers.IDs(out))
		}
	})

	t.Run("strict line: drop below-line providers", func(t *testing.T) {
		out := worthQuerying(slices.Clone(all), nil, Request{Level: lyrics.SyncLine, Strict: true})
		if len(out) != 2 {
			t.Errorf("expected [word line], got %v", providers.IDs(out))
		}
	})

	t.Run("cache has censored line: keep word and line providers", func(t *testing.T) {
		cached := []*lyrics.Result{result("cached", lyrics.SyncLine, "c**sored")}
		out := worthQuerying(slices.Clone(all), cached, Request{Level: lyrics.SyncWord})
		if len(out) != 2 {
			t.Errorf("expected [word line], got %v", providers.IDs(out))
		}
		for _, p := range out {
			if p.ID() == "none" {
				t.Error("none should be excluded (below cached level)")
			}
		}
	})

	t.Run("cache has censored word: keep word provider", func(t *testing.T) {
		cached := []*lyrics.Result{result("cached", lyrics.SyncWord, "c**sored")}
		out := worthQuerying(slices.Clone(all), cached, Request{Level: lyrics.SyncWord})
		if len(out) != 1 || out[0].ID() != "word" {
			t.Errorf("expected [word], got %v", providers.IDs(out))
		}
	})
}
