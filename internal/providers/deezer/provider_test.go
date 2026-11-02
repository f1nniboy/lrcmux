package deezer

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"github.com/f1nniboy/lrcmux/internal/cache"
	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

const (
	foundBody   = `{"data":{"track":{"lyrics":{` + wordLines + `}}}}`
	expiredBody = `{"errors":[{"message":"expired","extensions":{"type":"JwtTokenExpiredError"}}],"data":{"track":null}}`
)

func lyricsError(typ string) string {
	return `{"errors":[{"message":"no","extensions":{"type":"` + typ + `"}}],"data":{"track":{"lyrics":null}}}`
}

// sends every request to the test server, so the provider keeps its real urls
type redirect struct{ target *url.URL }

func (rt redirect) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.URL.Scheme, req.URL.Host = rt.target.Scheme, rt.target.Host
	return http.DefaultTransport.RoundTrip(req)
}

// mints tokens on the auth path and answers the n-th lyrics request with pipe
func testProvider(t *testing.T, pipe func(w http.ResponseWriter, n int32)) (*Provider, *atomic.Int32) {
	t.Helper()
	var mints, reqs atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login/anonymous" {
			fmt.Fprintf(w, `{"jwt":"token-%d"}`, mints.Add(1))
			return
		}
		var body struct {
			Variables struct {
				TrackID string `json:"trackId"`
			} `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Variables.TrackID != "7" || r.Header.Get("Authorization") == "" {
			t.Errorf("bad lyrics request: track %q, %v", body.Variables.TrackID, err)
		}
		pipe(w, reqs.Add(1))
	}))
	t.Cleanup(srv.Close)
	target, _ := url.Parse(srv.URL)
	p := &Provider{}
	p.SetDeps(&http.Client{Transport: redirect{target}}, cache.NewMemory(), nil)
	p.Init()
	return p, &mints
}

func reply(bodies ...string) func(http.ResponseWriter, int32) {
	return func(w http.ResponseWriter, n int32) { fmt.Fprint(w, bodies[min(int(n), len(bodies))-1]) }
}

func TestSearch(t *testing.T) {
	q := lyrics.Query{Track: lyrics.Track{Title: "Song", Artist: "Artist", Duration: 200, DeezerID: 7}}
	cases := []struct {
		pipe  func(http.ResponseWriter, int32)
		name  string
		want  string
		mints int32
	}{
		{name: "found", pipe: reply(foundBody), want: "word", mints: 1},
		{name: "no lyrics", pipe: reply(lyricsError("LyricsNotFoundError")), want: "not found", mints: 1},
		{name: "an expired token is renewed", pipe: reply(expiredBody, foundBody), want: "word", mints: 2},
		{name: "a token rejected twice", pipe: reply(expiredBody), want: "error", mints: 2},
		{
			name:  "another error beside lyrics",
			pipe:  reply(`{"errors":[{"message":"no","extensions":{"type":"SomethingError"}}],"data":{"track":{"lyrics":{` + wordLines + `}}}}`),
			want:  "error",
			mints: 1,
		},
		{name: "json without data", pipe: reply(`{}`), want: "error", mints: 1},
		{
			name:  "blocked",
			pipe:  func(w http.ResponseWriter, _ int32) { w.WriteHeader(http.StatusForbidden) },
			want:  "rate limited",
			mints: 1,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p, mints := testProvider(t, c.pipe)
			res, err := p.Search(t.Context(), q)
			var got string
			switch {
			case err == nil:
				got = res.SyncLevel.String()
			case errors.Is(err, lyrics.ErrNotFound):
				got = "not found"
			case errors.Is(err, providers.ErrRateLimited):
				got = "rate limited"
			default:
				got = "error"
			}
			if got != c.want || mints.Load() != c.mints {
				t.Errorf("got %s (%v) after %d mints, want %s after %d", got, err, mints.Load(), c.want, c.mints)
			}
		})
	}
}
