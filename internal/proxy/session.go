package proxy

import (
	"context"
	"net/http"
	"sync"
)

type sessionKey struct{}

// one lyrics lookup shares a proxy and its connections, so a provider making many
// requests pays the handshakes once, while separate lookups still spread across the pool
type session struct {
	transports map[*Pool]*http.Transport
	mu         sync.Mutex
	closed     bool
}

func WithSession(ctx context.Context) (context.Context, func()) {
	s := &session{transports: make(map[*Pool]*http.Transport)}
	return context.WithValue(ctx, sessionKey{}, s), s.close
}

// nil once the session ended, so work that outlives it falls back to fresh connections
func (s *session) transport(p *Pool) *http.Transport {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	t, ok := s.transports[p]
	if !ok {
		t = p.transport(p.pick())
		s.transports[p] = t
	}
	return t
}

// a request still running keeps its connection until the transport's idle timeout
func (s *session) close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	for _, t := range s.transports {
		t.CloseIdleConnections()
	}
}
