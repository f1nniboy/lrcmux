package providers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/f1nniboy/lrcmux/internal/cache"
	"github.com/f1nniboy/lrcmux/internal/config"
	"github.com/f1nniboy/lrcmux/internal/logging"
)

type provider struct {
	Impl
	id string
}

func (p *provider) ID() string { return p.id }

type ClientResolver func(proxy string, timeout time.Duration) *http.Client

func BuildAll(cfg *config.Root, c cache.Cache, resolve ClientResolver) ([]Provider, error) {
	timeout := cfg.Provider.Timeout.Duration
	out := make([]Provider, 0, len(registry))
	for name, factory := range registry {
		common, enabled, err := loadCommon(cfg, name)
		if err != nil {
			return nil, err
		}
		if !enabled {
			continue
		}
		prim := cfg.Providers[name]
		decode := func(into any) error {
			if err := cfg.Meta.PrimitiveDecode(prim, into); err != nil {
				return fmt.Errorf("provider %q: %w", name, err)
			}
			if err := config.Validate(into); err != nil {
				return fmt.Errorf("provider %q: %w", name, err)
			}
			return nil
		}
		client := withUserAgent(resolve(common.Proxy, timeout))
		impl, err := factory(FactoryArgs{
			Decode: decode,
			Client: client,
			Cache:  c,
			Log:    logging.New("providers." + name),
		})
		if err != nil {
			return nil, err
		}
		out = append(out, &provider{Impl: impl, id: name})
	}
	return out, nil
}
