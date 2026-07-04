package providers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sort"

	"github.com/f1nniboy/lrcmux/internal/cache"
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
	Enable bool   `toml:"enable"`
	Proxy  string `toml:"proxy,omitempty"`
}

func Names() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
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
