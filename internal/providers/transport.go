package providers

import (
	"net/http"

	"github.com/f1nniboy/lrcmux/internal/meta"
)

var UserAgent = meta.AppName + "/" + meta.Version + " (+https://" + meta.AppDomain + ")"

type uaTransport struct{ inner http.RoundTripper }

func (t uaTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Header.Get("User-Agent") == "" {
		clone := r.Clone(r.Context())
		clone.Header.Set("User-Agent", UserAgent)
		r = clone
	}
	return t.inner.RoundTrip(r)
}

func withUserAgent(c *http.Client) *http.Client {
	inner := c.Transport
	if inner == nil {
		inner = http.DefaultTransport
	}
	out := *c
	out.Transport = uaTransport{inner: inner}
	return &out
}
