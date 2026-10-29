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

// long enough to spare deezer on repeat requests, finite since the key holds
// client chosen fields and so matcher fixes reach old lookups
const searchTTL = 30 * 24 * time.Hour

// deezer answers a malformed isrc with a bare 403 or 404 instead of its
// not-found error, which would never be cached
var isrcRE = regexp.MustCompile(`^[A-Z]{2}[A-Z0-9]{3}[0-9]{7}$`)

func (r *Resolver) Resolve(ctx context.Context, in ResolveInput) (lyrics.Track, error) {
	if in.ISRC != "" {
		isrc := strings.ToUpper(strings.ReplaceAll(in.ISRC, "-", ""))
		if !isrcRE.MatchString(isrc) {
			return lyrics.Track{}, lyrics.ErrNotFound
		}
		return r.cached(ctx, r.log.With("isrc", isrc), metaKey(isrc), 0, func(ctx context.Context) (lyrics.Track, bool, error) {
			track, err := r.lookupMeta(ctx, isrc)
			return track, true, err
		})
	}
	in.Artist, in.Title = normalize.Query(in.Artist, in.Title)
	// a title alone would match anyone's song of that name
	if in.Artist == "" {
		return lyrics.Track{}, lyrics.ErrNotFound
	}
	return r.cached(ctx, r.log.With("artist", in.Artist, "title", in.Title), searchKey(in), searchTTL, func(ctx context.Context) (lyrics.Track, bool, error) {
		return r.lookup(ctx, in)
	})
}

// serves key from the cache, or runs lookup once for every concurrent caller,
// keeping complete hits for ttl and clean misses for missTTL
func (r *Resolver) cached(ctx context.Context, log *slog.Logger, key string, ttl time.Duration, lookup func(context.Context) (lyrics.Track, bool, error)) (lyrics.Track, error) {
	switch track, status, _ := cache.Get[lyrics.Track](ctx, r.cache, key); status {
	case cache.Hit:
		log.Debug("isrc cache hit")
		return track, nil
	case cache.KnownMiss:
		log.Debug("isrc negative cache hit")
		return lyrics.Track{}, lyrics.ErrNotFound
	}

	v, err, _ := r.group.Do(key, func() (any, error) {
		// shared by every waiter, so one caller going away mustn't fail the rest
		track, complete, err := lookup(context.WithoutCancel(ctx))
		if errors.Is(err, lyrics.ErrNotFound) {
			log.Debug("isrc not found")
			cache.SetMiss(context.Background(), r.cache, key, r.missTTL)
			return lyrics.Track{}, err
		}
		if err != nil {
			log.Warn("isrc lookup failed", "err", err)
			return lyrics.Track{}, err
		}
		if complete {
			go cache.Set(context.Background(), r.cache, key, track, ttl)
		}
		log.Debug("isrc resolved", "isrc", track.ISRC, "complete", complete)
		return track, nil
	})
	if err != nil {
		return lyrics.Track{}, err
	}
	return v.(lyrics.Track), nil
}

func metaKey(isrc string) string {
	return "isrc2track:" + isrc
}

// album is part of the key since pickBest scores on it, and the parts are
// hashed since they are client input of any length
func searchKey(in ResolveInput) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		normalize.String(in.Artist),
		normalize.String(in.Title),
		normalize.String(in.Album),
		strconv.FormatInt(in.Duration, 10),
	}, "\x00")))
	return "search2track:" + hex.EncodeToString(sum[:16])
}
