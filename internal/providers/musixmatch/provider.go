package musixmatch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

// based on https://github.com/omkarcloud/musixmatch-scraper

// only plain lyrics are served anonymously, track.subtitle.get and
// track.richsync.get answer 403 without a paid plan, and the desktop api
// (apic-desktop) hands out all-zero tokens instead
const (
	siteURL = "https://www.musixmatch.com"
	apiURL  = siteURL + "/ws/1.1/"
	appID   = "community-app-v1.0"
)

var errBadSignature = errors.New("invalid signature")

//nolint:govet // fieldalignment
type Provider struct {
	providers.Common

	signer *signer
}

func (p *Provider) ID() string                 { return "musixmatch" }
func (p *Provider) Name() string               { return "Musixmatch" }
func (p *Provider) URL() string                { return siteURL }
func (p *Provider) Desc() string               { return "Extensive library, but only plain text lyrics" }
func (p *Provider) MaxLevel() lyrics.SyncLevel { return lyrics.SyncNone }

func (p *Provider) Init() {
	p.signer = &signer{client: p.HTTP}
}

func (p *Provider) Search(ctx context.Context, q lyrics.Query) (*lyrics.Result, error) {
	body, err := p.get(ctx, "track.lyrics.get", url.Values{"track_isrc": {q.Track.ISRC}})
	if err != nil {
		return nil, err
	}

	var resp struct {
		Lyrics struct {
			Body         string `json:"lyrics_body"`
			Instrumental int    `json:"instrumental"`
		} `json:"lyrics"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if resp.Lyrics.Instrumental == 1 {
		return &lyrics.Result{Instrumental: true}, nil
	}
	lines := lyrics.ParsePlain(resp.Lyrics.Body)
	if len(lines) == 0 {
		return nil, lyrics.ErrNotFound
	}
	return &lyrics.Result{Lines: lines, SyncLevel: lyrics.SyncNone}, nil
}

func (p *Provider) get(ctx context.Context, method string, params url.Values) (json.RawMessage, error) {
	params.Set("app_id", appID)
	params.Set("format", "json")

	key, err := p.signer.key(ctx)
	if err != nil {
		return nil, err
	}
	body, err := p.do(ctx, method, params, key)
	if errors.Is(err, errBadSignature) {
		// the key rotates with every site deploy
		if key, err = p.signer.refresh(ctx, key); err != nil {
			return nil, err
		}
		body, err = p.do(ctx, method, params, key)
	}
	return body, err
}

func (p *Provider) do(ctx context.Context, method string, params url.Values, key string) (json.RawMessage, error) {
	u := sign(apiURL+method+"?"+params.Encode(), key, time.Now())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	// a navigation-style accept gets a varnish 503
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Origin", siteURL)
	req.Header.Set("Referer", siteURL+"/")

	resp, err := p.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusUnauthorized:
		return nil, errBadSignature
	case blocked(resp.StatusCode):
		return nil, providers.ErrRateLimited
	case resp.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	default:
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
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	switch res.Message.Header.StatusCode {
	case http.StatusOK:
		return res.Message.Body, nil
	case http.StatusNotFound:
		return nil, lyrics.ErrNotFound
	default:
		return nil, fmt.Errorf("api %d (%s)", res.Message.Header.StatusCode, res.Message.Header.Hint)
	}
}

// aws waf answers a challenge with 202 and a captcha with 405
func blocked(status int) bool {
	switch status {
	case http.StatusAccepted, http.StatusForbidden, http.StatusMethodNotAllowed, http.StatusTooManyRequests:
		return true
	default:
		return false
	}
}
