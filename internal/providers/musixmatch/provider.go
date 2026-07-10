package musixmatch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

// mostly borrowed from https://github.com/OrfiDev/orpheusdl-musixmatch

const (
	baseURL   = "https://apic-desktop.musixmatch.com/ws/1.1/"
	appID     = "web-desktop-app-v1.0"
	userAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Musixmatch/0.19.4 Chrome/58.0.3029.110 Electron/1.7.6 Safari/537.36"
)

type tier struct {
	extra        url.Values
	parse        func(string) []lyrics.Line
	endpoint     string
	bodyKey      string
	contentField string
}

var tierByLevel = map[lyrics.SyncLevel]tier{
	lyrics.SyncWord: {endpoint: "track.richsync.get", bodyKey: "richsync", contentField: "richsync_body", parse: parseRichsync},
	lyrics.SyncLine: {endpoint: "track.subtitle.get", bodyKey: "subtitle", contentField: "subtitle_body", parse: parseSubtitles, extra: url.Values{"subtitle_format": {"mxm"}}},
	lyrics.SyncNone: {endpoint: "track.lyrics.get", bodyKey: "lyrics", contentField: "lyrics_body", parse: parseLyrics},
}

type Provider struct {
	pool *tokenPool
	providers.Common
	PoolSize int `toml:"pool_size,commented,omitempty" comment:"how many tokens to use in rotation"`
}

func (p *Provider) ID() string { return "musixmatch" }
func (p *Provider) Init() {
	if p.PoolSize <= 0 {
		p.PoolSize = 1
	}
	p.pool = newTokenPool(p.PoolSize, p.HTTP, p.Cache, p.Log)
}

func (p *Provider) Name() string               { return "Musixmatch" }
func (p *Provider) Desc() string               { return "Extensive library and good word-level sync coverage" }
func (p *Provider) MaxLevel() lyrics.SyncLevel { return lyrics.SyncWord }

func (p *Provider) Search(ctx context.Context, q lyrics.Query) (*lyrics.Result, error) {
	for {
		token, idx, err := p.pool.get(ctx)
		if err != nil {
			return nil, err
		}

		meta, err := p.fetchTrackMeta(ctx, q.Track.ISRC, token, idx)
		if errors.Is(err, providers.ErrRateLimited) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if meta == nil {
			return nil, lyrics.ErrNotFound
		}

		t := tierByLevel[meta.syncLevel]
		lines, err := p.fetchTier(ctx, t, q.Track.ISRC, token, idx)
		if errors.Is(err, providers.ErrRateLimited) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if len(lines) > 0 {
			return &lyrics.Result{Lines: lines, SyncLevel: meta.syncLevel}, nil
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, lyrics.ErrNotFound
	}
}

type trackMeta struct {
	syncLevel lyrics.SyncLevel
}

func (p *Provider) fetchTrackMeta(ctx context.Context, isrc, token string, idx int) (*trackMeta, error) {
	body, err := p.get(ctx, "track.get", url.Values{"track_isrc": {isrc}}, token, idx)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Track struct {
			HasRichsync  int `json:"has_richsync"`
			HasSubtitles int `json:"has_subtitles"`
			HasLyrics    int `json:"has_lyrics"`
		} `json:"track"`
	}
	json.Unmarshal(body, &resp)
	track := resp.Track
	switch {
	case track.HasRichsync == 1:
		return &trackMeta{syncLevel: lyrics.SyncWord}, nil
	case track.HasSubtitles == 1:
		return &trackMeta{syncLevel: lyrics.SyncLine}, nil
	case track.HasLyrics == 1:
		return &trackMeta{syncLevel: lyrics.SyncNone}, nil
	}
	return nil, nil
}

func (p *Provider) fetchTier(ctx context.Context, t tier, isrc, token string, idx int) ([]lyrics.Line, error) {
	params := url.Values{"track_isrc": {isrc}}
	maps.Copy(params, t.extra)

	body, err := p.get(ctx, t.endpoint, params, token, idx)
	if err != nil {
		if errors.Is(err, providers.ErrRateLimited) {
			return nil, err
		}
		if ctx.Err() == nil && !errors.Is(err, lyrics.ErrNotFound) {
			p.Log.Debug("fetch failed", "tier", t.bodyKey, "isrc", isrc, "err", err)
			return nil, err
		}
		return nil, nil
	}

	var outer map[string]map[string]json.RawMessage
	json.Unmarshal(body, &outer)
	var content string
	json.Unmarshal(outer[t.bodyKey][t.contentField], &content)
	if content == "" {
		return nil, nil
	}
	return t.parse(content), nil
}

func (p *Provider) get(ctx context.Context, endpoint string, extra url.Values, token string, idx int) (json.RawMessage, error) {
	p.Log.Debug("using token", "slot", idx, "token", token[:8])
	body, err := p.doRequest(ctx, endpoint, extra, token)
	if err == nil {
		return body, nil
	}
	if errors.Is(err, errTokenUnusable) {
		p.Log.Debug("token unusable, retiring", "slot", idx)
		p.pool.retire(idx)
		return nil, providers.ErrRateLimited
	}
	return nil, err
}

func (p *Provider) doRequest(ctx context.Context, endpoint string, extra url.Values, token string) (json.RawMessage, error) {
	params := url.Values{}
	params.Set("app_id", appID)
	params.Set("format", "json")
	params.Set("usertoken", token)
	maps.Copy(params, extra)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := p.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	var res struct {
		Message struct {
			Header struct {
				Hint       string `json:"hint"`
				StatusCode int    `json:"status_code"`
			} `json:"header"`
			Body json.RawMessage `json:"body"`
		} `json:"message"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	switch res.Message.Header.StatusCode {
	case 200:
		return res.Message.Body, nil
	case 401:
		return nil, errTokenUnusable
	case 404:
		return nil, lyrics.ErrNotFound
	}
	p.Log.Debug("api error", "status", res.Message.Header.StatusCode, "hint", res.Message.Header.Hint)
	return nil, fmt.Errorf("api %d (%s)", res.Message.Header.StatusCode, res.Message.Header.Hint)
}
