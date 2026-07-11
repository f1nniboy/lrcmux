package providers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/f1nniboy/lrcmux/internal/cache"
	"github.com/f1nniboy/lrcmux/internal/lyrics"
)

type Provider interface {
	ID() string
	Name() string
	URL() string
	Desc() string
	MaxLevel() lyrics.SyncLevel
	Search(ctx context.Context, q lyrics.Query) (*lyrics.Result, error)
	SetDeps(*http.Client, cache.Cache, *slog.Logger)
	Init()
}

var ErrRateLimited = errors.New("provider rate limited")

func Source(p Provider) lyrics.Source {
	return lyrics.Source{ID: p.ID(), Name: p.Name(), URL: p.URL()}
}

func IDs(provs []Provider) []string {
	ids := make([]string, len(provs))
	for i, p := range provs {
		ids[i] = p.ID()
	}
	return ids
}

type Common struct {
	Cache  cache.Cache  `toml:"-"`
	HTTP   *http.Client `toml:"-"`
	Log    *slog.Logger `toml:"-"`
	Proxy  string       `toml:"proxy,omitempty,commented"`
	Enable bool         `toml:"enable"`
}

func (c *Common) URL() string { return "" }

func (c *Common) SetDeps(client *http.Client, ca cache.Cache, log *slog.Logger) {
	c.HTTP = client
	c.Cache = ca
	c.Log = log
}

func (c *Common) Init() {}
