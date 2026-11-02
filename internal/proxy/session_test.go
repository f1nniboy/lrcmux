package proxy

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"
)

// a stand-in proxy that answers every request itself and counts the connections it gets
func countingProxy(t *testing.T) (*Pool, *atomic.Int32) {
	var conns atomic.Int32
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	srv.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			conns.Add(1)
		}
	}
	srv.Start()
	t.Cleanup(srv.Close)
	u, _ := url.Parse(srv.URL)
	return &Pool{urls: []*url.URL{u}}, &conns
}

func TestSession(t *testing.T) {
	tests := []struct {
		name    string
		session bool
		want    int32
	}{
		{name: "requests in a session share a connection", session: true, want: 1},
		{name: "requests outside a session each get one", want: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool, conns := countingProxy(t)
			c := &http.Client{Timeout: 5 * time.Second, Transport: &poolTransport{pool: pool, fresh: pool.transport(nil)}}
			ctx := t.Context()
			if tt.session {
				var end func()
				ctx, end = WithSession(ctx)
				defer end()
			}
			for range 3 {
				get(ctx, t, c)
			}
			if got := conns.Load(); got != tt.want {
				t.Errorf("connections = %d, want %d", got, tt.want)
			}
		})
	}
}

func get(ctx context.Context, t *testing.T, c *http.Client) {
	t.Helper()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://upstream.invalid/", nil)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}
