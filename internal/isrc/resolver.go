package isrc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/agnivade/levenshtein"
	"golang.org/x/sync/singleflight"

	"github.com/f1nniboy/lrcmux/internal/cache"
	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/utils"
)

const lookupBase = "https://api.deezer.com/search"

type ResolveInput struct {
	Artist   string
	Title    string
	Album    string
	Duration int64
	ISRC     string
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

type deezerTrack struct {
	ID             int64        `json:"id"`
	ISRC           string       `json:"isrc"`
	Title          string       `json:"title"`
	TitleShort     string       `json:"title_short"`
	Duration       int64        `json:"duration"`
	Preview        string       `json:"preview,omitempty"`
	ExplicitLyrics bool         `json:"explicit_lyrics"`
	Artist         deezerArtist `json:"artist"`
	Album          deezerAlbum  `json:"album"`
}

type deezerSearchResponse struct {
	Data []deezerTrack `json:"data"`
}

type deezerArtist struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	PictureSmall  string `json:"picture_small,omitempty"`
	PictureMedium string `json:"picture_medium,omitempty"`
}

type deezerAlbum struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	CoverSmall  string `json:"cover_small,omitempty"`
	CoverMedium string `json:"cover_medium,omitempty"`
	CoverBig    string `json:"cover_big,omitempty"`
}

func toTrack(dt deezerTrack) lyrics.Track {
	return lyrics.Track{
		ISRC:     dt.ISRC,
		Title:    dt.Title,
		Duration: dt.Duration,
		Artist:   dt.Artist.Name,
		Album:    dt.Album.Title,
		Cover: lyrics.TrackCover{
			Small:  dt.Album.CoverSmall,
			Medium: dt.Album.CoverMedium,
			Big:    dt.Album.CoverBig,
		},
	}
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
	case cache.Found:
		r.log.Debug("isrc meta cache hit", "isrc", isrc)
		return track, nil
	case cache.KnownMiss:
		r.log.Debug("isrc meta negative cache hit", "isrc", isrc)
		return lyrics.Track{}, lyrics.ErrNotFound
	}

	lookupCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	track, err := r.lookupMeta(lookupCtx, isrc)
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
	switch _, status, _ := cache.Get[string](ctx, r.cache, key); status {
	case cache.KnownMiss:
		r.log.Debug("isrc negative cache hit", "artist", in.Artist, "title", in.Title)
		return lyrics.Track{}, lyrics.ErrNotFound
	}

	sfKey := utils.Normalize(in.Artist) + ":" + utils.Normalize(in.Title)
	v, err, _ := r.group.Do(sfKey, func() (any, error) {
		lookupCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		track, err := r.lookup(lookupCtx, in)
		if err != nil && !errors.Is(err, lyrics.ErrNotFound) {
			r.log.Warn("isrc lookup failed", "artist", in.Artist, "title", in.Title, "err", err)
			return lyrics.Track{}, err
		}

		if errors.Is(err, lyrics.ErrNotFound) {
			if primary := utils.PrimaryArtist(in.Artist); primary != "" && utils.Normalize(primary) != utils.Normalize(in.Artist) {
				primaryIn := in
				primaryIn.Artist = primary
				track, err = r.lookup(lookupCtx, primaryIn)
			}
		}

		if err != nil {
			r.log.Debug("isrc not found on deezer", "artist", in.Artist, "title", in.Title)
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

func (r *Resolver) lookup(ctx context.Context, in ResolveInput) (lyrics.Track, error) {
	q := fmt.Sprintf(`artist:"%s" track:"%s"`, in.Artist, in.Title)
	endpoint := lookupBase + "?q=" + url.QueryEscape(q) + "&limit=5"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return lyrics.Track{}, err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return lyrics.Track{}, fmt.Errorf("deezer request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return lyrics.Track{}, fmt.Errorf("deezer status %d", resp.StatusCode)
	}

	var dr deezerSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&dr); err != nil {
		return lyrics.Track{}, fmt.Errorf("deezer decode: %w", err)
	}

	if len(dr.Data) == 0 {
		return lyrics.Track{}, lyrics.ErrNotFound
	}

	best := pickBest(dr.Data, in)
	return toTrack(best), nil
}

func pickBest(tracks []deezerTrack, in ResolveInput) deezerTrack {
	if len(tracks) == 1 {
		return tracks[0]
	}

	wantTitle := utils.NormalizeTitle(in.Title)
	wantArtist := utils.Normalize(in.Artist)

	type scored struct {
		track deezerTrack
		score int
	}
	var best scored
	for _, t := range tracks {
		s := 0

		gotTitle := utils.NormalizeTitle(t.Title)
		s += max(0, 5-levenshtein.ComputeDistance(gotTitle, wantTitle))

		if utils.Normalize(t.Artist.Name) == wantArtist {
			s += 3
		}

		if in.Duration > 0 {
			diff := math.Abs(float64(t.Duration - in.Duration))
			if diff <= 1 {
				s += 3
			} else if diff <= 3 {
				s += 1
			}
		}
		if in.Album != "" {
			if utils.Normalize(t.Album.Title) == utils.Normalize(in.Album) {
				s += 2
			}
		}
		if s > best.score || best.score == 0 {
			best = scored{track: t, score: s}
		}
	}
	return best.track
}

func (r *Resolver) lookupMeta(ctx context.Context, isrc string) (lyrics.Track, error) {
	endpoint := "https://api.deezer.com/track/isrc:" + isrc

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return lyrics.Track{}, err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return lyrics.Track{}, fmt.Errorf("deezer request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return lyrics.Track{}, lyrics.ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return lyrics.Track{}, fmt.Errorf("deezer status %d", resp.StatusCode)
	}

	var dt deezerTrack
	if err := json.NewDecoder(resp.Body).Decode(&dt); err != nil {
		return lyrics.Track{}, fmt.Errorf("deezer decode: %w", err)
	}

	return toTrack(dt), nil
}

func trackKey(artist, title string) string {
	s := utils.Normalize(artist) + ":" + utils.Normalize(title)
	sum := sha256.Sum256([]byte(s))
	return "track2isrc:" + hex.EncodeToString(sum[:16])
}

func metaKey(isrc string) string {
	return "isrc2track:" + isrc
}
