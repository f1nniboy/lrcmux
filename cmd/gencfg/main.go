package main

import (
	"fmt"
	"os"
	"time"

	"github.com/pelletier/go-toml/v2"

	"github.com/f1nniboy/lrcmux/internal/config"
	"github.com/f1nniboy/lrcmux/internal/logging"
	"github.com/f1nniboy/lrcmux/internal/providers"
	_ "github.com/f1nniboy/lrcmux/internal/providers/all"
)

func main() {
	names := providers.Names()
	provs := make(map[string]any, len(names))
	for _, name := range names {
		provs[name] = providers.Common{Enable: true}
	}

	cfg := config.Root{
		Server: config.Server{
			Listen:            ":8080",
			RequireCloudflare: false,
		},
		Cache: config.Cache{
			RedisURL: "redis://localhost:6379",
			TTL:      config.Duration{Duration: 24 * time.Hour},
			MissTTL:  config.Duration{Duration: time.Hour},
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
		Providers: provs,
	}

	b, err := toml.Marshal(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Stdout.Write(b)
}
