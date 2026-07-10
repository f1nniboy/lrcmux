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
	"github.com/f1nniboy/lrcmux/internal/utils"
)

const lookupBase = "https://api.deezer.com/search"

type deezerTrack struct {
	ID             int64        `json:"id"`
	ISRC           string       `json:"isrc"`
	Title          string       `json:"title"`
	TitleShort     string       `json:"title_short"`
	Duration       int64        `json:"duration"`
	Preview        string       `json:"preview,omitempty"`
	ExplicitLyrics bool         `json:"explicit_lyrics"`
	ReleaseDate    string       `json:"release_date,omitempty"`
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

	wantTitle := utils.NormalizeTitle(in.Title)
	wantArtist := utils.Normalize(in.Artist)

	type scored struct {
		track deezerTrack
		score float64
	}
	best := scored{score: -1}
	for _, t := range tracks {
		var s float64

		gotTitle := utils.NormalizeTitle(t.Title)
		s += float64(max(0, 5-levenshtein.ComputeDistance(gotTitle, wantTitle)))

		gotArtist := utils.Normalize(t.Artist.Name)
		if utils.ArtistMatch(t.Artist.Name, in.Artist) || utils.ArtistMatch(in.Artist, t.Artist.Name) {
			s += 3
		} else if d := levenshtein.ComputeDistance(gotArtist, wantArtist); d <= 3 {
			s += float64(max(0, 2-d))
		}

		if in.Duration > 0 {
			diff := math.Abs(float64(t.Duration - in.Duration))
			s += max(0, 3.0-diff/2.0)
		}

		if in.Album != "" && utils.Normalize(t.Album.Title) == utils.Normalize(in.Album) {
			s += 2
		}

		if s > best.score {
			best = scored{track: t, score: s}
		}
	}
	return best.track
}
