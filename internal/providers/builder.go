package providers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/pelletier/go-toml/v2"

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
	for _, name := range Names() {
		raw, ok := cfg.Providers[name]
		if !ok {
			continue
		}
		b, err := toml.Marshal(raw)
		if err != nil {
			return nil, fmt.Errorf("provider %q: %w", name, err)
		}
		var common Common
		if err := toml.Unmarshal(b, &common); err != nil {
			return nil, fmt.Errorf("provider %q: %w", name, err)
		}
		if !common.Enable {
			continue
		}
		decode := func(into any) error {
			if err := toml.Unmarshal(b, into); err != nil {
				return fmt.Errorf("provider %q: %w", name, err)
			}
			return config.Validate(into)
		}
		client := withUserAgent(resolve(common.Proxy, timeout))
		impl, err := registry[name](FactoryArgs{
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
