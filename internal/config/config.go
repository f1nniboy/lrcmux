package config

import (
	"fmt"
	"os"
	"time"

	"github.com/pelletier/go-toml/v2"

	"github.com/f1nniboy/lrcmux/internal/logging"
)

type Server struct {
	Listen            string `toml:"listen" validate:"required" comment:"address to listen on, overridable via LISTEN env var"`
	RequireCloudflare bool   `toml:"require_cloudflare,commented" comment:"reject requests not arriving via Cloudflare"`
}

type CacheTTL struct {
	Word Duration `toml:"word" comment:"word-level results"`
	Line Duration `toml:"line" comment:"line-level results"`
	None Duration `toml:"none" comment:"unsynced results"`
	Miss Duration `toml:"miss" comment:"provider not-found results"`
}

type Cache struct {
	RedisURL string   `toml:"redis_url,commented" comment:"Redis URL, overridable via REDIS_URL env var, omit to use in-memory cache"`
	TTL      CacheTTL `toml:"ttl" comment:"comment out keys to disable expiration"`
}

type Proxy struct {
	URLs []string `toml:"urls" validate:"required,min=1,dive,url"`
}

type ProviderOptions struct {
	Timeout Duration `toml:"timeout" comment:"timeout for provider requests"`
	Hide    bool     `toml:"hide" comment:"hide source from responses and /stats endpoint"`
}

type RateLimit struct {
	Limit  int64    `toml:"limit" comment:"max requests per window per IP"`
	Window Duration `toml:"window" comment:"rate limit time window"`
}

type Metrics struct {
	Listen string `toml:"listen" comment:"address to expose Prometheus metrics on"`
}

type Root struct {
	Server    Server           `toml:"server"`
	Cache     Cache            `toml:"cache"`
	Log       logging.Config   `toml:"log"`
	Provider  ProviderOptions  `toml:"provider"`
	RateLimit RateLimit        `toml:"ratelimit"`
	Metrics   Metrics          `toml:"metrics,commented"`
	Proxies   map[string]Proxy `toml:"proxies,commented"`
	Providers map[string]any   `toml:"providers"`
}

func Load(path string) (*Root, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var r Root
	if err := toml.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	if r.Server.Listen == "" {
		r.Server.Listen = ":8080"
	}
	if r.Provider.Timeout.Duration == 0 {
		r.Provider.Timeout.Duration = 5 * time.Second
	}
	if r.RateLimit.Limit > 0 {
		if r.RateLimit.Window.Duration <= 0 {
			return nil, fmt.Errorf("ratelimit.window must be set when enabled")
		}
	}
	if v := os.Getenv("REDIS_URL"); v != "" {
		r.Cache.RedisURL = v
	}
	if v := os.Getenv("LISTEN"); v != "" {
		r.Server.Listen = v
	}

	if err := Validate(&r); err != nil {
		return nil, err
	}
	return &r, nil
}
