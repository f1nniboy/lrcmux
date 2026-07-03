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

const tokenTTL = 24 * time.Hour

var (
	errRenew   = errors.New("token expired")
	errCaptcha = errors.New("token rate limited")
)

type tokenSlot struct {
	mu    sync.Mutex
	token string
}

type tokenPool struct {
	slots   []*tokenSlot
	mu      sync.RWMutex
	current int
	client  *http.Client
	cache   cache.Cache
	log     *slog.Logger
}

func newTokenPool(n int, client *http.Client, c cache.Cache, log *slog.Logger) *tokenPool {
	slots := make([]*tokenSlot, n)
	for i := range n {
		slots[i] = &tokenSlot{}
	}
	return &tokenPool{slots: slots, client: client, cache: c, log: log}
}

func (p *tokenPool) cacheKey(idx int) string {
	return fmt.Sprintf("musixmatch:token:%d", idx)
}

func (p *tokenPool) get(ctx context.Context) (string, int, error) {
	p.mu.RLock()
	idx := p.current
	p.mu.RUnlock()

	slot := p.slots[idx]
	slot.mu.Lock()
	defer slot.mu.Unlock()

	if slot.token != "" {
		return slot.token, idx, nil
	}

	if p.cache != nil {
		val, status, err := cache.Get[string](ctx, p.cache, p.cacheKey(idx))
		if err == nil && status == cache.Found && val != "" {
			slot.token = val
			return slot.token, idx, nil
		}
	}

	token, err := p.fetch(ctx)
	if err != nil {
		return "", -1, err
	}
	slot.token = token
	if p.cache != nil {
		if err := cache.Set(ctx, p.cache, p.cacheKey(idx), token, tokenTTL); err != nil {
			p.log.Warn("token cache set failed", "slot", idx, "err", err)
		}
	}
	p.log.Debug("token fetched", "slot", idx)
	return token, idx, nil
}

func (p *tokenPool) retire(idx int) {
	slot := p.slots[idx]
	slot.mu.Lock()
	alreadyRetired := slot.token == ""
	slot.token = ""
	slot.mu.Unlock()

	if alreadyRetired {
		return
	}

	if p.cache != nil {
		p.cache.Delete(context.Background(), p.cacheKey(idx))
	}

	p.mu.Lock()
	if p.current == idx {
		p.current = (idx + 1) % len(p.slots)
	}
	p.mu.Unlock()

	go p.refreshSlot(idx)
}

func (p *tokenPool) refreshSlot(idx int) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	token, err := p.fetch(ctx)
	if err != nil {
		p.log.Warn("token slot refresh failed", "slot", idx, "err", err)
		return
	}

	slot := p.slots[idx]
	slot.mu.Lock()
	slot.token = token
	slot.mu.Unlock()

	if p.cache != nil {
		if err := cache.Set(ctx, p.cache, p.cacheKey(idx), token, tokenTTL); err != nil {
			p.log.Warn("token cache set failed", "slot", idx, "err", err)
		}
	}
	p.log.Debug("token slot refreshed", "slot", idx)
}

func (p *tokenPool) fetch(ctx context.Context) (string, error) {
	params := url.Values{"user_language": {"en"}, "app_id": {desktopAppID}}
	u := desktopBaseURL + "token.get?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", desktopUserAgent)

	resp, err := p.client.Do(req)
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
		p.log.Debug("token decode failed", "body", string(raw[:min(len(raw), 256)]))
		return "", fmt.Errorf("token decode: %w", err)
	}

	if res.Message.Header.StatusCode == 401 && res.Message.Header.Hint == "captcha" {
		return "", providers.ErrRateLimited
	}
	if res.Message.Header.StatusCode != 200 || res.Message.Body.UserToken == "" {
		return "", fmt.Errorf("token api %d (%s)", res.Message.Header.StatusCode, res.Message.Header.Hint)
	}

	return res.Message.Body.UserToken, nil
}
