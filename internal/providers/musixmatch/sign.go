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
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/f1nniboy/lrcmux/internal/providers"
)

var (
	bundleRE = regexp.MustCompile(`src="([^"]*/_next/static/chunks/pages/_app-[^"]+\.js)"`)
	// the key ships in the app bundle reversed and base64 encoded,
	// as Buffer.from("...".split("").reverse().join(""), "base64")
	keyRE    = regexp.MustCompile(`from\(\s*"(.*?)"\s*\.split`)
	hexKeyRE = regexp.MustCompile(`^[0-9a-f]{32}$`)
)

type signer struct {
	client *http.Client
	sf     singleflight.Group

	current string
	mu      sync.Mutex
}

func (s *signer) key(ctx context.Context) (string, error) {
	s.mu.Lock()
	k := s.current
	s.mu.Unlock()
	if k != "" {
		return k, nil
	}
	return s.refresh(ctx, "")
}

// stale is the key the caller failed with, so a caller that lost the race
// to an already finished refresh takes the new key instead of scraping again
func (s *signer) refresh(ctx context.Context, stale string) (string, error) {
	v, err, _ := s.sf.Do("key", func() (any, error) {
		s.mu.Lock()
		k := s.current
		s.mu.Unlock()
		if k != "" && k != stale {
			return k, nil
		}

		k, err := s.scrape(ctx)
		if err != nil {
			return "", err
		}
		s.mu.Lock()
		s.current = k
		s.mu.Unlock()
		return k, nil
	})
	if err != nil {
		return "", err
	}
	return v.(string), nil
}

func (s *signer) scrape(ctx context.Context) (string, error) {
	home, err := s.fetch(ctx, siteURL+"/")
	if err != nil {
		return "", fmt.Errorf("home page: %w", err)
	}
	m := bundleRE.FindStringSubmatch(home)
	if m == nil {
		return "", errors.New("app bundle not found")
	}
	js, err := s.fetch(ctx, m[1])
	if err != nil {
		return "", fmt.Errorf("app bundle: %w", err)
	}
	key, ok := extractKey(js)
	if !ok {
		return "", errors.New("signing key not found in app bundle")
	}
	return key, nil
}

func (s *signer) fetch(ctx context.Context, u string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	resp, err := s.client.Do(req)
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

// the signed message is the url plus the utc date, so a signature is only
// valid for the day it was made
func sign(rawURL, key string, now time.Time) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(rawURL + now.UTC().Format("20060102")))
	return rawURL + "&" + url.Values{
		"signature":          {base64.StdEncoding.EncodeToString(mac.Sum(nil))},
		"signature_protocol": {"sha256"},
	}.Encode()
}
