package genius

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/normalize"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

const (
	searchURL = "https://genius.com/api/search/multi"
	// the embed page is way smaller than the normal song page
	embedURL  = "https://genius.com/songs/%d/embed.js"
	userAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:151.0) Gecko/20100101 Firefox/151.0"
)

type Provider struct {
	providers.Common
}

func (p *Provider) ID() string                 { return "genius" }
func (p *Provider) Name() string               { return "Genius" }
func (p *Provider) URL() string                { return "https://genius.com" }
func (p *Provider) Desc() string               { return "Best song coverage, but only plain text lyrics" }
func (p *Provider) MaxLevel() lyrics.SyncLevel { return lyrics.SyncNone }

func (p *Provider) Search(ctx context.Context, q lyrics.Query) (*lyrics.Result, error) {
	id, instrumental, err := p.findSong(ctx, q)
	if err != nil {
		return nil, err
	}
	if instrumental {
		return &lyrics.Result{Instrumental: true}, nil
	}

	lines, err := p.lyrics(ctx, id)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, lyrics.ErrNotFound
	}
	return &lyrics.Result{Lines: lines, SyncLevel: lyrics.SyncNone}, nil
}

type searchResponse struct {
	Response struct {
		Sections []struct {
			Type string `json:"type"`
			Hits []struct {
				Type   string `json:"type"`
				Result struct {
					Title        string `json:"title"`
					ArtistNames  string `json:"artist_names"`
					ID           int    `json:"id"`
					Instrumental bool   `json:"instrumental"`
				} `json:"result"`
			} `json:"hits"`
		} `json:"sections"`
	} `json:"response"`
}

func (p *Provider) findSong(ctx context.Context, q lyrics.Query) (id int, instrumental bool, err error) {
	endpoint := searchURL + "?per_page=5&q=" + url.QueryEscape(q.Track.Artist+" "+q.Track.Title)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, false, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Referer", "https://genius.com/search/embed")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	resp, err := p.HTTP.Do(req)
	if err != nil {
		return 0, false, fmt.Errorf("search: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusForbidden, http.StatusTooManyRequests:
		return 0, false, providers.ErrRateLimited
	default:
		return 0, false, fmt.Errorf("search status %d", resp.StatusCode)
	}

	var sr searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return 0, false, fmt.Errorf("search decode: %w", err)
	}

	for _, section := range sr.Response.Sections {
		for _, hit := range section.Hits {
			if hit.Type != "song" {
				continue
			}
			r := hit.Result
			if normalize.Match(q.Track.Title, q.Track.Artist, r.Title, r.ArtistNames) {
				return r.ID, r.Instrumental, nil
			}
		}
	}
	return 0, false, lyrics.ErrNotFound
}

func (p *Provider) lyrics(ctx context.Context, id int) ([]lyrics.Line, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(embedURL, id), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Sec-Fetch-Dest", "script")
	req.Header.Set("Sec-Fetch-Mode", "no-cors")
	req.Header.Set("Sec-Fetch-Site", "cross-site")

	resp, err := p.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return nil, lyrics.ErrNotFound
	case http.StatusForbidden, http.StatusTooManyRequests:
		return nil, providers.ErrRateLimited
	default:
		return nil, fmt.Errorf("embed status %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("embed: %w", err)
	}
	return parseEmbed(string(b))
}
