package netease

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/normalize"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

const (
	baseURL = "https://music.163.com/api"
)

// borrowed from https://github.com/moehmeni/syncedlyrics/blob/main/syncedlyrics/providers/netease.py#L17
const cookie = "NMTID=00OAVK3xqDG726ITU6jopU6jF2yMk0AAAGCO8l1BA; JSESSIONID-WYYY=8KQo11YK2GZP45RMlz8Kn80vHZ9%2FGvwzRKQXXy0iQoFKycWdBlQjbfT0MJrFa6hwRfmpfBYKeHliUPH287JC3hNW99WQjrh9b9RmKT%2Fg1Exc2VwHZcsqi7ITxQgfEiee50po28x5xTTZXKoP%2FRMctN2jpDeg57kdZrXz%2FD%2FWghb%5C4DuZ%3A1659124633932; _iuqxldmzr_=32; _ntes_nnid=0db6667097883aa9596ecfe7f188c3ec,1659122833973; _ntes_nuid=0db6667097883aa9596ecfe7f188c3ec"

type Provider struct {
	providers.Common
}

func (p *Provider) ID() string                 { return "netease" }
func (p *Provider) Name() string               { return "NetEase" }
func (p *Provider) URL() string                { return "https://music.163.com" }
func (p *Provider) Desc() string               { return "Good coverage of Asian music" }
func (p *Provider) MaxLevel() lyrics.SyncLevel { return lyrics.SyncWord }

type searchResult struct {
	Result struct {
		Songs []apiSong `json:"songs"`
	} `json:"result"`
	Code int `json:"code"`
}

type apiSong struct {
	Name    string      `json:"name"`
	Artists []apiArtist `json:"artists"`
	ID      int64       `json:"id"`
	// in milliseconds
	Duration int64 `json:"duration"`
}

func (s *apiSong) primaryArtist() string {
	if len(s.Artists) == 0 {
		return ""
	}
	return s.Artists[0].Name
}

type apiArtist struct {
	Name string `json:"name"`
}

type lyricsResult struct {
	Lrc struct {
		Lyric string `json:"lyric"`
	} `json:"lrc"`
	Yrc struct {
		Lyric string `json:"lyric"`
	} `json:"yrc"`
	Code int `json:"code"`
}

func (p *Provider) Search(ctx context.Context, q lyrics.Query) (*lyrics.Result, error) {
	song, err := p.findSong(ctx, q)
	if err != nil {
		return nil, err
	}
	return p.fetchLyrics(ctx, song.ID)
}

func (p *Provider) findSong(ctx context.Context, q lyrics.Query) (*apiSong, error) {
	params := url.Values{}
	params.Set("s", q.Track.Title+" "+q.Track.Artist)
	params.Set("type", "1")
	params.Set("limit", "10")
	params.Set("offset", "0")

	var sr searchResult
	if err := p.do(ctx, baseURL+"/search/pc?"+params.Encode(), &sr); err != nil {
		return nil, err
	}
	if len(sr.Result.Songs) == 0 {
		return nil, lyrics.ErrNotFound
	}

	for i, s := range sr.Result.Songs {
		if normalize.Match(q.Track.Title, q.Track.Artist, s.Name, s.primaryArtist()) {
			return &sr.Result.Songs[i], nil
		}
	}
	return nil, lyrics.ErrNotFound
}

func (p *Provider) fetchLyrics(ctx context.Context, id int64) (*lyrics.Result, error) {
	params := url.Values{}
	params.Set("id", fmt.Sprintf("%d", id))
	params.Set("lv", "1")
	params.Set("yv", "1")

	var lr lyricsResult
	if err := p.do(ctx, baseURL+"/song/lyric?"+params.Encode(), &lr); err != nil {
		return nil, err
	}

	if lr.Yrc.Lyric != "" {
		lines := parseYRC(lr.Yrc.Lyric)
		lines = filterCredits(lines)
		if hasPlaceholder(lines) {
			return nil, lyrics.ErrNotFound
		}
		if len(lines) > 0 {
			return &lyrics.Result{Lines: lines, SyncLevel: lyrics.SyncWord}, nil
		}
	}
	if lr.Lrc.Lyric != "" {
		lines, level := lyrics.ParseLRC(lr.Lrc.Lyric)
		lines = filterCredits(applyHalfWidth(lines))
		if hasPlaceholder(lines) {
			return nil, lyrics.ErrNotFound
		}
		if len(lines) > 0 {
			return &lyrics.Result{Lines: lines, SyncLevel: level}, nil
		}
	}

	return nil, lyrics.ErrNotFound
}

func (p *Provider) do(ctx context.Context, endpoint string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Referer", "https://music.163.com/")
	req.Header.Set("Cookie", cookie)
	resp, err := p.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return lyrics.ErrNotFound
	case http.StatusForbidden, http.StatusUnauthorized:
		return providers.ErrRateLimited
	default:
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	return nil
}
