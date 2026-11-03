package providers

import (
	"net/http"

	"github.com/f1nniboy/lrcmux/internal/meta"
)

// sent unless a request sets its own
const UserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:151.0) Gecko/20100101 Firefox/151.0"

// for open APIs that ask clients to identify themselves
var AppUserAgent = meta.AppName + "/" + meta.Version + " (https://" + meta.AppDomain + ")"

type uaTransport struct{ inner http.RoundTripper }

func (t uaTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Header.Get("User-Agent") == "" {
		clone := r.Clone(r.Context())
		clone.Header.Set("User-Agent", UserAgent)
		r = clone
	}
	return t.inner.RoundTrip(r)
}

func WithUserAgent(c *http.Client) *http.Client {
	inner := c.Transport
	if inner == nil {
		inner = http.DefaultTransport
	}
	out := *c
	out.Transport = uaTransport{inner: inner}
	return &out
}
