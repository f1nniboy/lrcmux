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
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/f1nniboy/lrcmux/internal/cache"
	"github.com/f1nniboy/lrcmux/internal/utils"
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
	group  singleflight.Group
}

func newSigner(client *http.Client, c cache.Cache, log *slog.Logger) *signer {
	return &signer{
		client: client,
		cache:  c,
		log:    log,
	}
}

func (s *signer) Get(ctx context.Context) ([]byte, error) {
	if s.cache != nil {
		val, status, err := cache.Get[string](ctx, s.cache, signerCacheKey)
		if err != nil {
			return nil, err
		}
		if status == cache.Found && val != "" {
			return []byte(val), nil
		}
	}
	return s.Refresh()
}

func (s *signer) Refresh() ([]byte, error) {
	v, err, _ := s.group.Do("refresh", func() (any, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		secret, err := s.scrape(ctx)
		if err != nil {
			return nil, err
		}
		if s.cache != nil {
			if err := cache.Set(ctx, s.cache, signerCacheKey, string(secret), 0); err != nil {
				s.log.Warn("signing key cache set failed", "err", err)
			}
		}
		s.log.Info("got signing key", "bytes", len(secret))
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	return v.([]byte), nil
}

func (s *signer) scrape(ctx context.Context) ([]byte, error) {
	html, err := s.fetchText(ctx, searchPage)
	if err != nil {
		return nil, fmt.Errorf("search page: %w", err)
	}
	matches := appChunkRE.FindAllStringSubmatch(html, -1)
	if len(matches) == 0 {
		return nil, errors.New("no app chunk in search page")
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
	decoded, err := base64.StdEncoding.DecodeString(utils.ReverseString(m[1]))
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
