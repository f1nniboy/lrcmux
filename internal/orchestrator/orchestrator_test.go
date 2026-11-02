package orchestrator

import (
	"context"
	"io"
	"log/slog"
	"slices"
	"sync/atomic"
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
	providers.Common
	search   func(context.Context) (*lyrics.Result, error)
	id       string
	calls    atomic.Int32
	maxLevel lyrics.SyncLevel
}

func (s *stubProvider) ID() string                 { return s.id }
func (s *stubProvider) Name() string               { return s.id }
func (s *stubProvider) URL() string                { return "" }
func (s *stubProvider) Desc() string               { return "" }
func (s *stubProvider) MaxLevel() lyrics.SyncLevel { return s.maxLevel }
func (s *stubProvider) Search(ctx context.Context, _ lyrics.Query) (*lyrics.Result, error) {
	s.calls.Add(1)
	if s.search == nil {
		return nil, lyrics.ErrNotFound
	}
	return s.search(ctx)
}

func prov(id string, level lyrics.SyncLevel) providers.Provider {
	return &stubProvider{id: id, maxLevel: level}
}

func TestWorthQuerying(t *testing.T) {
	all := []providers.Provider{prov("word", lyrics.SyncWord), prov("line", lyrics.SyncLine), prov("none", lyrics.SyncNone)}
	tests := []struct {
		name   string
		cached []*lyrics.Result
		want   []string
	}{
		{name: "nothing cached", want: []string{"word", "line", "none"}},
		{name: "line cached", cached: []*lyrics.Result{result("cached", lyrics.SyncLine, "x")}, want: []string{"word"}},
		{name: "word cached", cached: []*lyrics.Result{result("cached", lyrics.SyncWord, "x")}},
		{name: "instrumental cached, the others may have lyrics", cached: []*lyrics.Result{{Source: lyrics.Source{ID: "cached"}, Instrumental: true}}, want: []string{"word", "line", "none"}},
		{name: "censored word cached, anything clean beats it", cached: []*lyrics.Result{result("cached", lyrics.SyncWord, "c**sored")}, want: []string{"word", "line", "none"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := providers.IDs(worthQuerying(slices.Clone(all), tc.cached)); !slices.Equal(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}
