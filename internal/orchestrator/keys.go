package orchestrator

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

func cacheKey(isrc, source string) string {
	sum := sha256.Sum256([]byte(isrc + ":" + source))
	return "lyrics:" + hex.EncodeToString(sum[:16])
}

func queryKey(q lyrics.Query, level lyrics.SyncLevel) string {
	return q.Track.ISRC + ":" + level.String()
}

func providerIDs(provs []providers.Provider) []string {
	ids := make([]string, len(provs))
	for i, p := range provs {
		ids[i] = p.ID()
	}
	return ids
}

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
