package orchestrator

import (
	"cmp"
	"strings"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
)

func (o *Orchestrator) pick(results []*lyrics.Result, level lyrics.SyncLevel) *lyrics.Result {
	var best *lyrics.Result
	for _, r := range results {
		if r.SyncLevel >= level && (best == nil || rankResult(r, best) > 0) {
			best = r
		}
	}
	if best != nil {
		o.log.Debug("pick selected", "provider", best.Source.ID, "sync", best.SyncLevel.String())
	} else {
		o.log.Debug("pick: no result meets target level")
	}
	return best
}

// reports whether a result is good enough to stop the fanout early
func satisfies(r *lyrics.Result, level lyrics.SyncLevel) bool {
	return r.SyncLevel >= level && cleanScore(r) == 1
}

func rankResult(a, b *lyrics.Result) int {
	// ranked by weight, e.g. uncensored will beat even sync level
	return cmp.Or(
		cmp.Compare(cleanScore(a), cleanScore(b)),
		cmp.Compare(a.SyncLevel, b.SyncLevel),
		cmp.Compare(len(a.Lines), len(b.Lines)),
		cmp.Compare(b.Source.ID, a.Source.ID), // tiebreaker
	)
}

func cleanScore(r *lyrics.Result) int {
	for _, l := range r.Lines {
		if strings.Contains(l.Text, "*") {
			return 0
		}
	}
	return 1
}
