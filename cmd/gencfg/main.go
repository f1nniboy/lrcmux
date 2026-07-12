package main

import (
	"fmt"
	"os"
	"time"

	"github.com/pelletier/go-toml/v2"

	"github.com/f1nniboy/lrcmux/internal/config"
	"github.com/f1nniboy/lrcmux/internal/logging"
	"github.com/f1nniboy/lrcmux/internal/providers"
	"github.com/f1nniboy/lrcmux/internal/providers/genius"
	"github.com/f1nniboy/lrcmux/internal/providers/kugou"
	"github.com/f1nniboy/lrcmux/internal/providers/lrclib"
	"github.com/f1nniboy/lrcmux/internal/providers/musixmatch"
	"github.com/f1nniboy/lrcmux/internal/providers/netease"
	"github.com/f1nniboy/lrcmux/internal/providers/ytmusic"
)

func main() {
	enabled := providers.Common{Enable: true}

	cfg := config.Root{
		Server: config.Server{
			Listen:            ":8080",
			RequireCloudflare: false,
		},
		Cache: config.Cache{
			RedisURL: "redis://localhost:6379",
			TTL: config.CacheTTL{
				Word: config.Duration{Duration: 0},
				Line: config.Duration{Duration: 720 * time.Hour},
				None: config.Duration{Duration: 168 * time.Hour},
				Miss: config.Duration{Duration: 24 * time.Hour},
			},
		},
		Log: logging.Config{Level: "info", Format: "text"},
		Provider: config.ProviderOptions{
			Timeout: config.Duration{Duration: 5 * time.Second},
		},
		RateLimit: config.RateLimit{
			Limit:  45,
			Window: config.Duration{Duration: time.Hour},
		},
		Metrics: config.Metrics{Listen: ":9091"},
		Proxies: map[string]config.Proxy{
			"mypool": {URLs: []string{"socks5://user:pass@proxy1.example.com:1080"}},
		},
		Providers: map[string]any{
			"genius":     genius.Provider{Common: providers.Common{Enable: true, Proxy: "mypool"}},
			"kugou":      kugou.Provider{Common: enabled},
			"lrclib":     lrclib.Provider{Common: enabled, BaseURL: "https://lrclib.net"},
			"musixmatch": musixmatch.Provider{Common: enabled, PoolSize: 5},
			"netease":    netease.Provider{Common: enabled},
			"ytmusic":    ytmusic.Provider{Common: enabled},
		},
	}

	b, err := toml.Marshal(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Stdout.Write(b)
}
