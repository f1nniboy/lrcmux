package providers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/f1nniboy/lrcmux/internal/cache"
	"github.com/f1nniboy/lrcmux/internal/config"
	"github.com/f1nniboy/lrcmux/internal/lyrics"
)

// the full interface used throughout the application
type Provider interface {
	ID() string
	Name() string
	Desc() string
	MaxLevel() lyrics.SyncLevel
	Search(ctx context.Context, q lyrics.Query) (*lyrics.Result, error)
}

// the interface individual provider packages implement
// BuildAll wraps each Impl with its ID to create a full Provider
type Impl interface {
	Name() string
	Desc() string
	MaxLevel() lyrics.SyncLevel
	Search(ctx context.Context, q lyrics.Query) (*lyrics.Result, error)
}

var ErrRateLimited = errors.New("provider rate limited")

type Common struct {
	Enabled bool   `toml:"enabled"`
	Proxy   string `toml:"proxy"`
}

type DecodeFunc func(into any) error

type FactoryArgs struct {
	Decode DecodeFunc
	Client *http.Client
	Cache  cache.Cache
	Log    *slog.Logger
}

type Factory func(args FactoryArgs) (Impl, error)

var registry = map[string]Factory{}

func Register(name string, f Factory) {
	registry[name] = f
}

func loadCommon(cfg *config.Root, name string) (Common, bool, error) {
	var common Common
	prim, ok := cfg.Providers[name]
	if !ok {
		return common, false, nil
	}
	if err := cfg.Meta.PrimitiveDecode(prim, &common); err != nil {
		return common, false, fmt.Errorf("provider %q: %w", name, err)
	}
	return common, common.Enabled, nil
}
