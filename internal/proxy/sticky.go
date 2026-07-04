package proxy

import (
	"context"
	"net/url"
	"sync"
)

type stickyKey struct{}

type stickyEntry struct {
	mu  sync.Mutex
	url *url.URL
}

func Sticky(ctx context.Context) context.Context {
	return context.WithValue(ctx, stickyKey{}, &stickyEntry{})
}

func (p *Pool) pick(ctx context.Context) *url.URL {
	if e, ok := ctx.Value(stickyKey{}).(*stickyEntry); ok {
		e.mu.Lock()
		defer e.mu.Unlock()
		if e.url == nil {
			e.url = p.next()
		}
		return e.url
	}
	return p.next()
}
