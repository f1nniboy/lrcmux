package orchestrator

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
)

func cacheKey(isrc, source string) string {
	sum := sha256.Sum256([]byte(isrc + ":" + source))
	return "lyrics:" + hex.EncodeToString(sum[:16])
}

func queryKey(q lyrics.Query, level lyrics.SyncLevel) string {
	return q.Track.ISRC + ":" + level.String()
}
