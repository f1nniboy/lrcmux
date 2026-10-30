package musixmatch

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"time"

	"github.com/f1nniboy/lrcmux/internal/providers"
)

var (
	bundleRE = regexp.MustCompile(`src="([^"]*/_next/static/chunks/pages/_app-[^"]+\.js)"`)
	// the bundle ships the key reversed and base64 encoded
	keyRE    = regexp.MustCompile(`from\(\s*"([^"]*)"\s*\.split`)
	hexKeyRE = regexp.MustCompile(`^[0-9a-f]{32}$`)
)

func (p *Provider) scrape(ctx context.Context) (string, error) {
	home, err := p.fetch(ctx, siteURL+"/")
	if err != nil {
		return "", fmt.Errorf("home page: %w", err)
	}
	m := bundleRE.FindStringSubmatch(home)
	if m == nil {
		return "", errors.New("app bundle not found")
	}
	js, err := p.fetch(ctx, m[1])
	if err != nil {
		return "", fmt.Errorf("app bundle: %w", err)
	}
	key, ok := extractKey(js)
	if !ok {
		return "", errors.New("signing key not found in app bundle")
	}
	return key, nil
}

func (p *Provider) fetch(ctx context.Context, u string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	resp, err := p.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if blocked(resp.StatusCode) {
		return "", providers.ErrRateLimited
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	return string(b), err
}

func extractKey(js string) (string, bool) {
	for _, m := range keyRE.FindAllStringSubmatch(js, -1) {
		b := []byte(m[1])
		slices.Reverse(b)
		key, err := base64.StdEncoding.DecodeString(string(b))
		if err == nil && hexKeyRE.Match(key) {
			return string(key), true
		}
	}
	return "", false
}

// the utc date is signed too, so a signature expires daily
func sign(rawURL, key string, now time.Time) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(rawURL + now.UTC().Format("20060102")))
	return rawURL + "&" + url.Values{
		"signature":          {base64.StdEncoding.EncodeToString(mac.Sum(nil))},
		"signature_protocol": {"sha256"},
	}.Encode()
}
