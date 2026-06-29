package musixmatch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/f1nniboy/lrcmux/internal/cache"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

const tokenCacheKey = "mxm:user_token"
const tokenTTL = 24 * time.Hour

var errRenew = errors.New("token renewal required")

type tokenClient struct {
	client *http.Client
	cache  cache.Cache
	log    *slog.Logger
	mu     sync.Mutex
	token  string
}

func newTokenClient(client *http.Client, c cache.Cache, log *slog.Logger) *tokenClient {
	return &tokenClient{client: client, cache: c, log: log}
}

func (tc *tokenClient) Get(ctx context.Context) (string, error) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if tc.token != "" {
		return tc.token, nil
	}
	if tc.cache != nil {
		val, status, err := cache.Get[string](ctx, tc.cache, tokenCacheKey)
		if err == nil && status == cache.Found && val != "" {
			tc.token = val
			return tc.token, nil
		}
	}
	return tc.fetch(ctx)
}

func (tc *tokenClient) Refresh(ctx context.Context) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.token = ""
	_, err := tc.fetch(ctx)
	return err
}

func (tc *tokenClient) fetch(ctx context.Context) (string, error) {
	params := url.Values{"user_language": {"en"}, "app_id": {desktopAppID}}
	u := desktopBaseURL + "token.get?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", desktopUserAgent)

	resp, err := tc.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("token read: %w", err)
	}

	var res struct {
		Message struct {
			Header struct {
				StatusCode int    `json:"status_code"`
				Hint       string `json:"hint"`
			} `json:"header"`
			Body struct {
				UserToken string `json:"user_token"`
			} `json:"body"`
		} `json:"message"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return "", fmt.Errorf("token decode: %w", err)
	}

	if res.Message.Header.StatusCode == 401 && res.Message.Header.Hint == "captcha" {
		return "", providers.ErrRateLimited
	}
	if res.Message.Header.StatusCode != 200 || res.Message.Body.UserToken == "" {
		return "", fmt.Errorf("token api %d (%s)", res.Message.Header.StatusCode, res.Message.Header.Hint)
	}

	tc.token = res.Message.Body.UserToken
	if tc.cache != nil {
		if err := cache.Set(ctx, tc.cache, tokenCacheKey, tc.token, tokenTTL); err != nil {
			tc.log.Warn("token cache set failed", "err", err)
		}
	}
	tc.log.Debug("usertoken refreshed")
	return tc.token, nil
}
