package orchestrator

import (
	"testing"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

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
			t.Errorf("tier 0 should have 2 word providers, got %v", providers.IDs(tiers[0]))
		}
		if len(tiers[1]) != 1 || tiers[1][0].ID() != "line1" {
			t.Errorf("tier 1 should be [line1], got %v", providers.IDs(tiers[1]))
		}
		if len(tiers[2]) != 1 || tiers[2][0].ID() != "none1" {
			t.Errorf("tier 2 should be [none1], got %v", providers.IDs(tiers[2]))
		}
	})

	t.Run("line request: two tiers, word providers in tier 0", func(t *testing.T) {
		tiers := buildTiers(all, lyrics.SyncLine)
		if len(tiers) != 2 {
			t.Fatalf("expected 2 tiers, got %d", len(tiers))
		}
		if len(tiers[0]) != 3 {
			t.Errorf("tier 0 should have word+line providers (3), got %v", providers.IDs(tiers[0]))
		}
		if len(tiers[1]) != 1 || tiers[1][0].ID() != "none1" {
			t.Errorf("tier 1 should be [none1], got %v", providers.IDs(tiers[1]))
		}
	})

	t.Run("none request: single tier with all providers", func(t *testing.T) {
		tiers := buildTiers(all, lyrics.SyncNone)
		if len(tiers) != 1 {
			t.Fatalf("expected 1 tier, got %d", len(tiers))
		}
		if len(tiers[0]) != 4 {
			t.Errorf("tier 0 should contain all 4 providers, got %v", providers.IDs(tiers[0]))
		}
	})

	t.Run("no providers at requested level: only fallback tiers returned", func(t *testing.T) {
		tiers := buildTiers([]providers.Provider{none1}, lyrics.SyncWord)
		if len(tiers) != 1 || len(tiers[0]) != 1 || tiers[0][0].ID() != "none1" {
			t.Errorf("expected one fallback tier with [none1], got %v", tiers)
		}
	})
}
