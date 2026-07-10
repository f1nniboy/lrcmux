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
	pool, ok := r.Pool(name)
	t := http.DefaultTransport.(*http.Transport).Clone()
	if ok {
		t.Proxy = func(_ *http.Request) (*url.URL, error) {
			return pool.urls[rand.IntN(len(pool.urls))], nil
		}
		t.DisableKeepAlives = true
	}
	return &http.Client{Timeout: timeout, Transport: t}
}
