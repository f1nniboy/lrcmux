package musixmatch

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"

	"github.com/f1nniboy/lrcmux/internal/cache"
)

const signerCacheKey = "mxm:signing_key"

const searchPage = "https://www.musixmatch.com/search"

var (
	appChunkRE = regexp.MustCompile(`src="([^"]*/_next/static/chunks/pages/_app-[^"]+\.js)"`)
	keyRE      = regexp.MustCompile(`from\(\s*"(.*?)"\s*\.split`)
)

var errInvalidSignature = errors.New("invalid signature")

type signer struct {
	client *http.Client
	cache  cache.Cache
	log    *slog.Logger
	mu     sync.Mutex
	secret []byte
}

func newSigner(c cache.Cache, log *slog.Logger) *signer {
	return &signer{
		client: &http.Client{Timeout: 10 * time.Second},
		cache:  c,
		log:    log,
	}
}

func (s *signer) Ensure() error {
	_, err := s.load(false)
	return err
}

func (s *signer) Sign(canonical string) (string, error) {
	secret, err := s.load(false)
	if err != nil {
		return "", err
	}
	return signWith(canonical, secret, time.Now()), nil
}

// called after an invalid_signature response, to get the new signing key
func (s *signer) Refresh() error {
	_, err := s.load(true)
	return err
}

func (s *signer) load(force bool) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !force && s.secret != nil {
		return s.secret, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if !force && s.cache != nil {
		if val, status, err := cache.Get[string](ctx, s.cache, signerCacheKey); err == nil && status == cache.Found && val != "" {
			s.secret = []byte(val)
			s.log.Debug("signing key loaded from cache", "bytes", len(s.secret))
			return s.secret, nil
		}
	}

	secret, err := s.scrape(ctx)
	if err != nil {
		return nil, err
	}
	s.secret = secret
	if s.cache != nil {
		if err := cache.Set(ctx, s.cache, signerCacheKey, string(secret), 24*time.Hour); err != nil {
			s.log.Warn("signing key cache set failed", "err", err)
		}
	}
	s.log.Info("signing key refreshed", "bytes", len(secret))
	return secret, nil
}

func (s *signer) scrape(ctx context.Context) ([]byte, error) {
	html, err := s.fetchText(ctx, searchPage)
	if err != nil {
		return nil, fmt.Errorf("search page: %w", err)
	}
	matches := appChunkRE.FindAllStringSubmatch(html, -1)
	if len(matches) == 0 {
		return nil, errors.New("no _app chunk in search page")
	}
	chunkURL := matches[len(matches)-1][1]
	js, err := s.fetchText(ctx, chunkURL)
	if err != nil {
		return nil, fmt.Errorf("app chunk: %w", err)
	}
	m := keyRE.FindStringSubmatch(js)
	if len(m) < 2 {
		return nil, errors.New("encoded key not found in app chunk")
	}
	decoded, err := base64.StdEncoding.DecodeString(reverseString(m[1]))
	if err != nil {
		return nil, fmt.Errorf("decode key: %w", err)
	}
	return decoded, nil
}

func (s *signer) fetchText(ctx context.Context, u string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Cookie", sessionCookie)
	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func signWith(canonical string, secret []byte, now time.Time) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(canonical))
	mac.Write([]byte(now.UTC().Format("20060102")))
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return canonical + "&signature=" + url.QueryEscape(sig) + "&signature_protocol=sha256"
}

func reverseString(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
