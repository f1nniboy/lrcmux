package kugou

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

const (
	searchURL   = "http://krcs.kugou.com/search"
	downloadURL = "https://lyrics.kugou.com/download"
)

type Provider struct {
	providers.Common
}

func (p *Provider) ID() string                 { return "kugou" }
func (p *Provider) Name() string               { return "Kugou" }
func (p *Provider) Desc() string               { return "Word-level sync for most songs (AI?), censors profanity" }
func (p *Provider) MaxLevel() lyrics.SyncLevel { return lyrics.SyncWord }

type searchCandidate struct {
	ID         string `json:"id"`
	AccessKey  string `json:"accesskey"`
	SongName   string `json:"song"`
	SingerName string `json:"singer"`
	Duration   int64  `json:"duration"`
	Score      int64  `json:"score"`
	KRCType    int    `json:"krctype"`
}

var reBrackets = regexp.MustCompile(`[\(\[（【][^\)\]）】]*[\)\]）】]`)

func (p *Provider) Search(ctx context.Context, q lyrics.Query) (*lyrics.Result, error) {
	type attempt struct {
		artist, title string
	}

	stripped := func(s string) string {
		return strings.TrimSpace(reBrackets.ReplaceAllString(s, ""))
	}
	attempts := []attempt{
		{q.Track.Artist, q.Track.Title},
		{stripped(q.Track.Artist), stripped(q.Track.Title)},
	}

	for _, a := range attempts {
		if a.artist == "" || a.title == "" {
			continue
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		cand, err := p.findCandidate(ctx, a.artist, a.title, q.Track.Duration*1000)
		if err != nil {
			return nil, err
		}
		if cand == nil {
			continue
		}
		return p.download(ctx, cand)
	}
	return nil, lyrics.ErrNotFound
}

func (p *Provider) findCandidate(ctx context.Context, artist, title string, durationMs int64) (*searchCandidate, error) {
	params := url.Values{}
	params.Set("ver", "1")
	params.Set("man", "yes")
	params.Set("client", "mobi")
	params.Set("keyword", artist+" - "+title)
	params.Set("hash", "")
	params.Set("album_audio_id", "")
	if durationMs > 0 {
		params.Set("duration", strconv.FormatInt(durationMs, 10))
	} else {
		params.Set("duration", "")
	}

	var sr struct {
		Candidates []searchCandidate `json:"candidates"`
		Status     int               `json:"status"`
	}
	if err := p.do(ctx, searchURL+"?"+params.Encode(), &sr); err != nil {
		return nil, err
	}
	if len(sr.Candidates) == 0 {
		return nil, nil
	}
	return bestCandidate(sr.Candidates, durationMs), nil
}

func bestCandidate(candidates []searchCandidate, durationMs int64) *searchCandidate {
	best := &candidates[0]
	for i := range candidates[1:] {
		c := &candidates[i+1]
		if durationMs > 0 {
			bestDiff := abs64(best.Duration - durationMs)
			cDiff := abs64(c.Duration - durationMs)
			if cDiff < bestDiff || (cDiff == bestDiff && c.Score > best.Score) {
				best = c
			}
		} else if c.Score > best.Score {
			best = c
		}
	}
	return best
}

func abs64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

func (p *Provider) download(ctx context.Context, cand *searchCandidate) (*lyrics.Result, error) {
	params := url.Values{}
	params.Set("ver", "1")
	params.Set("client", "pc")
	params.Set("id", cand.ID)
	params.Set("accesskey", cand.AccessKey)
	params.Set("fmt", "krc")
	params.Set("charset", "utf8")

	var resp struct {
		Content string `json:"content"`
		Status  int    `json:"status"`
	}
	if err := p.do(ctx, downloadURL+"?"+params.Encode(), &resp); err != nil {
		return nil, err
	}
	if resp.Status != 200 || resp.Content == "" {
		return nil, lyrics.ErrNotFound
	}

	krcText, err := decodeKRC(resp.Content)
	if err != nil {
		return nil, err
	}

	lines := parseKRC(krcText)
	if len(lines) == 0 {
		return nil, lyrics.ErrNotFound
	}

	// horrible i guess, but no one wants censored lyrics
	for _, l := range lines {
		if strings.Contains(l.Text, "**") {
			return nil, lyrics.ErrNotFound
		}
	}

	return &lyrics.Result{
		Lines:     lines,
		SyncLevel: lyrics.SyncWord,
	}, nil
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
