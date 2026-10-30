package deezer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/f1nniboy/lrcmux/internal/providers"
)

const authURL = "https://auth.deezer.com/login/anonymous?jo=p&rto=c&i=c"

// an anonymous jwt, its expiry is never read since a rejection renews it
func (p *Provider) mint(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, authURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := p.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("token: %w", err)
	}
	defer resp.Body.Close()
	if blocked(resp.StatusCode) {
		return "", providers.RateLimitedFrom(resp)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token: http %d", resp.StatusCode)
	}

	var body struct {
		JWT string `json:"jwt"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("token: %w", err)
	}
	if body.JWT == "" {
		return "", errors.New("token: empty jwt")
	}
	return body.JWT, nil
}
