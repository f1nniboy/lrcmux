package api

import (
	"context"
	"log/slog"
	"net"
	"net/http"

	"github.com/getsentry/sentry-go"
)

type ipKey struct{}

func withIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.Header.Get("CF-Connecting-IP")
		if ip == "" {
			ip = r.Header.Get("Fly-Client-IP")
		}
		if ip == "" {
			ip, _, _ = net.SplitHostPort(r.RemoteAddr)
			if ip == "" {
				ip = r.RemoteAddr
			}
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ipKey{}, ip)))
	})
}

func clientIP(ctx context.Context) string {
	ip, _ := ctx.Value(ipKey{}).(string)
	return ip
}

func recoverer(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic", "err", rec, "path", r.URL.Path)
					sentry.WithScope(func(scope *sentry.Scope) {
						scope.SetContext("request", sentry.Context{
							"method": r.Method,
							"path":   r.URL.Path,
							"query":  r.URL.RawQuery,
						})
						sentry.CurrentHub().Recover(rec)
					})
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
