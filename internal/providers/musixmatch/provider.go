package musixmatch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"net/url"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

// mostly borrowed from https://github.com/OrfiDev/orpheusdl-musixmatch

const (
	desktopBaseURL   = "https://apic-desktop.musixmatch.com/ws/1.1/"
	desktopAppID     = "web-desktop-app-v1.0"
	desktopUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Musixmatch/0.19.4 Chrome/58.0.3029.110 Electron/1.7.6 Safari/537.36"
)

type tier struct {
	endpoint     string
	extra        url.Values
	bodyKey      string
	contentField string
	parse        func(string) []lyrics.Line
}

var tierByLevel = map[lyrics.SyncLevel]tier{
	lyrics.SyncWord: {endpoint: "track.richsync.get", bodyKey: "richsync", contentField: "richsync_body", parse: parseRichsync},
	lyrics.SyncLine: {endpoint: "track.subtitle.get", bodyKey: "subtitle", contentField: "subtitle_body", parse: parseSubtitles, extra: url.Values{"subtitle_format": {"mxm"}}},
	lyrics.SyncNone: {endpoint: "track.lyrics.get", bodyKey: "lyrics", contentField: "lyrics_body", parse: parseLyrics},
}

func init() {
	providers.Register("musixmatch", factory)
}

func factory(args providers.FactoryArgs) (providers.Impl, error) {
	var c Config
	if err := args.Decode(&c); err != nil {
		return nil, err
	}
	if c.PoolSize <= 0 {
		c.PoolSize = 1
	}
	return &Provider{
		client: &http.Client{},
		pool:   newTokenPool(c.PoolSize, args.Client, args.Cache, args.Log),
		log:    args.Log,
	}, nil
}

type Provider struct {
	client *http.Client
	pool   *tokenPool
	log    *slog.Logger
}

func (p *Provider) Name() string               { return "Musixmatch" }
func (p *Provider) Desc() string               { return "Extensive library and good word-level sync coverage" }
func (p *Provider) MaxLevel() lyrics.SyncLevel { return lyrics.SyncWord }

func (p *Provider) Search(ctx context.Context, q lyrics.Query) (*lyrics.Result, error) {
	meta, err := p.fetchTrackMeta(ctx, q.Track.ISRC)
	if err != nil {
		return nil, err
	}
	if meta == nil {
		return nil, lyrics.ErrNotFound
	}

	t := tierByLevel[meta.syncLevel]
	lines, err := p.fetchTier(ctx, t, q.Track.ISRC)
	if err != nil {
		return nil, err
	}
	if len(lines) > 0 {
		return &lyrics.Result{Lines: lines, SyncLevel: meta.syncLevel}, nil
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return nil, lyrics.ErrNotFound
}

type trackMeta struct {
	syncLevel lyrics.SyncLevel
}

func (p *Provider) fetchTrackMeta(ctx context.Context, isrc string) (*trackMeta, error) {
	body, err := p.getDesktop(ctx, "track.get", url.Values{"track_isrc": {isrc}})
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
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, nil
	}
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

func (p *Provider) fetchTier(ctx context.Context, t tier, isrc string) ([]lyrics.Line, error) {
	params := url.Values{"track_isrc": {isrc}}
	maps.Copy(params, t.extra)

	body, err := p.getDesktop(ctx, t.endpoint, params)
	if err != nil {
		if errors.Is(err, providers.ErrRateLimited) {
			return nil, err
		}
		if ctx.Err() == nil && !errors.Is(err, lyrics.ErrNotFound) {
			p.log.Debug("fetch failed", "tier", t.bodyKey, "isrc", isrc, "err", err)
			return nil, err
		}
		return nil, nil
	}

	var outer map[string]map[string]json.RawMessage
	if err := json.Unmarshal(body, &outer); err != nil {
		return nil, nil
	}
	var content string
	if err := json.Unmarshal(outer[t.bodyKey][t.contentField], &content); err != nil || content == "" {
		return nil, nil
	}
	return t.parse(content), nil
}

func (p *Provider) getDesktop(ctx context.Context, endpoint string, extra url.Values) (json.RawMessage, error) {
	for range len(p.pool.slots) {
		token, idx, err := p.pool.get(ctx)
		if err != nil {
			return nil, err
		}
		body, err := p.doDesktopRequest(ctx, endpoint, extra, token)
		if err == nil {
			return body, nil
		}
		if errors.Is(err, errRenew) {
			p.log.Debug("token expired, rotating", "slot", idx)
		} else if errors.Is(err, errCaptcha) {
			p.log.Debug("token rate limited, rotating", "slot", idx)
		} else {
			return nil, err
		}
		p.pool.retire(idx)
	}
	return nil, providers.ErrRateLimited
}

func (p *Provider) doDesktopRequest(ctx context.Context, endpoint string, extra url.Values, token string) (json.RawMessage, error) {
	params := url.Values{}
	params.Set("app_id", desktopAppID)
	params.Set("format", "json")
	params.Set("usertoken", token)
	maps.Copy(params, extra)

	u := desktopBaseURL + endpoint + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", desktopUserAgent)

	resp, err := p.client.Do(req)
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
				StatusCode int    `json:"status_code"`
				Hint       string `json:"hint"`
			} `json:"header"`
			Body json.RawMessage `json:"body"`
		} `json:"message"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	switch {
	case res.Message.Header.StatusCode == 200:
		return res.Message.Body, nil
	case res.Message.Header.StatusCode == 401 && res.Message.Header.Hint == "renew":
		return nil, errRenew
	case res.Message.Header.StatusCode == 401 && res.Message.Header.Hint == "captcha":
		return nil, errCaptcha
	case res.Message.Header.StatusCode == 404:
		return nil, lyrics.ErrNotFound
	}
	p.log.Debug("api error", "status", res.Message.Header.StatusCode, "hint", res.Message.Header.Hint)
	return nil, fmt.Errorf("api %d (%s)", res.Message.Header.StatusCode, res.Message.Header.Hint)
}
