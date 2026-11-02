package proxy

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/url"
	"time"

	"github.com/f1nniboy/lrcmux/internal/config"
)

type Pool struct {
	urls []*url.URL
}

type Registry struct {
	pools map[string]*Pool
}

func LoadAll(cfgs map[string]config.Proxy) (*Registry, error) {
	reg := &Registry{pools: make(map[string]*Pool, len(cfgs))}
	for name, c := range cfgs {
		urls := make([]*url.URL, 0, len(c.URLs))
		for _, raw := range c.URLs {
			u, err := url.Parse(raw)
			if err != nil {
				return nil, fmt.Errorf("proxy pool %q: bad url %q: %w", name, raw, err)
			}
			urls = append(urls, u)
		}
		reg.pools[name] = &Pool{urls: urls}
	}
	return reg, nil
}

func (r *Registry) Pool(name string) (*Pool, bool) {
	if name == "" {
		return nil, false
	}
	p, ok := r.pools[name]
	return p, ok
}

func (r *Registry) ClientFor(name string, timeout time.Duration) *http.Client {
	var rt http.RoundTripper = http.DefaultTransport.(*http.Transport).Clone()
	if pool, ok := r.Pool(name); ok {
		rt = &poolTransport{pool: pool, fresh: pool.transport(nil)}
	}
	return &http.Client{Timeout: timeout, Transport: rt}
}

type poolTransport struct {
	pool  *Pool
	fresh *http.Transport
}

func (t *poolTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if s, ok := r.Context().Value(sessionKey{}).(*session); ok {
		if st := s.transport(t.pool); st != nil {
			return st.RoundTrip(r)
		}
	}
	return t.fresh.RoundTrip(r)
}

func (p *Pool) transport(fixed *url.URL) *http.Transport {
	t := http.DefaultTransport.(*http.Transport).Clone()
	if fixed != nil {
		t.Proxy = http.ProxyURL(fixed)
		return t
	}
	t.Proxy = func(*http.Request) (*url.URL, error) { return p.pick(), nil }
	// a reused connection keeps its exit ip, so the load would never spread
	t.DisableKeepAlives = true
	return t
}

func (p *Pool) pick() *url.URL {
	return p.urls[rand.IntN(len(p.urls))]
}
