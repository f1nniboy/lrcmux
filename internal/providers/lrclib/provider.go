package lrclib

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/f1nniboy/lrcmux/internal/format"
	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

//nolint:govet // fieldalignment
type Provider struct {
	providers.Common
	BaseURL string `toml:"base_url,commented,omitempty" comment:"which instance to use"`
}

func (p *Provider) ID() string                 { return "lrclib" }
func (p *Provider) Name() string               { return "LRCLIB" }
func (p *Provider) URL() string                { return p.BaseURL }
func (p *Provider) Desc() string               { return "Community-sourced lyrics database" }
func (p *Provider) MaxLevel() lyrics.SyncLevel { return lyrics.SyncLine }

func (p *Provider) Init() {
	if p.BaseURL == "" {
		p.BaseURL = "https://lrclib.net"
	}
	p.BaseURL = strings.TrimRight(p.BaseURL, "/")
}

type apiResult struct {
	LyricsFile string `json:"lyricsfile"`
}

func (p *Provider) Search(ctx context.Context, q lyrics.Query) (*lyrics.Result, error) {
	params := url.Values{}
	params.Set("artist_name", q.Track.Artist)
	params.Set("track_name", q.Track.Title)
	params.Set("album_name", q.Track.Album)
	params.Set("duration", strconv.FormatInt(q.Track.Duration, 10))
	endpoint := p.BaseURL + "/api/get?" + params.Encode()

	var r apiResult
	if err := p.do(ctx, endpoint, &r); err != nil {
		return nil, err
	}
	res, err := toResult(r)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, lyrics.ErrNotFound
	}
	return res, nil
}

func (p *Provider) do(ctx context.Context, endpoint string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	resp, err := p.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return lyrics.ErrNotFound
	case http.StatusTooManyRequests:
		return providers.RateLimitedFrom(resp)
	default:
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	return nil
}

func toResult(r apiResult) (*lyrics.Result, error) {
	if r.LyricsFile == "" {
		return nil, nil
	}
	res, err := format.ParseLyricsFile([]byte(r.LyricsFile))
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	if !res.Instrumental && len(res.Lines) == 0 {
		return nil, nil
	}
	return res, nil
}
