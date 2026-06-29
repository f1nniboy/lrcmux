package isrc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/f1nniboy/lrcmux/internal/cache"
	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/utils"
)

const lookupBase = "https://api.deezer.com/search"

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

type TrackMeta struct {
	Artist   string
	Title    string
	Album    string
	Duration int64
}

// looks up an ISRC for the given artist and title via Deezer
func (r *Resolver) Resolve(ctx context.Context, q lyrics.Query) (string, error) {
	key := trackKey(q.Artist, q.Title)
	switch isrc, status, _ := cache.Get[string](ctx, r.cache, key); status {
	case cache.Found:
		r.log.Debug("isrc cache hit", "artist", q.Artist, "title", q.Title, "isrc", isrc)
		return isrc, nil
	case cache.KnownMiss:
		r.log.Debug("isrc negative cache hit", "artist", q.Artist, "title", q.Title)
		return "", lyrics.ErrNotFound
	}

	sfKey := utils.Normalize(q.Artist) + ":" + utils.Normalize(q.Title)
	v, err, _ := r.group.Do(sfKey, func() (any, error) {
		lookupCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		isrc, meta, err := r.lookup(lookupCtx, q.Artist, q.Title)
		if err != nil {
			r.log.Warn("isrc lookup failed", "artist", q.Artist, "title", q.Title, "err", err)
			return "", err
		}
		if isrc == "" {
			if primary := utils.PrimaryArtist(q.Artist); primary != "" && utils.Normalize(primary) != utils.Normalize(q.Artist) {
				isrc, meta, err = r.lookup(lookupCtx, primary, q.Title)
				if err != nil {
					r.log.Warn("isrc lookup failed (primary artist)", "artist", primary, "title", q.Title, "err", err)
				}
			}
		}
		if isrc == "" {
			r.log.Debug("isrc not found on deezer", "artist", q.Artist, "title", q.Title)
			go cache.SetMiss(context.Background(), r.cache, key, r.missTTL)
			return "", lyrics.ErrNotFound
		}

		go func() {
			bg := context.Background()
			cache.Set(bg, r.cache, key, isrc, 0)
			cache.Set(bg, r.cache, metaKey(isrc), meta, 0)
		}()

		r.log.Debug("isrc resolved", "artist", q.Artist, "title", q.Title, "isrc", isrc)
		return isrc, nil
	})
	if err != nil {
		return "", err
	}
	return v.(string), nil
}

// resolves track metadata from an ISRC via Deezer
func (r *Resolver) LookupMeta(ctx context.Context, isrc string) (TrackMeta, bool) {
	key := metaKey(isrc)
	switch meta, status, _ := cache.Get[TrackMeta](ctx, r.cache, key); status {
	case cache.Found:
		r.log.Debug("isrc meta cache hit", "isrc", isrc)
		return meta, true
	case cache.KnownMiss:
		r.log.Debug("isrc meta negative cache hit", "isrc", isrc)
		return TrackMeta{}, false
	}

	lookupCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	meta, err := r.lookupMeta(lookupCtx, isrc)
	if errors.Is(err, lyrics.ErrNotFound) {
		r.log.Debug("isrc not found on deezer", "isrc", isrc)
		cache.SetMiss(context.Background(), r.cache, key, r.missTTL)
		return TrackMeta{}, false
	}
	if err != nil {
		r.log.Warn("isrc meta lookup failed", "isrc", isrc, "err", err)
		return TrackMeta{}, false
	}

	go func() {
		bg := context.Background()
		cache.Set(bg, r.cache, key, meta, 0)
		cache.Set(bg, r.cache, trackKey(meta.Artist, meta.Title), isrc, 0)
	}()

	r.log.Debug("isrc meta resolved", "isrc", isrc, "artist", meta.Artist, "title", meta.Title)
	return meta, true
}

type deezerSearchResponse struct {
	Data []deezerTrack `json:"data"`
}

type deezerTrack struct {
	ISRC     string       `json:"isrc"`
	Title    string       `json:"title"`
	Duration int64        `json:"duration"`
	Artist   deezerArtist `json:"artist"`
	Album    deezerAlbum  `json:"album"`
}

type deezerArtist struct {
	Name string `json:"name"`
}

type deezerAlbum struct {
	Title string `json:"title"`
}

func (r *Resolver) lookup(ctx context.Context, artist, title string) (string, TrackMeta, error) {
	q := fmt.Sprintf(`artist:"%s" track:"%s"`, artist, title)
	endpoint := lookupBase + "?q=" + url.QueryEscape(q) + "&limit=1"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", TrackMeta{}, err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return "", TrackMeta{}, fmt.Errorf("deezer request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", TrackMeta{}, fmt.Errorf("deezer status %d", resp.StatusCode)
	}

	var dr deezerSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&dr); err != nil {
		return "", TrackMeta{}, fmt.Errorf("deezer decode: %w", err)
	}

	if len(dr.Data) == 0 {
		return "", TrackMeta{}, nil
	}
	t := dr.Data[0]
	return t.ISRC, TrackMeta{Artist: t.Artist.Name, Title: t.Title, Album: t.Album.Title, Duration: t.Duration}, nil
}

func (r *Resolver) lookupMeta(ctx context.Context, isrc string) (TrackMeta, error) {
	endpoint := "https://api.deezer.com/track/isrc:" + isrc

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return TrackMeta{}, err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return TrackMeta{}, fmt.Errorf("deezer request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return TrackMeta{}, lyrics.ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return TrackMeta{}, fmt.Errorf("deezer status %d", resp.StatusCode)
	}

	var t deezerTrack
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return TrackMeta{}, fmt.Errorf("deezer decode: %w", err)
	}

	return TrackMeta{Artist: t.Artist.Name, Title: t.Title, Album: t.Album.Title, Duration: t.Duration}, nil
}

func trackKey(artist, title string) string {
	s := utils.Normalize(artist) + ":" + utils.Normalize(title)
	sum := sha256.Sum256([]byte(s))
	return "track2isrc:" + hex.EncodeToString(sum[:16])
}

func metaKey(isrc string) string {
	return "isrc2track:" + isrc
}
