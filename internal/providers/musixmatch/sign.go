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
	chunkRE = regexp.MustCompile(`src="([^"]*/_next/static/chunks/[^"]+\.js)"`)
	// a chunk contains the key reversed and base64 encoded
	keyRE    = regexp.MustCompile(`from\(\s*"([^"]*)"\s*\.split`)
	hexKeyRE = regexp.MustCompile(`^[0-9a-f]{32}$`)
)

// the home page is a separate landing app whose key the lyrics api rejects
func (p *Provider) scrape(ctx context.Context) (string, error) {
	page, err := p.fetch(ctx, siteURL+"/discover")
	if err != nil {
		return "", fmt.Errorf("lyrics page: %w", err)
	}
	chunks := chunkRE.FindAllStringSubmatch(page, -1)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	type found struct {
		err error
		key string
	}
	ch := make(chan found, len(chunks))
	for _, m := range chunks {
		go func() {
			js, err := p.fetch(ctx, m[1])
			key, _ := extractKey(js)
			ch <- found{key: key, err: err}
		}()
	}

	var firstErr error
	for range chunks {
		f := <-ch
		if f.key != "" {
			return f.key, nil
		}
		if firstErr == nil {
			firstErr = f.err
		}
	}
	if firstErr != nil {
		return "", fmt.Errorf("chunk: %w", firstErr)
	}
	return "", errors.New("signing key not found in any chunk")
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
