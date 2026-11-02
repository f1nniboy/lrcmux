package deezer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

const (
	siteURL = "https://www.deezer.com"
	pipeURL = "https://pipe.deezer.com/api"
)

const lyricsQuery = `query GetLyrics($trackId: String!) {
  track(trackId: $trackId) {
    lyrics {
      text
      synchronizedWordByWordLines { words { start end word } }
      synchronizedLines { line milliseconds duration }
    }
  }
}`

//nolint:govet // fieldalignment
type Provider struct {
	providers.Common

	token *providers.Secret
}

func (p *Provider) ID() string                 { return "deezer" }
func (p *Provider) Name() string               { return "Deezer" }
func (p *Provider) URL() string                { return siteURL }
func (p *Provider) Desc() string               { return "Word-level sync for most songs, uncensored" }
func (p *Provider) MaxLevel() lyrics.SyncLevel { return lyrics.SyncWord }

func (p *Provider) Init() {
	p.token = providers.NewSecret(p.Cache, p.ID(), p.mint)
}

func (p *Provider) Search(ctx context.Context, q lyrics.Query) (*lyrics.Result, error) {
	var res *lyrics.Result
	err := p.token.Do(ctx, func(jwt string) error {
		var err error
		res, err = p.fetch(ctx, jwt, q.Track)
		return err
	})
	return res, err
}

func (p *Provider) fetch(ctx context.Context, jwt string, track lyrics.Track) (*lyrics.Result, error) {
	type variables struct {
		TrackID string `json:"trackId"`
	}
	payload, err := json.Marshal(struct {
		Query     string    `json:"query"`
		Variables variables `json:"variables"`
	}{lyricsQuery, variables{strconv.FormatInt(track.DeezerID, 10)}})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, pipeURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+jwt)

	resp, err := p.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()
	if blocked(resp.StatusCode) {
		return nil, providers.RateLimitedFrom(resp)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}

	var body struct {
		Data *struct {
			Track struct {
				Lyrics lyricsData `json:"lyrics"`
			} `json:"track"`
		} `json:"data"`
		Errors []struct {
			Message    string `json:"message"`
			Extensions struct {
				Type string `json:"type"`
			} `json:"extensions"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	// parse can't tell a failure from missing lyrics, so every other error stops here
	for _, e := range body.Errors {
		switch e.Extensions.Type {
		case "LyricsNotFoundError", "TrackNotFoundError":
		case "JwtTokenExpiredError", "JwtTokenInvalidError":
			return nil, fmt.Errorf("%w: %s", providers.ErrRejected, e.Extensions.Type)
		default:
			return nil, fmt.Errorf("%s: %s", e.Extensions.Type, e.Message)
		}
	}
	if body.Data == nil {
		return nil, errors.New("no data")
	}
	return parse(body.Data.Track.Lyrics, track.Duration)
}

// a 403 is akamai blocking the exit, not a bad request
func blocked(status int) bool {
	return status == http.StatusForbidden || status == http.StatusTooManyRequests
}
