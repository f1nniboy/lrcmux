package lrclib

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/f1nniboy/lrcmux/internal/format"
	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

type apiResult struct {
	Lyricsfile string `json:"lyricsfile"`
}

func (r apiResult) toResult() (*lyrics.Result, error) {
	if r.Lyricsfile == "" {
		return nil, nil
	}
	res, err := format.ParseLyricsfile([]byte(r.Lyricsfile))
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	return res, nil
}

//nolint:govet // fieldalignment
type Provider struct {
	providers.Common
	BaseURL string `toml:"base_url,commented,omitempty" comment:"which instance to use"`
	APIKey  string `toml:"api_key,commented,omitempty" validate:"required_with=AppID" comment:"secret key of an application registered at https://lrclib.net/developers\nyou only need this if you exceed the 90req/min limit"`
	AppID   string `toml:"app_id,commented,omitempty" validate:"required_with=APIKey" comment:"ID of that application"`
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
	res, err := r.toResult()
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
	req.Header.Set("User-Agent", providers.AppUserAgent)
	if p.APIKey != "" {
		req.Header.Set("Lrclib-Api-Key", p.APIKey)
	}
	if p.AppID != "" {
		req.Header.Set("Lrclib-App-Id", p.AppID)
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
	case http.StatusTooManyRequests, http.StatusServiceUnavailable:
		return providers.RateLimitedFrom(resp)
	case http.StatusUnauthorized:
		// lrclib's json names the failed check, a 401 from its proxy has no body
		var body struct {
			Name string `json:"name"`
		}
		if json.NewDecoder(resp.Body).Decode(&body) != nil || body.Name == "" {
			return errors.New("unauthorized")
		}
		return fmt.Errorf("unauthorized: %s", body.Name)
	default:
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	return nil
}
