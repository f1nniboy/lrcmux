package musixmatch

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestSign(t *testing.T) {
	const raw = "https://www.musixmatch.com/ws/1.1/track.lyrics.get?app_id=community-app-v1.0&format=json&track_isrc=XX0000000000"
	// reference from the implementation verified against the live api
	const want = raw + "&signature=ag%2BtwtauIB3A8iJGO8xnjA8umuKJ%2BC45cdrX9jPMDOM%3D&signature_protocol=sha256"

	// 01:00 on jan 3 at utc+2 is still jan 2 in utc, which is what gets signed
	now := time.Date(2026, 1, 3, 1, 0, 0, 0, time.FixedZone("", 2*60*60))
	if got := sign(raw, "0123456789abcdef0123456789abcdef", now); got != want {
		t.Errorf("sign() =\n%s\nwant\n%s", got, want)
	}
}

func TestExtractKey(t *testing.T) {
	js := `a=Buffer.from("not a key".split(""));` +
		`b=Buffer.from("=YWZkNmYhlDO3YTN0MjMxAjZlR2YiFWO4cjN1QzMyEDM".split("").reverse().join(""),"base64")`
	if key, ok := extractKey(js); !ok || key != "0123456789abcdef0123456789abcdef" {
		t.Errorf("extractKey() = %q, %v", key, ok)
	}
	minified := `a=Buffer.from("deadbeef","hex");` +
		`b=Buffer.from("=YWZkNmYhlDO3YTN0MjMxAjZlR2YiFWO4cjN1QzMyEDM".split("").reverse().join(""),"base64")`
	if key, ok := extractKey(minified); !ok || key != "0123456789abcdef0123456789abcdef" {
		t.Errorf("extractKey(minified) = %q, %v", key, ok)
	}
	if _, ok := extractKey(`c=Buffer.from("x".split(""))`); ok {
		t.Error("extractKey() found a key where there is none")
	}
}

// sends every request to the test server, so the signer keeps its real urls
type redirect struct{ target *url.URL }

func (rt redirect) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.URL.Scheme, req.URL.Host = rt.target.Scheme, rt.target.Host
	return http.DefaultTransport.RoundTrip(req)
}

func TestRefreshLeavesAtDeadline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		time.Sleep(300 * time.Millisecond)
	}))
	t.Cleanup(srv.Close)
	target, _ := url.Parse(srv.URL)
	s := &signer{client: &http.Client{Transport: redirect{target}}}

	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancel()
	start := time.Now()
	if _, err := s.refresh(ctx, ""); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want DeadlineExceeded", err)
	}
	if d := time.Since(start); d > 150*time.Millisecond {
		t.Errorf("refresh took %v, want it to leave at its deadline", d)
	}
}
