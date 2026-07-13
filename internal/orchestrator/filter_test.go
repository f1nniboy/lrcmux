package orchestrator

import (
	"errors"
	"testing"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

func TestFilterBySources(t *testing.T) {
	word1 := prov("word1", lyrics.SyncWord)
	word2 := prov("word2", lyrics.SyncWord)
	none := prov("none", lyrics.SyncNone)
	all := []providers.Provider{word1, word2, none}

	t.Run("empty sources: return all", func(t *testing.T) {
		out, err := filterBySources(all, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(out) != 3 {
			t.Errorf("expected 3 providers, got %d", len(out))
		}
	})

	t.Run("single include: keep only that", func(t *testing.T) {
		out, err := filterBySources(all, []string{"word1"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(out) != 1 || out[0].ID() != "word1" {
			t.Errorf("expected [word1], got %v", providers.IDs(out))
		}
	})

	t.Run("multiple includes: keep all listed", func(t *testing.T) {
		out, err := filterBySources(all, []string{"word1", "none"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(out) != 2 {
			t.Errorf("expected 2 providers, got %v", providers.IDs(out))
		}
	})

	t.Run("single exclude: drop that", func(t *testing.T) {
		out, err := filterBySources(all, []string{"!none"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(out) != 2 {
			t.Errorf("expected 2 providers, got %v", providers.IDs(out))
		}
		for _, p := range out {
			if p.ID() == "none" {
				t.Error("none should have been excluded")
			}
		}
	})

	t.Run("unknown provider: error", func(t *testing.T) {
		_, err := filterBySources(all, []string{"unknown"})
		if !errors.Is(err, ErrInvalidSource) {
			t.Errorf("expected ErrInvalidSource, got %v", err)
		}
	})

	t.Run("mixed include and exclude: error", func(t *testing.T) {
		_, err := filterBySources(all, []string{"word1", "!none"})
		if !errors.Is(err, ErrInvalidSource) {
			t.Errorf("expected ErrInvalidSource, got %v", err)
		}
	})

	t.Run("empty name: error", func(t *testing.T) {
		_, err := filterBySources(all, []string{"word1", ""})
		if !errors.Is(err, ErrInvalidSource) {
			t.Errorf("expected ErrInvalidSource, got %v", err)
		}
	})

	t.Run("case-insensitive: normalized to lowercase", func(t *testing.T) {
		out, err := filterBySources(all, []string{"WORD1"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(out) != 1 || out[0].ID() != "word1" {
			t.Errorf("expected [word1], got %v", providers.IDs(out))
		}
	})

	t.Run("whitespace trimmed", func(t *testing.T) {
		out, err := filterBySources(all, []string{" word1 "})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(out) != 1 || out[0].ID() != "word1" {
			t.Errorf("expected [word1], got %v", providers.IDs(out))
		}
	})
}
