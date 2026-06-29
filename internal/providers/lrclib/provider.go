package lrclib

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

const defaultBaseURL = "https://lrclib.net"

func init() {
	providers.Register("lrclib", factory)
}

func factory(args providers.FactoryArgs) (providers.Impl, error) {
	var c Config
	if err := args.Decode(&c); err != nil {
		return nil, err
	}
	base := c.BaseURL
	if base == "" {
		base = defaultBaseURL
	}
	return &Provider{http: args.Client, log: args.Log, baseURL: strings.TrimRight(base, "/")}, nil
}

type Provider struct {
	http    *http.Client
	log     *slog.Logger
	baseURL string
}

func (p *Provider) Name() string               { return "LRCLIB" }
func (p *Provider) Desc() string               { return "Community-sourced lyrics database" }
func (p *Provider) MaxLevel() lyrics.SyncLevel { return lyrics.SyncLine }

type apiResult struct {
	ID           int64   `json:"id"`
	TrackName    string  `json:"trackName"`
	ArtistName   string  `json:"artistName"`
	AlbumName    string  `json:"albumName"`
	Duration     float64 `json:"duration"`
	Instrumental bool    `json:"instrumental"`
	PlainLyrics  string  `json:"plainLyrics"`
	SyncedLyrics string  `json:"syncedLyrics"`
}

func (p *Provider) Search(ctx context.Context, q lyrics.Query) (*lyrics.Result, error) {
	params := url.Values{}
	params.Set("artist_name", q.Artist)
	params.Set("track_name", q.Title)
	if q.Album != "" {
		params.Set("album_name", q.Album)
	}
	if q.Duration > 0 {
		params.Set("duration", strconv.FormatInt(q.Duration, 10))
	}
	endpoint := p.baseURL + "/api/get?" + params.Encode()

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
	resp, err := p.http.Do(req)
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
		res.Lines = parseSynced(r.SyncedLyrics)
		res.SyncLevel = lyrics.SyncLine
	} else {
		res.Lines = plainToLines(r.PlainLyrics)
		res.SyncLevel = lyrics.SyncNone
	}
	return res
}

func plainToLines(s string) []lyrics.Line {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return nil
	}
	parts := strings.Split(s, "\n")
	out := make([]lyrics.Line, 0, len(parts))
	for _, p := range parts {
		out = append(out, lyrics.Line{Text: p})
	}
	return out
}
