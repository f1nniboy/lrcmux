package orchestrator

import (
	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

func buildTiers(provs []providers.Provider, level lyrics.SyncLevel) [][]providers.Provider {
	// group providers by their maximum sync levels
	byLevel := make(map[lyrics.SyncLevel][]providers.Provider)
	for _, p := range provs {
		byLevel[p.MaxLevel()] = append(byLevel[p.MaxLevel()], p)
	}

	// tier 0: all providers that can satisfy the requested level
	var top []providers.Provider
	for _, l := range lyrics.Levels {
		if l >= level {
			top = append(top, byLevel[l]...)
		}
	}

	// fallback tiers: providers below the requested level, highest first
	// so we get the best available result if tier 0 comes up empty
	tiers := [][]providers.Provider{top}
	for _, l := range lyrics.Levels {
		if l < level {
			if t := byLevel[l]; len(t) > 0 {
				tiers = append(tiers, t)
			}
		}
	}
	return tiers
}
