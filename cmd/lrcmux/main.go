package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/redis/go-redis/v9"

	"github.com/f1nniboy/lrcmux/internal/api"
	"github.com/f1nniboy/lrcmux/internal/cache"
	"github.com/f1nniboy/lrcmux/internal/config"
	"github.com/f1nniboy/lrcmux/internal/isrc"
	"github.com/f1nniboy/lrcmux/internal/logging"
	"github.com/f1nniboy/lrcmux/internal/meta"
	"github.com/f1nniboy/lrcmux/internal/metrics"
	"github.com/f1nniboy/lrcmux/internal/orchestrator"
	"github.com/f1nniboy/lrcmux/internal/providers"
	"github.com/f1nniboy/lrcmux/internal/proxy"
	"github.com/f1nniboy/lrcmux/internal/ratelimit"

	// don't forget to add new providers here
	_ "github.com/f1nniboy/lrcmux/internal/providers/genius"
	_ "github.com/f1nniboy/lrcmux/internal/providers/kugou"
	_ "github.com/f1nniboy/lrcmux/internal/providers/lrclib"
	_ "github.com/f1nniboy/lrcmux/internal/providers/musixmatch"
	_ "github.com/f1nniboy/lrcmux/internal/providers/ytmusic"
)

func main() {
	cfgPath := flag.String("config", "config.toml", "path to config file")
	flag.Parse()

	if dsn := os.Getenv("SENTRY_DSN"); dsn != "" {
		_ = sentry.Init(sentry.ClientOptions{Dsn: dsn, Release: meta.Version, TracesSampleRate: 0})
		defer sentry.Flush(2 * time.Second)
	}

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		logging.Init(logging.Config{Level: "info", Format: "text"})
		slog.Error("config load failed", "err", err)
		os.Exit(1)
	}
	logging.Init(cfg.Log)
	log := logging.New("main")

	var rdb *redis.Client
	var cacheLayer cache.Cache
	if cfg.Cache.RedisURL == "" {
		log.Warn("no redis url set, using memory cache")
		cacheLayer = cache.NewMemory()
	} else {
		rdbOpts, err := redis.ParseURL(cfg.Cache.RedisURL)
		if err != nil {
			log.Error("invalid redis url", "err", err)
			os.Exit(1)
		}
		rdb = redis.NewClient(rdbOpts)
		defer rdb.Close()

		pingCtx, cancelPing := context.WithTimeout(context.Background(), 3*time.Second)
		if err := rdb.Ping(pingCtx).Err(); err != nil {
			cancelPing()
			log.Error("redis ping failed", "err", err, "url", cfg.Cache.RedisURL)
			os.Exit(1)
		}
		cancelPing()
		cacheLayer = cache.NewRedis(rdb)
	}

	pools, err := proxy.LoadAll(cfg.Proxies, logging.New("proxy"))
	if err != nil {
		log.Error("proxy load failed", "err", err)
		os.Exit(1)
	}
	resolve := func(name string, timeout time.Duration) *http.Client {
		if name != "" {
			if _, ok := pools.Pool(name); !ok {
				log.Warn("provider references unknown proxy pool, using default client", "pool", name)
			}
		}
		return pools.ClientFor(name, timeout)
	}

	provs, err := providers.BuildAll(cfg, cacheLayer, resolve)
	if err != nil {
		log.Error("provider build failed", "err", err)
		os.Exit(1)
	}
	if len(provs) == 0 {
		log.Warn("no providers enabled")
	}
	for _, p := range provs {
		log.Info("provider enabled", "id", p.ID(), "name", p.Name())
	}

	isrcClient := &http.Client{Timeout: 3 * time.Second}
	isrcResolver := isrc.New(isrcClient, cacheLayer, cfg.Cache.MissTTL.Duration, logging.New("isrc"))

	var coll *metrics.Collector
	if cfg.Metrics.Listen != "" {
		coll = metrics.New(cfg.Metrics.Listen)
		log.Info("metrics enabled", "addr", cfg.Metrics.Listen)
	}

	breaker := orchestrator.NewBreaker(cacheLayer, logging.New("breaker"))
	orch := orchestrator.New(provs, cacheLayer, breaker, isrcResolver, coll, orchestrator.Options{
		Timeout:      cfg.Provider.Timeout.Duration,
		CacheTTL:     cfg.Cache.TTL.Duration,
		CacheMissTTL: cfg.Cache.MissTTL.Duration,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var rl *ratelimit.Limiter
	if cfg.RateLimit.Limit > 0 && rdb != nil {
		rl = ratelimit.New(rdb, cfg.RateLimit.Limit, cfg.RateLimit.Window.Duration, logging.New("ratelimit"))
	}

	srv := api.NewServer(orch, rl, cfg.Provider.Hide, cfg.Server.RequireCloudflare, coll, logging.New("api"))
	runErr := srv.Run(ctx, cfg.Server.Listen)
	if runErr != nil {
		log.Error("server error", "err", runErr)
		os.Exit(1)
	}
	log.Info("bye")
}
