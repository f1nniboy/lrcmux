package config

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/f1nniboy/lrcmux/internal/logging"
)

type Server struct {
	Listen            string `toml:"listen" validate:"required"`
	RequireCloudflare bool   `toml:"require_cloudflare"`
}

type Cache struct {
	Redis   string   `toml:"redis" validate:"required"`
	TTL     Duration `toml:"ttl"`
	MissTTL Duration `toml:"miss_ttl"`
}

type Proxy struct {
	URLs []string `toml:"urls" validate:"required,min=1,dive,url"`
}

type ProviderOptions struct {
	Timeout Duration `toml:"timeout"`
	Hide    bool     `toml:"hide"`
}

type RateLimit struct {
	Enabled bool     `toml:"enabled"`
	Limit   int64    `toml:"limit"`
	Window  Duration `toml:"window"`
}

type Analytics struct {
	Key string `toml:"key"`
}

type Root struct {
	Server    Server                    `toml:"server"`
	Cache     Cache                     `toml:"cache"`
	Log       logging.Config            `toml:"log"`
	Provider  ProviderOptions           `toml:"provider"`
	RateLimit RateLimit                 `toml:"ratelimit"`
	Proxies   map[string]Proxy          `toml:"proxies"`
	Providers map[string]toml.Primitive `toml:"providers"`
	Analytics Analytics                 `toml:"analytics"`

	Meta toml.MetaData `toml:"-"`
}

func Load(path string) (*Root, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var r Root
	md, err := toml.Decode(string(data), &r)
	if err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	r.Meta = md

	if r.Server.Listen == "" {
		r.Server.Listen = ":8080"
	}
	if r.Provider.Timeout.Duration == 0 {
		r.Provider.Timeout.Duration = 5 * time.Second
	}
	if r.Cache.MissTTL.Duration == 0 {
		r.Cache.MissTTL.Duration = time.Hour
	}
	if r.RateLimit.Enabled {
		if r.RateLimit.Limit <= 0 {
			return nil, fmt.Errorf("ratelimit.limit must be positive when enabled")
		}
		if r.RateLimit.Window.Duration <= 0 {
			return nil, fmt.Errorf("ratelimit.window must be set when enabled")
		}
	}

	if v := os.Getenv("REDIS_URL"); v != "" {
		r.Cache.Redis = v
	}
	if v := os.Getenv("PORT"); v != "" {
		r.Server.Listen = ":" + v
	}

	if err := Validate(&r); err != nil {
		return nil, err
	}
	return &r, nil
}
