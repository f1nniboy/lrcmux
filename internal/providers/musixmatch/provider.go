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

const (
	baseURL       = "https://www.musixmatch.com/ws/1.1/"
	appID         = "web-desktop-app-v1.0"
	userAgent     = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
	sessionCookie = "mxm_bab=AB"
)

type tier struct {
	endpoint     string
	extra        url.Values
	syncLevel    lyrics.SyncLevel
	bodyKey      string
	contentField string
	parse        func(string) []lyrics.Line
}

var tiers = []tier{
	{endpoint: "track.richsync.get", syncLevel: lyrics.SyncWord, bodyKey: "richsync", contentField: "richsync_body", parse: parseRichsync},
	{endpoint: "track.subtitle.get", syncLevel: lyrics.SyncLine, bodyKey: "subtitle", contentField: "subtitle_body", parse: parseSubtitles, extra: url.Values{"subtitle_format": {"mxm"}}},
	{endpoint: "track.lyrics.get", syncLevel: lyrics.SyncNone, bodyKey: "lyrics", contentField: "lyrics_body", parse: parseLyrics},
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
	if q.ISRC == "" {
		return nil, lyrics.ErrNotFound
	}
	if err := p.signer.Ensure(); err != nil {
		return nil, fmt.Errorf("bootstrap: %w", err)
	}

	for _, t := range tiers {
		lines, err := p.fetchTier(ctx, t, q.ISRC)
		if err != nil {
			return nil, err
		}
		if len(lines) > 0 {
			return &lyrics.Result{Lines: lines, SyncLevel: t.syncLevel}, nil
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return nil, lyrics.ErrNotFound
}

func (p *Provider) fetchTier(ctx context.Context, t tier, isrc string) ([]lyrics.Line, error) {
	params := url.Values{"track_isrc": {isrc}}
	maps.Copy(params, t.extra)

	body, err := p.get(ctx, t.endpoint, params)
	if err != nil {
		if errors.Is(err, providers.ErrRateLimited) {
			return nil, err
		}
		if ctx.Err() == nil && !errors.Is(err, lyrics.ErrNotFound) {
			p.log.Debug("fetch failed", "tier", t.bodyKey, "isrc", isrc, "err", err)
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

func buildURL(endpoint string, extra url.Values) string {
	q := url.Values{}
	q.Set("app_id", appID)
	q.Set("format", "json")
	maps.Copy(q, extra)
	return baseURL + endpoint + "?" + q.Encode()
}

type mxmResponse struct {
	Message struct {
		Header struct {
			StatusCode int    `json:"status_code"`
			Hint       string `json:"hint"`
		} `json:"header"`
		Body json.RawMessage `json:"body"`
	} `json:"message"`
}

// Refreshes the secret and retries once on invalid_signature.
func (p *Provider) get(ctx context.Context, endpoint string, extra url.Values) (json.RawMessage, error) {
	body, err := p.sendSigned(ctx, endpoint, extra)
	if errors.Is(err, errInvalidSignature) {
		if rerr := p.signer.Refresh(); rerr != nil {
			return nil, fmt.Errorf("refresh secret: %w", rerr)
		}
		body, err = p.sendSigned(ctx, endpoint, extra)
	}
	return body, err
}

func (p *Provider) sendSigned(ctx context.Context, endpoint string, extra url.Values) (json.RawMessage, error) {
	signed, err := p.signer.Sign(buildURL(endpoint, extra))
	if err != nil {
		return nil, err
	}
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
	case http.StatusServiceUnavailable:
		// 503 here is a rate-limit symptom (the edge serves wrong-vhost HTML).
		return nil, providers.ErrRateLimited
	default:
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	var r mxmResponse
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	switch {
	case r.Message.Header.StatusCode == 200:
		return r.Message.Body, nil
	case r.Message.Header.StatusCode == 401 && r.Message.Header.Hint == "invalid_signature":
		return nil, errInvalidSignature
	case r.Message.Header.StatusCode == 401 && r.Message.Header.Hint == "captcha":
		return nil, providers.ErrRateLimited
	case r.Message.Header.StatusCode == 404:
		return nil, lyrics.ErrNotFound
	}
	p.log.Debug("api error", "status", r.Message.Header.StatusCode, "hint", r.Message.Header.Hint)
	return nil, fmt.Errorf("api %d (%s)", r.Message.Header.StatusCode, r.Message.Header.Hint)
}
