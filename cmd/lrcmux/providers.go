package main

import (
	"fmt"
	"log/slog"

	"github.com/pelletier/go-toml/v2"

	"github.com/f1nniboy/lrcmux/internal/cache"
	"github.com/f1nniboy/lrcmux/internal/config"
	"github.com/f1nniboy/lrcmux/internal/logging"
	"github.com/f1nniboy/lrcmux/internal/providers"
	"github.com/f1nniboy/lrcmux/internal/proxy"

	"github.com/f1nniboy/lrcmux/internal/providers/genius"
	"github.com/f1nniboy/lrcmux/internal/providers/kugou"
	"github.com/f1nniboy/lrcmux/internal/providers/lrclib"
	"github.com/f1nniboy/lrcmux/internal/providers/musixmatch"
	"github.com/f1nniboy/lrcmux/internal/providers/ytmusic"
)

func buildProviders(cfg *config.Root, c cache.Cache, pools *proxy.Registry, log *slog.Logger) ([]providers.Provider, error) {
	timeout := cfg.Provider.Timeout.Duration

	provs := []providers.Provider{
		&genius.Provider{},
		&kugou.Provider{},
		&lrclib.Provider{},
		&musixmatch.Provider{},
		//&stub.Provider{},
		&ytmusic.Provider{},
	}

	var out []providers.Provider
	for _, p := range provs {
		raw, ok := cfg.Providers[p.ID()]
		if !ok {
			continue
		}
		b, err := toml.Marshal(raw)
		if err != nil {
			return nil, fmt.Errorf("provider %q: %w", p.ID(), err)
		}
		var common providers.Common
		if err := toml.Unmarshal(b, &common); err != nil {
			return nil, fmt.Errorf("provider %q: %w", p.ID(), err)
		}
		if !common.Enable {
			continue
		}
		if err := toml.Unmarshal(b, p); err != nil {
			return nil, fmt.Errorf("provider %q: %w", p.ID(), err)
		}
		if common.Proxy != "" {
			if _, ok := pools.Pool(common.Proxy); !ok {
				log.Warn("unknown proxy pool, using default client", "provider", p.ID(), "pool", common.Proxy)
			}
		}
		client := providers.WithUserAgent(pools.ClientFor(common.Proxy, timeout))
		p.SetDeps(client, c, logging.New("providers."+p.ID()))
		p.Init()
		out = append(out, p)
	}
	return out, nil
}
