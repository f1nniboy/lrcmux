package isrc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"sync"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
)

const baseURL = "https://api.deezer.com"

type deezerTrack struct {
	ISRC           string       `json:"isrc"`
	Title          string       `json:"title"`
	TitleShort     string       `json:"title_short"`
	Preview        string       `json:"preview,omitempty"`
	ReleaseDate    string       `json:"release_date,omitempty"`
	Album          deezerAlbum  `json:"album"`
	Artist         deezerArtist `json:"artist"`
	ID             int64        `json:"id"`
	Duration       int64        `json:"duration"`
	ExplicitLyrics bool         `json:"explicit_lyrics"`
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
	}
}

type deezerArtist struct {
	Name          string `json:"name"`
	PictureSmall  string `json:"picture_small,omitempty"`
	PictureMedium string `json:"picture_medium,omitempty"`
	ID            int64  `json:"id"`
}

type deezerAlbum struct {
	Title       string `json:"title"`
	CoverSmall  string `json:"cover_small,omitempty"`
	CoverMedium string `json:"cover_medium,omitempty"`
	CoverBig    string `json:"cover_big,omitempty"`
	ID          int64  `json:"id"`
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

// deezer's code for a lookup that matched nothing
const codeNoData = 800

func (e *deezerError) err() error {
	if e.Code == codeNoData {
		return lyrics.ErrNotFound
	}
	return fmt.Errorf("deezer %s %d: %s", e.Type, e.Code, e.Message)
}

// complete is false once a request failed, since its results could have held a
// better match, so the track is served but not cached
func (r *Resolver) lookup(ctx context.Context, in ResolveInput) (track lyrics.Track, complete bool, err error) {
	reads := readings(in)

	// deezer's artist: filter returns nothing even for exact matches, so pair a
	// title-only filter (finds originals buried under remixes) with free text
	// (finds generic titles that drown in a title-only search)
	var queries []string
	for _, rd := range reads {
		for _, q := range []string{fmt.Sprintf(`track:"%s"`, rd.Title), rd.Artist + " " + rd.Title} {
			if !slices.Contains(queries, q) {
				queries = append(queries, q)
			}
		}
	}

	results := make([][]deezerTrack, len(queries))
	errs := make([]error, len(queries))
	var wg sync.WaitGroup
	for i, q := range queries {
		wg.Go(func() { results[i], errs[i] = r.search(ctx, q) })
	}
	wg.Wait()
	tracks := mergeTracks(results...)

	var best candidate
	found := false
	consider := func(reading int, c candidate, ok bool) {
		c.reading = reading
		if ok && (!found || c.beats(best)) {
			best, found = c, true
		}
	}
	for i, rd := range reads {
		c, ok := pickBest(tracks, rd, 0)
		consider(i, c, ok)
	}
	if !found {
		// an artist named by an alias or in another script ("米津玄師" for
		// "Kenshi Yonezu") fails the comparison, so ask deezer who it is
		ids := map[string]int64{}
		for i, rd := range reads {
			if !titleMatches(tracks, rd.Title) {
				continue
			}
			id, asked := ids[rd.Artist]
			if !asked {
				var err error
				if id, err = r.artistID(ctx, rd.Artist); err != nil {
					errs = append(errs, err)
				}
				ids[rd.Artist] = id
			}
			if id != 0 {
				c, ok := pickBest(tracks, rd, id)
				consider(i, c, ok)
			}
		}
	}

	var failed error
	for _, err := range errs {
		if err != nil && !errors.Is(err, lyrics.ErrNotFound) {
			failed = err
			break
		}
	}
	if !found {
		if failed != nil {
			return lyrics.Track{}, false, failed
		}
		return lyrics.Track{}, true, lyrics.ErrNotFound
	}
	return best.track.toTrack(), failed == nil, nil
}

func (r *Resolver) artistID(ctx context.Context, name string) (int64, error) {
	var res struct {
		deezerEnvelope
		Data []deezerArtist `json:"data"`
	}
	if err := r.get(ctx, "/search/artist?q="+url.QueryEscape(name)+"&limit=1", &res); err != nil {
		return 0, err
	}
	if len(res.Data) == 0 {
		return 0, nil
	}
	return res.Data[0].ID, nil
}

func (r *Resolver) search(ctx context.Context, q string) ([]deezerTrack, error) {
	var res struct {
		deezerEnvelope
		Data []deezerTrack `json:"data"`
	}
	if err := r.get(ctx, "/search?q="+url.QueryEscape(q)+"&limit=10", &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

func mergeTracks(lists ...[]deezerTrack) []deezerTrack {
	seen := make(map[int64]bool)
	var out []deezerTrack
	for _, list := range lists {
		for _, t := range list {
			if seen[t.ID] {
				continue
			}
			seen[t.ID] = true
			out = append(out, t)
		}
	}
	return out
}

func (r *Resolver) lookupMeta(ctx context.Context, isrc string) (lyrics.Track, error) {
	var res struct {
		deezerEnvelope
		deezerTrack
	}
	if err := r.get(ctx, "/track/isrc:"+isrc, &res); err != nil {
		return lyrics.Track{}, err
	}
	return res.toTrack(), nil
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
