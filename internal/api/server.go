package api

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"

	"github.com/f1nniboy/lrcmux/internal/config"
	"github.com/f1nniboy/lrcmux/internal/meta"
	"github.com/f1nniboy/lrcmux/internal/metrics"
	"github.com/f1nniboy/lrcmux/internal/orchestrator"
	"github.com/f1nniboy/lrcmux/internal/ratelimit"
)

//go:embed docs.md
var docsMD string

type Server struct {
	orch    *orchestrator.Orchestrator
	rate    *ratelimit.Limiter
	log     *slog.Logger
	srv     *http.Server
	api     huma.API
	cfg     *config.Root
	metrics *metrics.Collector
}

func NewServer(orch *orchestrator.Orchestrator, rate *ratelimit.Limiter, cfg *config.Root, coll *metrics.Collector, log *slog.Logger) *Server {
	if coll != nil {
		coll.Register(newBreakerCollector(orch))
	}
	return &Server{
		orch:    orch,
		rate:    rate,
		log:     log,
		cfg:     cfg,
		metrics: coll,
	}
}

func (s *Server) Run(ctx context.Context, listen string) error {
	r := chi.NewRouter()
	r.Use(cors, recoverer(s.log), withIP, accessLog(s.log))

	// why is there no good way to get the requesting client's IP in CURRENT_YEAR
	if s.cfg.Server.RequireCloudflare {
		if err := refreshCloudflareIPs(ctx); err != nil {
			return fmt.Errorf("initial cloudflare ip fetch: %w", err)
		}
		s.log.Info("cloudflare ip ranges loaded", "count", len(*cfPrefixes.Load()))
		go runCloudflareRefresh(ctx, s.log)
		r.Use(requireCloudflare)
	}

	docs, err := renderDocs(docsMD, s.orch, s.rate, s.cfg.Provider.Hide)
	if err != nil {
		s.log.Warn("docs render failed", "err", err)
		docs = docsMD
	}

	humaCfg := huma.DefaultConfig(meta.AppName, meta.Version)
	humaCfg.OpenAPI.Info.Description = docs
	humaCfg.DocsPath = ""
	humaCfg.OpenAPIPath = "/openapi"
	humaCfg.CreateHooks = nil // disable $schema injection in response bodies
	s.api = humachi.New(r, humaCfg)

	huma.Register(s.api, s.getOp(), s.handleGet)
	huma.Register(s.api, s.statsOp(), s.handleStats)
	huma.Register(s.api, s.kpoeOp(), s.handleKpoe)
	huma.Register(s.api, s.lrclibOp(), s.handleLrclib)
	huma.Register(s.api, s.lrclibSearchOp(), s.handleLrclibSearch)

	root := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("what are you doing here?"))
	}
	r.Get("/", root)
	r.Head("/", root)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	if s.metrics != nil && s.metrics.Listen != "" {
		metricsMux := http.NewServeMux()
		metricsMux.Handle("/metrics", s.metrics.Handler())
		metricsSrv := &http.Server{
			Addr:    s.metrics.Listen,
			Handler: metricsMux,
		}
		go func() {
			s.log.Info("metrics listening", "addr", s.metrics.Listen)
			if err := metricsSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				s.log.Warn("metrics server error", "err", err)
			}
		}()
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			metricsSrv.Shutdown(shutdownCtx)
		}()
	}

	s.srv = &http.Server{
		Addr:    listen,
		Handler: r,
	}

	errCh := make(chan error, 1)
	go func() {
		s.log.Info("listening", "addr", listen)
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
