package isrc

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/f1nniboy/lrcmux/internal/cache"
	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

const (
	noData = `{"error":{"type":"DataException","message":"no data","code":800}}`
	quota  = `{"error":{"type":"Exception","message":"Quota limit exceeded","code":4}}`
)

// keeps the real deezer urls, pointed at the test server
type redirect struct{ target *url.URL }

func (rt redirect) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.URL.Scheme, req.URL.Host = rt.target.Scheme, rt.target.Host
	return http.DefaultTransport.RoundTrip(req)
}

func reply(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, body) }
}

func testResolver(t *testing.T, h http.HandlerFunc) (*Resolver, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		h(w, r)
	}))
	t.Cleanup(srv.Close)
	target, _ := url.Parse(srv.URL)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(&http.Client{Transport: redirect{target}}, cache.NewMemory(), time.Hour, log), &calls
}

func TestResolveCaching(t *testing.T) {
	byISRC, bySearch := ResolveInput{ISRC: "XX0000000000"}, ResolveInput{Artist: "Artist", Title: "Song"}
	tests := []struct {
		err   error
		name  string
		body  string
		in    ResolveInput
		calls int32
	}{
		{name: "an isrc hit", in: byISRC, body: `{"id":1,"isrc":"XX0000000000","title":"Song","artist":{"name":"Artist"}}`, calls: 1},
		{name: "an isrc miss", in: byISRC, body: noData, err: lyrics.ErrNotFound, calls: 1},
		{name: "a quota error", in: byISRC, body: quota, err: providers.ErrRateLimited, calls: 2},
		{name: "a search hit", in: bySearch, body: `{"data":[{"id":1,"isrc":"XX0000000000","title":"Song","artist":{"name":"Artist"}}]}`, calls: 1},
		{name: "a search miss", in: bySearch, body: `{"data":[],"total":0}`, err: lyrics.ErrNotFound, calls: 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, calls := testResolver(t, reply(tc.body))
			for range 2 {
				track, err := r.Resolve(t.Context(), tc.in)
				if tc.err == nil && (err != nil || track.ISRC != "XX0000000000") || !errors.Is(err, tc.err) {
					t.Fatalf("got %q, %v, want %v", track.ISRC, err, tc.err)
				}
			}
			if calls.Load() != tc.calls {
				t.Errorf("deezer calls = %d, want %d", calls.Load(), tc.calls)
			}
			if _, hit, _ := r.Cached(t.Context(), tc.in); hit != (tc.calls == 1) {
				t.Errorf("cached = %v after %d deezer calls", hit, tc.calls)
			}
		})
	}
}

// a search gets body, a track its credits, an artist search artist
func deezer(body, credits, artist string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/search/artist":
			fmt.Fprint(w, artist)
		case strings.HasPrefix(r.URL.Path, "/track/"):
			fmt.Fprint(w, credits)
		default:
			fmt.Fprint(w, body)
		}
	}
}

func TestResolveReadings(t *testing.T) {
	cases := []struct {
		name, body, want string
		in               ResolveInput
	}{
		{
			name: "a dash that is part of the title",
			body: `{"data":[{"id":1,"isrc":"XX0000000001","title":"Song - Subtitle","artist":{"name":"Artist"}}]}`,
			in:   ResolveInput{Artist: "Artist", Title: "Song - Subtitle"},
			want: "XX0000000001",
		},
		{
			name: "an upload named after artist and title",
			body: `{"data":[{"id":1,"isrc":"XX0000000001","title":"Song","artist":{"name":"Artist"}}]}`,
			in:   ResolveInput{Artist: "Uploader", Title: "Artist - Song (Official Video)"},
			want: "XX0000000001",
		},
		{
			name: "the title as written wins over a reinterpretation, whatever the duration",
			body: `{"data":[{"id":1,"isrc":"XX0000000001","title":"Song - Subtitle","artist":{"name":"Artist"},"duration":260},` +
				`{"id":2,"isrc":"XX0000000002","title":"Song","artist":{"name":"Artist"},"duration":200}]}`,
			in:   ResolveInput{Artist: "Artist", Title: "Song - Subtitle", Duration: 200},
			want: "XX0000000001",
		},
		{
			name: "a version after the title",
			body: `{"data":[{"id":1,"isrc":"XX0000000001","title":"Song (Remastered)","artist":{"name":"Artist"}}]}`,
			in:   ResolveInput{Artist: "Artist", Title: "Song - Remastered"},
			want: "XX0000000001",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r, _ := testResolver(t, deezer(c.body, `{}`, `{"data":[]}`))
			if track, err := r.Resolve(t.Context(), c.in); err != nil || track.ISRC != c.want {
				t.Fatalf("got %q, %v, want %q", track.ISRC, err, c.want)
			}
		})
	}
}

func TestResolveArtistByID(t *testing.T) {
	body := `{"data":[{"id":1,"isrc":"XX0000000001","title":"Song","artist":{"id":7,"name":"Another Name"}}]}`
	in := ResolveInput{Artist: "Artist", Title: "Song"}

	t.Run("deezer naming the same artist confirms it", func(t *testing.T) {
		r, _ := testResolver(t, deezer(body, noData, `{"data":[{"id":7,"name":"Another Name"}]}`))
		if track, err := r.Resolve(t.Context(), in); err != nil || track.ISRC != "XX0000000001" {
			t.Fatalf("got %q, %v", track.ISRC, err)
		}
	})

	t.Run("deezer naming someone else does not", func(t *testing.T) {
		r, _ := testResolver(t, deezer(body, `{}`, `{"data":[{"id":8,"name":"Somebody Else"}]}`))
		if _, err := r.Resolve(t.Context(), in); !errors.Is(err, lyrics.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})
}

func TestResolveCoArtist(t *testing.T) {
	body := `{"data":[{"id":1,"isrc":"XX0000000001","title":"Song","artist":{"id":7,"name":"Another Name"}}]}`
	credits := `{"id":1,"isrc":"XX0000000001","title":"Song","artist":{"id":7,"name":"Another Name"},"contributors":[{"id":7,"name":"Another Name"},{"id":8,"name":"Artist"}]}`
	r, calls := testResolver(t, deezer(body, credits, quota))
	if track, err := r.Resolve(t.Context(), ResolveInput{Artist: "Artist", Title: "Song"}); err != nil || track.ISRC != "XX0000000001" {
		t.Fatalf("got %q, %v", track.ISRC, err)
	}
	if n := calls.Load(); n != 2 {
		t.Errorf("deezer calls = %d, want 2, the artist is never searched", n)
	}
}

func TestResolveFallbackFails(t *testing.T) {
	body := `{"data":[{"id":1,"isrc":"XX0000000001","title":"Song","artist":{"id":7,"name":"Another Name"}}]}`
	cases := []struct {
		name, credits, artist string
		calls                 int32
	}{
		{name: "the credits", credits: quota, artist: `{"data":[]}`, calls: 2},
		{name: "the artist search", credits: `{}`, artist: quota, calls: 3},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r, calls := testResolver(t, deezer(body, c.credits, c.artist))
			in := ResolveInput{Artist: "Artist", Title: "Song"}
			for range 2 {
				if _, err := r.Resolve(t.Context(), in); err == nil || errors.Is(err, lyrics.ErrNotFound) {
					t.Fatalf("err = %v, want a non-not-found error", err)
				}
			}
			if n := calls.Load(); n != 2*c.calls {
				t.Errorf("deezer calls = %d, want %d, a failure is never cached", n, 2*c.calls)
			}
		})
	}
}

func TestResolveEmptyInput(t *testing.T) {
	r, calls := testResolver(t, reply(noData))
	for _, in := range []ResolveInput{{Title: "Song"}, {Artist: "Artist", Title: "(feat. Another Artist)"}} {
		if _, err := r.Resolve(t.Context(), in); !errors.Is(err, lyrics.ErrNotFound) {
			t.Fatalf("%+v: err = %v, want ErrNotFound", in, err)
		}
	}
	if n := calls.Load(); n != 0 {
		t.Errorf("deezer calls = %d, want 0", n)
	}
}
