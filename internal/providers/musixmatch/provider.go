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
	"time"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

const (
	baseURL       = "https://www.musixmatch.com/ws/1.1/"
	appID         = "web-desktop-app-v1.0"
	userAgent     = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
	sessionCookie = "mxm_bab=AB"
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
	return &Provider{
		client: args.Client,
		signer: newSigner(args.Cache, args.Log),
		log:    args.Log,
	}, nil
}

type Provider struct {
	client *http.Client
	signer *signer
	log    *slog.Logger
}

func (p *Provider) Name() string               { return "Musixmatch" }
func (p *Provider) Desc() string               { return "Extensive library and good word-level sync coverage" }
func (p *Provider) MaxLevel() lyrics.SyncLevel { return lyrics.SyncWord }

func (p *Provider) Search(ctx context.Context, q lyrics.Query) (*lyrics.Result, error) {
	level, err := p.fetchTrackLevel(ctx, q.Track.ISRC)
	if err != nil {
		return nil, err
	}

	lines, err := p.fetchTier(ctx, tierByLevel[level], q.Track.ISRC)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, lyrics.ErrNotFound
	}
	return &lyrics.Result{Lines: lines, SyncLevel: level}, nil
}

func (p *Provider) fetchTrackLevel(ctx context.Context, isrc string) (lyrics.SyncLevel, error) {
	body, err := p.get(ctx, "track.get", url.Values{"track_isrc": {isrc}})
	if err != nil {
		return 0, err
	}
	var resp struct {
		Track struct {
			HasRichsync  int `json:"has_richsync"`
			HasSubtitles int `json:"has_subtitles"`
			HasLyrics    int `json:"has_lyrics"`
		} `json:"track"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return 0, lyrics.ErrNotFound
	}
	switch {
	case resp.Track.HasRichsync == 1:
		return lyrics.SyncWord, nil
	case resp.Track.HasSubtitles == 1:
		return lyrics.SyncLine, nil
	case resp.Track.HasLyrics == 1:
		return lyrics.SyncNone, nil
	}
	return 0, lyrics.ErrNotFound
}

func (p *Provider) fetchTier(ctx context.Context, t tier, isrc string) ([]lyrics.Line, error) {
	params := url.Values{"track_isrc": {isrc}}
	maps.Copy(params, t.extra)

	body, err := p.get(ctx, t.endpoint, params)
	if err != nil {
		if errors.Is(err, lyrics.ErrNotFound) {
			return nil, nil
		}
		if ctx.Err() == nil && !errors.Is(err, providers.ErrRateLimited) {
			p.log.Debug("fetch failed", "tier", t.bodyKey, "isrc", isrc, "err", err)
		}
		return nil, err
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

func buildURL(endpoint string, extra url.Values) string {
	q := url.Values{}
	q.Set("app_id", appID)
	q.Set("format", "json")
	maps.Copy(q, extra)
	return baseURL + endpoint + "?" + q.Encode()
}

func (p *Provider) get(ctx context.Context, endpoint string, extra url.Values) (json.RawMessage, error) {
	secret, err := p.signer.Get(ctx)
	if err != nil {
		return nil, err
	}
	body, err := p.sendSigned(ctx, endpoint, extra, secret)
	if errors.Is(err, errInvalidSignature) {
		secret, err = p.signer.Refresh()
		if err != nil {
			return nil, fmt.Errorf("refresh secret: %w", err)
		}
		body, err = p.sendSigned(ctx, endpoint, extra, secret)
	}
	return body, err
}

func (p *Provider) sendSigned(ctx context.Context, endpoint string, extra url.Values, secret []byte) (json.RawMessage, error) {
	signed := signWith(buildURL(endpoint, extra), secret, time.Now())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, signed, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://www.musixmatch.com/")
	req.Header.Set("Cookie", sessionCookie)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusServiceUnavailable: // could be actual server errors?
		return nil, providers.ErrRateLimited
	default:
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
	case res.Message.Header.StatusCode == 401 && res.Message.Header.Hint == "invalid_signature":
		return nil, errInvalidSignature
	case res.Message.Header.StatusCode == 401 && res.Message.Header.Hint == "captcha":
		return nil, providers.ErrRateLimited
	case res.Message.Header.StatusCode == 404:
		return nil, lyrics.ErrNotFound
	}
	return nil, fmt.Errorf("api %d (%s)", res.Message.Header.StatusCode, res.Message.Header.Hint)
}
