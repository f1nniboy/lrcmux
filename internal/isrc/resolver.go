package isrc

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
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

func (r *Resolver) Resolve(ctx context.Context, in ResolveInput) (lyrics.Track, error) {
	if in.ISRC != "" {
		return r.resolveByISRC(ctx, in.ISRC)
	}
	return r.resolveBySearch(ctx, in)
}

func (r *Resolver) resolveByISRC(ctx context.Context, isrc string) (lyrics.Track, error) {
	key := metaKey(isrc)
	switch track, status, _ := cache.Get[lyrics.Track](ctx, r.cache, key); status {
	case cache.Hit:
		r.log.Debug("isrc meta cache hit", "isrc", isrc)
		return track, nil
	case cache.KnownMiss:
		r.log.Debug("isrc meta negative cache hit", "isrc", isrc)
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

	go cache.Set(context.Background(), r.cache, key, track, 0)

	r.log.Debug("isrc meta resolved", "isrc", isrc, "artist", track.Artist, "title", track.Title)
	return track, nil
}

func (r *Resolver) resolveBySearch(ctx context.Context, in ResolveInput) (lyrics.Track, error) {
	key := normalize.String(in.Artist) + ":" + normalize.String(in.Title) + ":" + strconv.FormatInt(in.Duration, 10)

	v, err, _ := r.group.Do(key, func() (any, error) {
		track, err := r.lookup(ctx, in)
		if err != nil {
			if !errors.Is(err, lyrics.ErrNotFound) {
				r.log.Warn("isrc lookup failed", "artist", in.Artist, "title", in.Title, "err", err)
				return lyrics.Track{}, err
			}
			r.log.Debug("isrc not found", "artist", in.Artist, "title", in.Title)
			return lyrics.Track{}, lyrics.ErrNotFound
		}

		go cache.Set(context.Background(), r.cache, metaKey(track.ISRC), track, 0)

		r.log.Debug("isrc resolved", "artist", in.Artist, "title", in.Title, "isrc", track.ISRC)
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
