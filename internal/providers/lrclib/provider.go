package lrclib

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

type Provider struct {
	BaseURL string `toml:"base_url,commented,omitempty" comment:"which instance to use"`
	providers.Common
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
	TrackName    string  `json:"trackName"`
	ArtistName   string  `json:"artistName"`
	AlbumName    string  `json:"albumName"`
	PlainLyrics  string  `json:"plainLyrics"`
	SyncedLyrics string  `json:"syncedLyrics"`
	ID           int64   `json:"id"`
	Duration     float64 `json:"duration"`
	Instrumental bool    `json:"instrumental"`
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
	res := toResult(r)
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
	default:
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	return nil
}

func toResult(r apiResult) *lyrics.Result {
	if r.SyncedLyrics == "" && r.PlainLyrics == "" {
		return nil
	}
	res := &lyrics.Result{}
	if r.SyncedLyrics != "" {
		res.Lines, res.SyncLevel = lyrics.ParseLRC(r.SyncedLyrics)
	} else {
		res.Lines = lyrics.ParsePlain(r.PlainLyrics)
		res.SyncLevel = lyrics.SyncNone
	}
	return res
}
