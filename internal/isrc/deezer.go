package isrc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"slices"
	"sync"

	"github.com/agnivade/levenshtein"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/normalize"
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

type deezerSearchResponse struct {
	Data []deezerTrack `json:"data"`
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

func toTrack(raw deezerTrack) lyrics.Track {
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

func (r *Resolver) lookup(ctx context.Context, in ResolveInput) (lyrics.Track, error) {
	queries := []string{
		fmt.Sprintf(`artist:"%s" track:"%s"`, in.Artist, in.Title),
		in.Artist + " " + in.Title,
	}

	results := make([][]deezerTrack, len(queries))
	errs := make([]error, len(queries))

	var wg sync.WaitGroup
	for i, q := range queries {
		wg.Go(func() { results[i], errs[i] = r.search(ctx, q) })
	}
	wg.Wait()

	if !slices.Contains(errs, nil) {
		return lyrics.Track{}, errors.Join(errs...)
	}

	merged := mergeTracks(results...)
	if len(merged) == 0 {
		return lyrics.Track{}, lyrics.ErrNotFound
	}

	return toTrack(pickBest(merged, in)), nil
}

func (r *Resolver) search(ctx context.Context, q string) ([]deezerTrack, error) {
	endpoint := baseURL + "/search?q=" + url.QueryEscape(q) + "&limit=10"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("deezer request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("deezer status %d", resp.StatusCode)
	}

	var dr deezerSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&dr); err != nil {
		return nil, fmt.Errorf("deezer decode: %w", err)
	}
	return dr.Data, nil
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

func distScore(a, b string, n int) float64 {
	return float64(max(0, n-levenshtein.ComputeDistance(a, b)))
}

// shared score for the title/artist/duration signals, so no
// single one alone can outweigh an exact match in another
const matchScore = 5

const durationWindowSecs = 30

func pickBest(tracks []deezerTrack, in ResolveInput) deezerTrack {
	if len(tracks) == 1 {
		return tracks[0]
	}

	wantTitle := normalize.Title(in.Title)
	wantArtist := normalize.String(in.Artist)

	var best deezerTrack
	bestScore := -1.0

	for _, t := range tracks {
		var s float64

		// title
		s += distScore(normalize.Title(t.Title), wantTitle, matchScore)

		// artist
		if normalize.ArtistMatch(t.Artist.Name, in.Artist) || normalize.ArtistMatch(in.Artist, t.Artist.Name) {
			s += matchScore
		} else {
			s += distScore(normalize.String(t.Artist.Name), wantArtist, matchScore)
		}

		// duration
		if in.Duration > 0 {
			delta := math.Abs(float64(t.Duration - in.Duration))
			if delta < durationWindowSecs {
				s += matchScore * (1 - delta/durationWindowSecs)
			}
		}

		// album
		if in.Album != "" {
			s += distScore(normalize.String(t.Album.Title), normalize.String(in.Album), 2)
		}

		if s > bestScore {
			best, bestScore = t, s
		}
	}
	return best
}
