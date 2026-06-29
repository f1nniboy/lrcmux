package api

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/netip"
	"strings"
	"sync/atomic"
	"time"
)

const (
	cfIPv4URL         = "https://www.cloudflare.com/ips-v4"
	cfIPv6URL         = "https://www.cloudflare.com/ips-v6"
	cfRefreshInterval = 24 * time.Hour
	cfFetchTimeout    = 10 * time.Second
)

var cfPrefixes atomic.Pointer[[]netip.Prefix]

func parseCIDRList(s string) []netip.Prefix {
	var out []netip.Prefix
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if p, err := netip.ParsePrefix(line); err == nil {
			out = append(out, p)
		}
	}
	return out
}

func fromCloudflare(ip string) bool {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return false
	}
	loaded := cfPrefixes.Load()
	if loaded == nil {
		return false
	}
	for _, p := range *loaded {
		if p.Contains(addr) {
			return true
		}
	}
	return false
}

func fetchCFList(ctx context.Context, url string) ([]netip.Prefix, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return nil, err
	}
	prefixes := parseCIDRList(string(body))
	if len(prefixes) == 0 {
		return nil, fmt.Errorf("empty list")
	}
	return prefixes, nil
}

func refreshCloudflareIPs(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, cfFetchTimeout)
	defer cancel()

	v4, err := fetchCFList(ctx, cfIPv4URL)
	if err != nil {
		return fmt.Errorf("ipv4: %w", err)
	}
	v6, err := fetchCFList(ctx, cfIPv6URL)
	if err != nil {
		return fmt.Errorf("ipv6: %w", err)
	}
	merged := append(v4, v6...)
	cfPrefixes.Store(&merged)
	return nil
}

// refreshes the IP list on a timer, on failure the previous list stays
func runCloudflareRefresh(ctx context.Context, log *slog.Logger) {
	t := time.NewTicker(cfRefreshInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := refreshCloudflareIPs(ctx); err != nil {
				log.Warn("cloudflare ip refresh failed", "err", err)
				continue
			}
			log.Info("cloudflare ip ranges refreshed", "count", len(*cfPrefixes.Load()))
		}
	}
}

// rejects any request whose Fly edge connection did not originate from a
// Cloudflare IP (only for /api paths)
func requireCloudflare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api") || r.URL.Path == "/api/health" {
			next.ServeHTTP(w, r)
			return
		}
		if !fromCloudflare(r.Header.Get("Fly-Client-IP")) {
			http.Error(w, "glory to cloudflare", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
