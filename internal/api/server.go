package api

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"

	"github.com/f1nniboy/lrcmux/internal/meta"
	"github.com/f1nniboy/lrcmux/internal/orchestrator"
	"github.com/f1nniboy/lrcmux/internal/ratelimit"
)

//go:embed docs.md
var docsMD string

//go:embed docs.html
var docsHTML []byte

type Server struct {
	orch              *orchestrator.Orchestrator
	rl                *ratelimit.Limiter
	log               *slog.Logger
	srv               *http.Server
	api               huma.API
	hide              bool
	analyticsKey      string
	requireCloudflare bool
}

func NewServer(orch *orchestrator.Orchestrator, rl *ratelimit.Limiter, hide bool, analyticsKey string, requireCloudflare bool, log *slog.Logger) *Server {
	return &Server{orch: orch, rl: rl, log: log, hide: hide, analyticsKey: analyticsKey, requireCloudflare: requireCloudflare}
}

func (s *Server) Run(ctx context.Context, listen string) error {
	r := chi.NewRouter()
	r.Use(cors, recoverer(s.log), accessLog(s.log), withIP)

	// why is there no good way to get the requesting client's IP in CURRENT_YEAR
	if s.requireCloudflare {
		if err := refreshCloudflareIPs(ctx); err != nil {
			return fmt.Errorf("initial cloudflare ip fetch: %w", err)
		}
		s.log.Info("cloudflare ip ranges loaded", "count", len(*cfPrefixes.Load()))
		go runCloudflareRefresh(ctx, s.log)
		r.Use(requireCloudflare)
	}

	docs, err := renderDocs(docsMD, s.orch, s.rl, s.hide)
	if err != nil {
		s.log.Warn("docs render failed", "err", err)
		docs = docsMD
	}

	cfg := huma.DefaultConfig(meta.AppName, meta.Version)
	cfg.OpenAPI.Info.Description = docs
	cfg.DocsPath = ""
	cfg.OpenAPIPath = "/api/openapi"
	cfg.CreateHooks = nil // disable $schema injection in response bodies
	s.api = humachi.New(r, cfg)

	huma.Register(s.api, s.getOp(), s.handleGet)
	huma.Register(s.api, s.statsOp(), s.handleStats)
	huma.Register(s.api, s.kpoeOp(), s.handleKpoe)
	huma.Register(s.api, s.lrclibOp(), s.handleLrclib)

	r.Get("/api/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	r.Get("/docs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(docsHTML)
	})

	if s.analyticsKey != "" {
		proxy := &httputil.ReverseProxy{
			Rewrite: func(pr *httputil.ProxyRequest) {
				pr.Out.URL.Scheme = "https"
				pr.Out.URL.Host = "plausible.io"
				pr.Out.Host = "plausible.io"

				switch pr.In.URL.Path {
				case "/x/script.js":
					pr.Out.URL.Path = "/js/" + s.analyticsKey + ".js"
				case "/x/event":
					pr.Out.URL.Path = "/api/event"
				}
			},
		}
		r.Get("/x/script.js", proxy.ServeHTTP)
		r.Post("/x/event", proxy.ServeHTTP)
	}

	assets, err := loadFrontend()
	if err != nil {
		return err
	}
	r.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") {
			http.NotFound(w, r)
			return
		}
		assets.serveSPA(w, r)
	}))

	s.srv = &http.Server{
		Addr:              listen,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
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
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		s.log.Info("shutting down")
		return s.srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
