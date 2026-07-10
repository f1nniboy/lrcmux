package isrc

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"

	"github.com/agnivade/levenshtein"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/normalize"
)

const lookupBase = "https://api.deezer.com/search"

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
	t := lyrics.Track{
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
	return t
}

func (r *Resolver) lookup(ctx context.Context, in ResolveInput) (lyrics.Track, error) {
	q := fmt.Sprintf(`artist:"%s" track:"%s"`, in.Artist, in.Title)
	endpoint := lookupBase + "?q=" + url.QueryEscape(q) + "&limit=10"

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

	return toTrack(pickBest(dr.Data, in)), nil
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

func pickBest(tracks []deezerTrack, in ResolveInput) deezerTrack {
	if len(tracks) == 1 {
		return tracks[0]
	}

	wantTitle := normalize.Title(in.Title)
	wantArtist := normalize.String(in.Artist)

	type scored struct {
		track deezerTrack
		score float64
	}
	best := scored{score: -1}
	for _, t := range tracks {
		var s float64

		gotTitle := normalize.Title(t.Title)
		s += float64(max(0, 5-levenshtein.ComputeDistance(gotTitle, wantTitle)))

		gotArtist := normalize.String(t.Artist.Name)
		if normalize.ArtistMatch(t.Artist.Name, in.Artist) || normalize.ArtistMatch(in.Artist, t.Artist.Name) {
			s += 3
		} else if d := levenshtein.ComputeDistance(gotArtist, wantArtist); d <= 3 {
			s += float64(max(0, 2-d))
		}

		if in.Duration > 0 {
			diff := math.Abs(float64(t.Duration - in.Duration))
			s += max(0, 3.0-diff/2.0)
		}

		if in.Album != "" && normalize.String(t.Album.Title) == normalize.String(in.Album) {
			s += 2
		}

		if s > best.score {
			best = scored{track: t, score: s}
		}
	}
	return best.track
}
