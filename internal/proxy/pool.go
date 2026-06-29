package proxy

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/url"
	"time"

	"github.com/f1nniboy/lrcmux/internal/config"
)

type Pool struct {
	name string
	urls []*url.URL
	log  *slog.Logger
}

func (p *Pool) Name() string { return p.name }

func (p *Pool) next() *url.URL {
	u := p.urls[rand.IntN(len(p.urls))]
	p.log.Debug("proxy chosen", "pool", p.name, "url", u.Redacted())
	return u
}

// holds named pools and produces http.Client per pool
type Registry struct {
	pools map[string]*Pool
}

func LoadAll(cfgs map[string]config.Proxy, log *slog.Logger) (*Registry, error) {
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
		reg.pools[name] = &Pool{name: name, urls: urls, log: log}
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

// ClientFor returns an *http.Client routed through the named pool, or a
// default client if the name is empty.
func (r *Registry) ClientFor(name string, timeout time.Duration) *http.Client {
	pool, ok := r.Pool(name)
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConnsPerHost = 10
	if ok {
		t.Proxy = func(*http.Request) (*url.URL, error) {
			return pool.next(), nil
		}
	}
	return &http.Client{Timeout: timeout, Transport: t}
}
