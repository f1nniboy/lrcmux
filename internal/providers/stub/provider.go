package stub

import (
	"context"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

type Provider struct {
	providers.Common
}

func (p *Provider) ID() string                 { return "stub" }
func (p *Provider) Name() string               { return "Stub" }
func (p *Provider) Desc() string               { return "For testing only" }
func (p *Provider) MaxLevel() lyrics.SyncLevel { return lyrics.SyncWord }

func (p *Provider) Search(_ context.Context, _ lyrics.Query) (*lyrics.Result, error) {
	return &lyrics.Result{
		Lines: []lyrics.Line{
			{Text: "First line", StartMs: 0, EndMs: 1000, Words: []lyrics.Word{
				{Text: "First ", StartMs: 0, EndMs: 500},
				{Text: "line", StartMs: 500, EndMs: 1000},
			}},
			{Text: "Second line", StartMs: 2000, EndMs: 3000, Words: []lyrics.Word{
				{Text: "Second ", StartMs: 2000, EndMs: 2500},
				{Text: "line", StartMs: 2500, EndMs: 3000},
			}},
		},
		SyncLevel: lyrics.SyncWord,
	}, nil
}
