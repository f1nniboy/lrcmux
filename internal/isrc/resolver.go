package isrc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/f1nniboy/lrcmux/internal/cache"
	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/normalize"
)

type ResolveInput struct {
	Artist   string
	Title    string
	Album    string
	ISRC     string
	Duration int64
}

type Resolver struct {
	group   singleflight.Group
	cache   cache.Cache
	client  *http.Client
	log     *slog.Logger
	missTTL time.Duration
}

func New(client *http.Client, c cache.Cache, missTTL time.Duration, log *slog.Logger) *Resolver {
	return &Resolver{client: client, cache: c, missTTL: missTTL, log: log}
}

const Timeout = 3 * time.Second

// finite so matcher fixes reach old lookups
const searchTTL = 30 * 24 * time.Hour

// deezer answers a malformed isrc with a bare 403 or 404, which would never cache as a miss
var isrcRE = regexp.MustCompile(`^[A-Z]{2}[A-Z0-9]{3}[0-9]{7}$`)

type route struct {
	log    *slog.Logger
	lookup func(context.Context) (lyrics.Track, error)
	key    string
	ttl    time.Duration
}

func (r *Resolver) routeFor(in ResolveInput) (route, bool) {
	if in.ISRC != "" {
		isrc := strings.ToUpper(strings.ReplaceAll(in.ISRC, "-", ""))
		if !isrcRE.MatchString(isrc) {
			return route{}, false
		}
		return route{log: r.log.With("isrc", isrc), key: metaKey(isrc), lookup: func(ctx context.Context) (lyrics.Track, error) {
			t, err := r.track(ctx, "isrc:"+isrc)
			return t.toTrack(), err
		}}, true
	}
	in.Artist, in.Title = normalize.Query(in.Artist, in.Title)
	// a title alone would match anyone's song of that name
	if in.Artist == "" || in.Title == "" {
		return route{}, false
	}
	return route{log: r.log.With("artist", in.Artist, "title", in.Title), key: searchKey(in), ttl: searchTTL, lookup: func(ctx context.Context) (lyrics.Track, error) {
		return r.lookup(ctx, in)
	}}, true
}

// a known miss and an input that can never resolve are hits too
func (r *Resolver) Cached(ctx context.Context, in ResolveInput) (lyrics.Track, bool, error) {
	rt, ok := r.routeFor(in)
	if !ok {
		return lyrics.Track{}, true, lyrics.ErrNotFound
	}
	return r.fromCache(ctx, rt)
}

func (r *Resolver) Resolve(ctx context.Context, in ResolveInput) (lyrics.Track, error) {
	rt, ok := r.routeFor(in)
	if !ok {
		return lyrics.Track{}, lyrics.ErrNotFound
	}
	if track, hit, err := r.fromCache(ctx, rt); hit {
		return track, err
	}

	v, err, _ := r.group.Do(rt.key, func() (any, error) {
		track, err := rt.lookup(context.WithoutCancel(ctx))
		if errors.Is(err, lyrics.ErrNotFound) {
			rt.log.Debug("isrc not found")
			cache.SetMiss(context.Background(), r.cache, rt.key, r.missTTL)
			return lyrics.Track{}, err
		}
		if err != nil {
			rt.log.Warn("isrc lookup failed", "err", err)
			return lyrics.Track{}, err
		}
		// not async, so the next request finds it
		cache.Set(context.Background(), r.cache, rt.key, track, rt.ttl)
		rt.log.Debug("isrc resolved", "isrc", track.ISRC)
		return track, nil
	})
	if err != nil {
		return lyrics.Track{}, err
	}
	return v.(lyrics.Track), nil
}

func (r *Resolver) fromCache(ctx context.Context, rt route) (lyrics.Track, bool, error) {
	switch track, status, _ := cache.Get[lyrics.Track](ctx, r.cache, rt.key); status {
	case cache.Hit:
		rt.log.Debug("isrc cache hit")
		return track, true, nil
	case cache.KnownMiss:
		rt.log.Debug("isrc negative cache hit")
		return lyrics.Track{}, true, lyrics.ErrNotFound
	default:
		return lyrics.Track{}, false, nil
	}
}

func metaKey(isrc string) string {
	return "isrc2track:" + isrc
}

// hashed since the parts are client input of any length
func searchKey(in ResolveInput) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		normalize.String(in.Artist),
		normalize.String(in.Title),
		normalize.String(in.Album),
		strconv.FormatInt(in.Duration, 10),
	}, "\x00")))
	return "search2track:" + hex.EncodeToString(sum[:16])
}
