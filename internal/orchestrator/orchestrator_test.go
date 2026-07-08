package orchestrator

import (
	"context"
	"io"
	"log/slog"
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
	id       string
	maxLevel lyrics.SyncLevel
}

func (s *stubProvider) ID() string                 { return s.id }
func (s *stubProvider) Name() string               { return s.id }
func (s *stubProvider) Desc() string               { return "" }
func (s *stubProvider) MaxLevel() lyrics.SyncLevel { return s.maxLevel }
func (s *stubProvider) Search(_ context.Context, _ lyrics.Query) (*lyrics.Result, error) {
	return nil, nil
}

func prov(id string, level lyrics.SyncLevel) providers.Provider {
	return &stubProvider{id: id, maxLevel: level}
}

func TestBuildTiers(t *testing.T) {
	word1 := prov("word1", lyrics.SyncWord)
	word2 := prov("word2", lyrics.SyncWord)
	line1 := prov("line1", lyrics.SyncLine)
	none1 := prov("none1", lyrics.SyncNone)

	all := []providers.Provider{word1, word2, line1, none1}

	t.Run("word request: three tiers", func(t *testing.T) {
		tiers := buildTiers(all, lyrics.SyncWord)
		if len(tiers) != 3 {
			t.Fatalf("expected 3 tiers, got %d", len(tiers))
		}
		if len(tiers[0]) != 2 {
			t.Errorf("tier 0 should have 2 word providers, got %v", providerIDs(tiers[0]))
		}
		if len(tiers[1]) != 1 || tiers[1][0].ID() != "line1" {
			t.Errorf("tier 1 should be [line1], got %v", providerIDs(tiers[1]))
		}
		if len(tiers[2]) != 1 || tiers[2][0].ID() != "none1" {
			t.Errorf("tier 2 should be [none1], got %v", providerIDs(tiers[2]))
		}
	})

	t.Run("line request: two tiers, word providers in tier 0", func(t *testing.T) {
		tiers := buildTiers(all, lyrics.SyncLine)
		if len(tiers) != 2 {
			t.Fatalf("expected 2 tiers, got %d", len(tiers))
		}
		if len(tiers[0]) != 3 {
			t.Errorf("tier 0 should have word+line providers (3), got %v", providerIDs(tiers[0]))
		}
		if len(tiers[1]) != 1 || tiers[1][0].ID() != "none1" {
			t.Errorf("tier 1 should be [none1], got %v", providerIDs(tiers[1]))
		}
	})

	t.Run("none request: single tier with all providers", func(t *testing.T) {
		tiers := buildTiers(all, lyrics.SyncNone)
		if len(tiers) != 1 {
			t.Fatalf("expected 1 tier, got %d", len(tiers))
		}
		if len(tiers[0]) != 4 {
			t.Errorf("tier 0 should contain all 4 providers, got %v", providerIDs(tiers[0]))
		}
	})

	t.Run("no providers at requested level still returns tier 0", func(t *testing.T) {
		tiers := buildTiers([]providers.Provider{none1}, lyrics.SyncWord)
		if len(tiers[0]) != 0 {
			t.Errorf("tier 0 should be empty, got %v", providerIDs(tiers[0]))
		}
	})
}

func TestRankResult(t *testing.T) {
	cases := []struct {
		name string
		a, b *lyrics.Result
	}{
		{
			name: "word beats line",
			a:    result("a", lyrics.SyncWord, "x"),
			b:    result("b", lyrics.SyncLine, "x", "y", "z"),
		},
		{
			name: "word beats none",
			a:    result("a", lyrics.SyncWord),
			b:    result("b", lyrics.SyncNone, "x", "y"),
		},
		{
			name: "line beats none",
			a:    result("a", lyrics.SyncLine, "x"),
			b:    result("b", lyrics.SyncNone, "x", "y", "z"),
		},
		{
			name: "more lines wins at equal sync",
			a:    result("a", lyrics.SyncNone, "a", "b", "c", "d"),
			b:    result("b", lyrics.SyncNone, "x", "y"),
		},
		{
			name: "higher sync beats more lines",
			a:    result("a", lyrics.SyncWord, "x"),
			b:    result("b", lyrics.SyncNone, "x", "y", "z"),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if !rankResult(c.a, c.b) {
				t.Errorf("rankResult(%s, %s) = false, want true", c.a.Source.ID, c.b.Source.ID)
			}
		})
	}

	t.Run("equal results do not both outrank each other", func(t *testing.T) {
		a := result("a", lyrics.SyncLine, "x", "y")
		b := result("b", lyrics.SyncLine, "x", "y")
		if rankResult(a, b) && rankResult(b, a) {
			t.Error("two equal results must not both claim to outrank each other")
		}
	})
}

func TestPick(t *testing.T) {
	o := testOrch()

	t.Run("empty input", func(t *testing.T) {
		if got := o.pick(nil, lyrics.SyncWord); got != nil {
			t.Errorf("expected nil, got %+v", got)
		}
	})

	t.Run("no result meets required level", func(t *testing.T) {
		results := []*lyrics.Result{
			result("a", lyrics.SyncNone, "x"),
			result("b", lyrics.SyncLine, "x"),
		}
		if got := o.pick(results, lyrics.SyncWord); got != nil {
			t.Errorf("expected nil, got %s", got.Source.ID)
		}
	})

	t.Run("exact level match accepted", func(t *testing.T) {
		if got := o.pick([]*lyrics.Result{result("a", lyrics.SyncLine, "x")}, lyrics.SyncLine); got == nil {
			t.Error("exact level match should be accepted")
		}
	})

}
