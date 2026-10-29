package isrc

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/f1nniboy/lrcmux/internal/cache"
	"github.com/f1nniboy/lrcmux/internal/lyrics"
)

const (
	noData = `{"error":{"type":"DataException","message":"no data","code":800}}`
	quota  = `{"error":{"type":"Exception","message":"Quota limit exceeded","code":4}}`
)

// sends every request to the test server, so the resolver keeps its real
// deezer urls
type redirect struct{ target *url.URL }

func (rt redirect) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.URL.Scheme, req.URL.Host = rt.target.Scheme, rt.target.Host
	return http.DefaultTransport.RoundTrip(req)
}

func reply(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, body) }
}

func testResolver(t *testing.T, h http.HandlerFunc) (*Resolver, cache.Cache, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		h(w, r)
	}))
	t.Cleanup(srv.Close)
	target, _ := url.Parse(srv.URL)
	c := cache.NewMemory()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(&http.Client{Transport: redirect{target}}, c, time.Hour, log), c, &calls
}

// cache writes on a hit are async
func waitHit(t *testing.T, c cache.Cache, key string) {
	t.Helper()
	for deadline := time.Now().Add(time.Second); time.Now().Before(deadline); time.Sleep(5 * time.Millisecond) {
		if _, status, _ := cache.Get[lyrics.Track](t.Context(), c, key); status == cache.Hit {
			return
		}
	}
	t.Fatalf("%s never got cached", key)
}

func TestResolveByISRC(t *testing.T) {
	in := ResolveInput{ISRC: "XX0000000000"}

	t.Run("missing isrc is a cached miss, not an empty track", func(t *testing.T) {
		r, _, calls := testResolver(t, reply(noData))
		for range 2 {
			if _, err := r.Resolve(t.Context(), in); !errors.Is(err, lyrics.ErrNotFound) {
				t.Fatalf("err = %v, want ErrNotFound", err)
			}
		}
		if n := calls.Load(); n != 1 {
			t.Errorf("deezer calls = %d, want 1", n)
		}
	})

	t.Run("quota error is never cached", func(t *testing.T) {
		r, _, calls := testResolver(t, reply(quota))
		for range 2 {
			if _, err := r.Resolve(t.Context(), in); err == nil || errors.Is(err, lyrics.ErrNotFound) {
				t.Fatalf("err = %v, want a non-not-found error", err)
			}
		}
		if n := calls.Load(); n != 2 {
			t.Errorf("deezer calls = %d, want 2", n)
		}
	})
}

func TestResolveBySearch(t *testing.T) {
	in := ResolveInput{Artist: "Artist", Title: "Song"}

	t.Run("hit is cached, and the artist: filter is skipped", func(t *testing.T) {
		var mu sync.Mutex
		var queries []string
		r, c, calls := testResolver(t, func(w http.ResponseWriter, req *http.Request) {
			mu.Lock()
			queries = append(queries, req.URL.Query().Get("q"))
			mu.Unlock()
			fmt.Fprint(w, `{"data":[{"id":1,"isrc":"XX0000000000","title":"Song","artist":{"name":"Artist"}}]}`)
		})

		if track, err := r.Resolve(t.Context(), in); err != nil || track.ISRC != "XX0000000000" {
			t.Fatalf("got %q, %v", track.ISRC, err)
		}
		slices.Sort(queries)
		if want := []string{"artist song", `track:"song"`}; !slices.Equal(queries, want) {
			t.Errorf("queries = %q, want %q", queries, want)
		}

		waitHit(t, c, searchKey(in))
		if _, err := r.Resolve(t.Context(), in); err != nil {
			t.Fatal(err)
		}
		if n := calls.Load(); n != 2 {
			t.Errorf("deezer calls = %d, want 2", n)
		}
	})

	t.Run("clean miss is cached", func(t *testing.T) {
		r, _, calls := testResolver(t, reply(`{"data":[],"total":0}`))
		for range 2 {
			if _, err := r.Resolve(t.Context(), in); !errors.Is(err, lyrics.ErrNotFound) {
				t.Fatalf("err = %v, want ErrNotFound", err)
			}
		}
		if n := calls.Load(); n != 2 {
			t.Errorf("deezer calls = %d, want 2", n)
		}
	})

	t.Run("a miss with a failed query is not cached", func(t *testing.T) {
		r, _, calls := testResolver(t, func(w http.ResponseWriter, req *http.Request) {
			if strings.HasPrefix(req.URL.Query().Get("q"), "track:") {
				fmt.Fprint(w, quota)
				return
			}
			fmt.Fprint(w, noData)
		})
		for range 2 {
			if _, err := r.Resolve(t.Context(), in); err == nil || errors.Is(err, lyrics.ErrNotFound) {
				t.Fatalf("err = %v, want a non-not-found error", err)
			}
		}
		if n := calls.Load(); n != 4 {
			t.Errorf("deezer calls = %d, want 4", n)
		}
	})
}

// every request gets body, except /search/artist, which gets artist
func deezer(body, artist string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/search/artist" {
			fmt.Fprint(w, artist)
			return
		}
		fmt.Fprint(w, body)
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
			name: "the title as written beats its bare reading, whatever the duration",
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
			r, _, _ := testResolver(t, deezer(c.body, `{"data":[]}`))
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
		r, _, _ := testResolver(t, deezer(body, `{"data":[{"id":7,"name":"Another Name"}]}`))
		if track, err := r.Resolve(t.Context(), in); err != nil || track.ISRC != "XX0000000001" {
			t.Fatalf("got %q, %v", track.ISRC, err)
		}
	})

	t.Run("deezer naming someone else does not", func(t *testing.T) {
		r, _, _ := testResolver(t, deezer(body, `{"data":[{"id":8,"name":"Somebody Else"}]}`))
		if _, err := r.Resolve(t.Context(), in); !errors.Is(err, lyrics.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})
}

func TestResolvePartialNotCached(t *testing.T) {
	r, _, calls := testResolver(t, func(w http.ResponseWriter, req *http.Request) {
		if strings.HasPrefix(req.URL.Query().Get("q"), "track:") {
			fmt.Fprint(w, quota)
			return
		}
		fmt.Fprint(w, `{"data":[{"id":1,"isrc":"XX0000000001","title":"Song","artist":{"name":"Artist"}}]}`)
	})
	in := ResolveInput{Artist: "Artist", Title: "Song"}
	for range 2 {
		if track, err := r.Resolve(t.Context(), in); err != nil || track.ISRC != "XX0000000001" {
			t.Fatalf("got %q, %v", track.ISRC, err)
		}
	}
	if n := calls.Load(); n != 4 {
		t.Errorf("deezer calls = %d, want 4", n)
	}
}

func TestResolveISRCInput(t *testing.T) {
	t.Run("a malformed isrc never reaches deezer", func(t *testing.T) {
		r, _, calls := testResolver(t, reply(noData))
		if _, err := r.Resolve(t.Context(), ResolveInput{ISRC: "XX/0000"}); !errors.Is(err, lyrics.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
		if n := calls.Load(); n != 0 {
			t.Errorf("deezer calls = %d, want 0", n)
		}
	})

	t.Run("case and dashes are normalized", func(t *testing.T) {
		var path string
		r, _, _ := testResolver(t, func(w http.ResponseWriter, req *http.Request) {
			path = req.URL.Path
			fmt.Fprint(w, `{"id":1,"isrc":"XX0000000000","title":"Song","artist":{"name":"Artist"}}`)
		})
		if _, err := r.Resolve(t.Context(), ResolveInput{ISRC: "xx-000-00-00000"}); err != nil {
			t.Fatal(err)
		}
		if path != "/track/isrc:XX0000000000" {
			t.Errorf("path = %q", path)
		}
	})
}

func TestResolveArtistAskedOnce(t *testing.T) {
	var asked atomic.Int32
	r, _, _ := testResolver(t, func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/search/artist" {
			asked.Add(1)
			fmt.Fprint(w, `{"data":[{"id":8,"name":"Somebody Else"}]}`)
			return
		}
		fmt.Fprint(w, `{"data":[{"id":1,"isrc":"XX0000000001","title":"Song - Subtitle","artist":{"id":7,"name":"Another Name"}},`+
			`{"id":2,"isrc":"XX0000000002","title":"Song","artist":{"id":7,"name":"Another Name"}}]}`)
	})
	if _, err := r.Resolve(t.Context(), ResolveInput{Artist: "Artist", Title: "Song - Subtitle"}); !errors.Is(err, lyrics.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
	if n := asked.Load(); n != 1 {
		t.Errorf("artist searches = %d, want 1", n)
	}
}

func TestResolveWithoutArtist(t *testing.T) {
	r, _, calls := testResolver(t, reply(noData))
	if _, err := r.Resolve(t.Context(), ResolveInput{Title: "Song"}); !errors.Is(err, lyrics.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
	if n := calls.Load(); n != 0 {
		t.Errorf("deezer calls = %d, want 0", n)
	}
}
