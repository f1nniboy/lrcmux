package genius

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
	"github.com/f1nniboy/lrcmux/internal/utils"
)

const (
	searchURL = "https://genius.com/api/search/multi"
	userAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:151.0) Gecko/20100101 Firefox/151.0"
)

var reSection = regexp.MustCompile(`\[.*?\]`)

type Provider struct {
	providers.Common
}

func (p *Provider) ID() string                 { return "genius" }
func (p *Provider) Name() string               { return "Genius" }
func (p *Provider) Desc() string               { return "Best song coverage, but only plain text lyrics" }
func (p *Provider) MaxLevel() lyrics.SyncLevel { return lyrics.SyncNone }

func (p *Provider) Search(ctx context.Context, q lyrics.Query) (*lyrics.Result, error) {
	pageURL, err := p.search(ctx, q)
	if err != nil {
		return nil, err
	}

	lines, err := p.scrape(ctx, pageURL)
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
					Title       string `json:"title"`
					URL         string `json:"url"`
					ArtistNames string `json:"artist_names"`
				} `json:"result"`
			} `json:"hits"`
		} `json:"sections"`
	} `json:"response"`
}

func (p *Provider) search(ctx context.Context, q lyrics.Query) (string, error) {
	endpoint := searchURL + "?per_page=5&q=" + url.QueryEscape(q.Track.Artist+" "+q.Track.Title)
	p.Log.Debug("search", "artist", q.Track.Artist, "title", q.Track.Title)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
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
		return "", fmt.Errorf("search: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusForbidden, http.StatusTooManyRequests:
		return "", providers.ErrRateLimited
	default:
		return "", fmt.Errorf("search status %d", resp.StatusCode)
	}

	var sr searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return "", fmt.Errorf("search decode: %w", err)
	}

	wantTitle := utils.NormalizeTitle(q.Track.Title)
	wantArtist := utils.Normalize(q.Track.Artist)

	for _, section := range sr.Response.Sections {
		for _, hit := range section.Hits {
			if hit.Type != "song" {
				continue
			}
			r := hit.Result
			gotTitle := utils.NormalizeTitle(r.Title)
			titleOK := gotTitle == wantTitle
			artistOK := utils.ArtistMatch(r.ArtistNames, wantArtist)
			p.Log.Debug("candidate", "title", r.Title, "artist", r.ArtistNames, "title_ok", titleOK, "artist_ok", artistOK)
			if !titleOK || !artistOK {
				continue
			}
			p.Log.Debug("matched", "url", r.URL)
			return r.URL, nil
		}
	}
	p.Log.Debug("no match", "want_title", wantTitle, "want_artist", wantArtist)
	return "", lyrics.ErrNotFound
}

func (p *Provider) scrape(ctx context.Context, pageURL string) ([]lyrics.Line, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")

	resp, err := p.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("scrape: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return nil, lyrics.ErrNotFound
	case http.StatusForbidden, http.StatusTooManyRequests:
		return nil, providers.ErrRateLimited
	default:
		return nil, fmt.Errorf("scrape status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	var raw strings.Builder
	doc.Find(`div[data-lyrics-container="true"]`).Each(func(_ int, s *goquery.Selection) {
		s.Find(`div[data-exclude-from-selection="true"]`).Remove()
		s.Find("br").ReplaceWithHtml("\n")
		raw.WriteString(s.Text())
		raw.WriteByte('\n')
	})

	if raw.Len() == 0 {
		return nil, nil
	}

	text := raw.String()
	if i := strings.Index(text, "["); i >= 0 {
		text = text[i:]
	}
	text = reSection.ReplaceAllString(text, "")

	var lines []lyrics.Line
	for l := range strings.SplitSeq(text, "\n") {
		t := strings.TrimSpace(l)

		// strip lines like "(Intro: ...)", "(Chorus: ...)"
		if strings.HasPrefix(t, "(") && strings.HasSuffix(t, ")") && strings.Contains(t, ":") {
			continue
		}

		lines = append(lines, lyrics.Line{Text: t})
	}
	for len(lines) > 0 && lines[0].Text == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && lines[len(lines)-1].Text == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return nil, nil
	}
	return lines, nil
}
