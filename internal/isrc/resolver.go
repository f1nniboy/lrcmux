package isrc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/f1nniboy/lrcmux/internal/cache"
	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/utils"
)

type ResolveInput struct {
	Artist    string
	Title     string
	Album     string
	Duration  int64
	ISRC      string
	CacheOnly bool
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

func (r *Resolver) Resolve(ctx context.Context, in ResolveInput) (lyrics.Track, error) {
	if in.ISRC != "" {
		return r.resolveByISRC(ctx, in.ISRC, in.CacheOnly)
	}
	return r.resolveBySearch(ctx, in)
}

func (r *Resolver) resolveByISRC(ctx context.Context, isrc string, cacheOnly bool) (lyrics.Track, error) {
	key := metaKey(isrc)
	switch track, status, _ := cache.Get[lyrics.Track](ctx, r.cache, key); status {
	case cache.Found:
		r.log.Debug("isrc meta cache hit", "isrc", isrc)
		return track, nil
	case cache.KnownMiss:
		r.log.Debug("isrc meta negative cache hit", "isrc", isrc)
		return lyrics.Track{}, lyrics.ErrNotFound
	}

	if cacheOnly {
		return lyrics.Track{}, lyrics.ErrNotFound
	}

	track, err := r.lookupMeta(ctx, isrc)
	if errors.Is(err, lyrics.ErrNotFound) {
		r.log.Debug("isrc not found on deezer", "isrc", isrc)
		cache.SetMiss(context.Background(), r.cache, key, r.missTTL)
		return lyrics.Track{}, lyrics.ErrNotFound
	}
	if err != nil {
		r.log.Warn("isrc meta lookup failed", "isrc", isrc, "err", err)
		return lyrics.Track{}, err
	}

	go func() {
		bg := context.Background()
		cache.Set(bg, r.cache, key, track, 0)
		cache.Set(bg, r.cache, trackKey(track.Artist, track.Title), isrc, 0)
	}()

	r.log.Debug("isrc meta resolved", "isrc", isrc, "artist", track.Artist, "title", track.Title)
	return track, nil
}

func (r *Resolver) resolveBySearch(ctx context.Context, in ResolveInput) (lyrics.Track, error) {
	key := trackKey(in.Artist, in.Title)
	switch isrc, status, _ := cache.Get[string](ctx, r.cache, key); status {
	case cache.Found:
		return r.resolveByISRC(ctx, isrc, in.CacheOnly)
	case cache.KnownMiss:
		r.log.Debug("isrc negative cache hit", "artist", in.Artist, "title", in.Title)
		return lyrics.Track{}, lyrics.ErrNotFound
	}

	if in.CacheOnly {
		return lyrics.Track{}, lyrics.ErrNotFound
	}

	sfKey := utils.Normalize(in.Artist) + ":" + utils.Normalize(in.Title)
	v, err, _ := r.group.Do(sfKey, func() (any, error) {
		track, err := r.lookup(ctx, in)
		if err != nil && !errors.Is(err, lyrics.ErrNotFound) {
			r.log.Warn("isrc lookup failed", "artist", in.Artist, "title", in.Title, "err", err)
			return lyrics.Track{}, err
		}

		if errors.Is(err, lyrics.ErrNotFound) {
			if primary := utils.PrimaryArtist(in.Artist); primary != "" && utils.Normalize(primary) != utils.Normalize(in.Artist) {
				primaryIn := in
				primaryIn.Artist = primary
				track, err = r.lookup(ctx, primaryIn)
			}
		}

		if err != nil {
			r.log.Debug("isrc not found", "artist", in.Artist, "title", in.Title)
			go cache.SetMiss(context.Background(), r.cache, key, r.missTTL)
			return lyrics.Track{}, lyrics.ErrNotFound
		}

		go func() {
			bg := context.Background()
			cache.Set(bg, r.cache, key, track.ISRC, 0)
			cache.Set(bg, r.cache, metaKey(track.ISRC), track, 0)
		}()

		r.log.Debug("isrc resolved", "artist", in.Artist, "title", in.Title, "isrc", track.ISRC)
		return track, nil
	})
	if err != nil {
		return lyrics.Track{}, err
	}
	return v.(lyrics.Track), nil
}

func trackKey(artist, title string) string {
	s := utils.Normalize(artist) + ":" + utils.Normalize(title)
	sum := sha256.Sum256([]byte(s))
	return "track2isrc:" + hex.EncodeToString(sum[:16])
}

func metaKey(isrc string) string {
	return "isrc2track:" + isrc
}
