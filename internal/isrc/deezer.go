package isrc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

const (
	baseURL     = "https://api.deezer.com"
	searchLimit = 50
)

type deezerTrack struct {
	ISRC   string       `json:"isrc"`
	Title  string       `json:"title"`
	Album  deezerAlbum  `json:"album"`
	Artist deezerArtist `json:"artist"`
	// only /track has them, a search result only names the first main artist
	Contributors []deezerArtist `json:"contributors"`
	ID           int64          `json:"id"`
	Duration     int64          `json:"duration"`
}

func (raw deezerTrack) toTrack() lyrics.Track {
	return lyrics.Track{
		ISRC:     raw.ISRC,
		Title:    raw.Title,
		Duration: raw.Duration,
		Artist:   raw.Artist.Name,
		Album:    raw.Album.Title,
		Cover: lyrics.TrackCover{
			Small:  raw.Album.CoverSmall,
			Medium: raw.Album.CoverMedium,
			Big:    raw.Album.CoverBig,
		},
		DeezerID: raw.ID,
	}
}

type deezerArtist struct {
	Name string `json:"name"`
	ID   int64  `json:"id"`
}

type deezerAlbum struct {
	Title       string `json:"title"`
	CoverSmall  string `json:"cover_small"`
	CoverMedium string `json:"cover_medium"`
	CoverBig    string `json:"cover_big"`
}

// deezer answers errors with http 200 and an error object
type deezerEnvelope struct {
	Error *deezerError `json:"error"`
}

func (e deezerEnvelope) apiErr() *deezerError { return e.Error }

type deezerError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

const (
	codeQuota  = 4
	codeNoData = 800
)

func (e *deezerError) err() error {
	switch e.Code {
	case codeNoData:
		return lyrics.ErrNotFound
	case codeQuota:
		return fmt.Errorf("deezer: %w", providers.ErrRateLimited)
	default:
		return fmt.Errorf("deezer %s %d: %s", e.Type, e.Code, e.Message)
	}
}

func (r *Resolver) lookup(ctx context.Context, in ResolveInput) (lyrics.Track, error) {
	tracks, err := r.search(ctx, in.Artist+" "+in.Title)
	if err != nil {
		return lyrics.Track{}, err
	}
	reads := readings(in)
	if c, ok := best(tracks, reads, byName); ok {
		return c.track.toTrack(), nil
	}
	c, ok := best(tracks, reads, anyArtist)
	if !ok {
		return lyrics.Track{}, lyrics.ErrNotFound
	}

	full, err := r.track(ctx, strconv.FormatInt(c.track.ID, 10))
	if err != nil && !errors.Is(err, lyrics.ErrNotFound) {
		return lyrics.Track{}, err
	}
	if err == nil {
		for i := range tracks {
			if tracks[i].ID == full.ID {
				tracks[i] = full
			}
		}
		if c, ok := best(tracks, reads, byName); ok {
			return c.track.toTrack(), nil
		}
	}

	// aliases and names in other scripts fail the name check
	id, err := r.artistID(ctx, in.Artist)
	if err != nil {
		return lyrics.Track{}, err
	}
	if c, ok := best(tracks, reads, byID(id)); id != 0 && ok {
		return c.track.toTrack(), nil
	}
	return lyrics.Track{}, lyrics.ErrNotFound
}

func (r *Resolver) artistID(ctx context.Context, name string) (int64, error) {
	var res struct {
		deezerEnvelope
		Data []deezerArtist `json:"data"`
	}
	err := r.get(ctx, "/search/artist?q="+url.QueryEscape(name)+"&limit=1", &res)
	if errors.Is(err, lyrics.ErrNotFound) || err == nil && len(res.Data) == 0 {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return res.Data[0].ID, nil
}

func (r *Resolver) track(ctx context.Context, ref string) (deezerTrack, error) {
	var res struct {
		deezerEnvelope
		deezerTrack
	}
	if err := r.get(ctx, "/track/"+ref, &res); err != nil {
		return deezerTrack{}, err
	}
	return res.deezerTrack, nil
}

func (r *Resolver) search(ctx context.Context, q string) ([]deezerTrack, error) {
	var res struct {
		deezerEnvelope
		Data []deezerTrack `json:"data"`
	}
	if err := r.get(ctx, "/search?q="+url.QueryEscape(q)+"&limit="+strconv.Itoa(searchLimit), &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

func (r *Resolver) get(ctx context.Context, path string, out interface{ apiErr() *deezerError }) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+path, nil)
	if err != nil {
		return err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("deezer request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("deezer status %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("deezer decode: %w", err)
	}
	if e := out.apiErr(); e != nil {
		return e.err()
	}
	return nil
}
