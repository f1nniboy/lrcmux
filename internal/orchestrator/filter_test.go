package orchestrator

import (
	"errors"
	"slices"
	"testing"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

func TestFilterBySources(t *testing.T) {
	all := []providers.Provider{prov("word1", lyrics.SyncWord), prov("word2", lyrics.SyncWord), prov("none", lyrics.SyncNone)}
	tests := []struct {
		err     error
		name    string
		sources []string
		want    []string
	}{
		{name: "include", sources: []string{"word1"}, want: []string{"word1"}},
		{name: "exclude", sources: []string{"!none"}, want: []string{"word1", "word2"}},
		{name: "an unknown include is dropped", sources: []string{"word1", "unknown"}, want: []string{"word1"}},
		{name: "only unknown", sources: []string{"unknown"}, err: ErrInvalidSource},
		{name: "include and exclude", sources: []string{"word1", "!none"}, err: ErrInvalidSource},
		{name: "empty name", sources: []string{"word1", ""}, err: ErrInvalidSource},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := filterBySources(all, tc.sources)
			if !errors.Is(err, tc.err) || !slices.Equal(providers.IDs(out), tc.want) {
				t.Errorf("got %v, %v, want %v, %v", providers.IDs(out), err, tc.want, tc.err)
			}
		})
	}
}
