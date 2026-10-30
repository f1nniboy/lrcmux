package orchestrator

import (
	"slices"
	"testing"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

func TestBuildTiers(t *testing.T) {
	word1, word2 := prov("word1", lyrics.SyncWord), prov("word2", lyrics.SyncWord)
	line, none := prov("line", lyrics.SyncLine), prov("none", lyrics.SyncNone)
	all := []providers.Provider{word1, word2, line, none}
	tests := []struct {
		name  string
		provs []providers.Provider
		want  [][]string
		level lyrics.SyncLevel
	}{
		{name: "word", provs: all, level: lyrics.SyncWord, want: [][]string{{"word1", "word2"}, {"line"}, {"none"}}},
		{name: "line takes word providers first", provs: all, level: lyrics.SyncLine, want: [][]string{{"word1", "word2", "line"}, {"none"}}},
		{name: "nothing at the level", provs: []providers.Provider{none}, level: lyrics.SyncWord, want: [][]string{{"none"}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got [][]string
			for _, tier := range buildTiers(tc.provs, tc.level) {
				got = append(got, providers.IDs(tier))
			}
			if !slices.EqualFunc(got, tc.want, slices.Equal) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}
