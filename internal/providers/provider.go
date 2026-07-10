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
	Desc() string
	MaxLevel() lyrics.SyncLevel
	Search(ctx context.Context, q lyrics.Query) (*lyrics.Result, error)
	SetDeps(*http.Client, cache.Cache, *slog.Logger)
	Init()
}

var ErrRateLimited = errors.New("provider rate limited")

type Common struct {
	Enable bool   `toml:"enable"`
	Proxy  string `toml:"proxy,omitempty,commented"`

	HTTP  *http.Client `toml:"-"`
	Cache cache.Cache  `toml:"-"`
	Log   *slog.Logger `toml:"-"`
}

func (c *Common) SetDeps(http *http.Client, cache cache.Cache, log *slog.Logger) {
	c.HTTP = http
	c.Cache = cache
	c.Log = log
}

func (c *Common) Init() {}
