package ytmusic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/providers"
)

const baseURL = "https://music.youtube.com/youtubei/v1"

type Provider struct {
	providers.Common
}

func (p *Provider) ID() string                 { return "ytmusic" }
func (p *Provider) Name() string               { return "YouTube Music" }
func (p *Provider) URL() string                { return "https://music.youtube.com" }
func (p *Provider) Desc() string               { return "Uses various sources, mix of line-synced and plain text" }
func (p *Provider) MaxLevel() lyrics.SyncLevel { return lyrics.SyncLine }

func (p *Provider) Search(ctx context.Context, q lyrics.Query) (*lyrics.Result, error) {
	videoID, err := p.searchVideoID(ctx, q)
	if err != nil {
		return nil, err
	}

	browseID, err := p.resolveBrowseID(ctx, videoID)
	if err != nil {
		return nil, err
	}

	return p.fetchLyrics(ctx, browseID)
}

func (p *Provider) searchVideoID(ctx context.Context, q lyrics.Query) (string, error) {
	body := map[string]any{
		"query":   q.Track.Title + " " + q.Track.Artist,
		"params":  "EgWKAQIIAWoMEA4QChADEAQQCRAF",
		"context": webContext(),
	}
	var resp searchResp
	if err := p.post(ctx, "search", body, &resp); err != nil {
		return "", err
	}
	id := resp.videoID(q.Track.Title, q.Track.Artist)
	if id == "" {
		return "", lyrics.ErrNotFound
	}
	return id, nil
}

func (p *Provider) resolveBrowseID(ctx context.Context, videoID string) (string, error) {
	body := map[string]any{
		"videoId":                       videoID,
		"playlistId":                    "RDAMVM" + videoID,
		"enablePersistentPlaylistPanel": true,
		"isAudioOnly":                   true,
		"tunerSettingValue":             "AUTOMIX_SETTING_NORMAL",
		"watchEndpointMusicSupportedConfigs": map[string]any{
			"watchEndpointMusicConfig": map[string]any{
				"hasPersistentPlaylistPanel": true,
				"musicVideoType":             "MUSIC_VIDEO_TYPE_ATV",
			},
		},
		"context": webContext(),
	}
	var resp nextResp
	if err := p.post(ctx, "next", body, &resp); err != nil {
		return "", err
	}
	id := resp.browseID()
	if id == "" {
		return "", lyrics.ErrNotFound
	}
	return id, nil
}

func (p *Provider) fetchLyrics(ctx context.Context, browseID string) (*lyrics.Result, error) {
	body := map[string]any{
		"browseId": browseID,
		"context":  androidContext(),
	}
	var resp timedBrowseResp
	if err := p.post(ctx, "browse", body, &resp); err != nil {
		return nil, err
	}
	lines, _ := resp.parse()
	if len(lines) == 0 {
		return nil, lyrics.ErrNotFound
	}
	level := lyrics.SyncNone
	if isSynced(lines) {
		level = lyrics.SyncLine
	}
	return &lyrics.Result{
		Lines:     lines,
		SyncLevel: level,
	}, nil
}

func isSynced(lines []lyrics.Line) bool {
	for _, l := range lines {
		if l.StartMs != 0 || l.EndMs != 0 {
			return true
		}
	}
	return false
}

func (p *Provider) post(ctx context.Context, endpoint string, body any, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/"+endpoint+"?alt=json", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://music.youtube.com")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:88.0) Gecko/20100101 Firefox/88.0")

	resp, err := p.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return lyrics.ErrNotFound
	default:
		return fmt.Errorf("status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	return nil
}

func webContext() map[string]any {
	return map[string]any{
		"client": map[string]any{
			"clientName":    "WEB_REMIX",
			"clientVersion": clientVersion(),
		},
		"user": map[string]any{},
	}
}

func androidContext() map[string]any {
	return map[string]any{
		"client": map[string]any{
			"clientName":    "ANDROID_MUSIC",
			"clientVersion": "7.21.50",
		},
		"user": map[string]any{},
	}
}

func clientVersion() string {
	t := time.Now().UTC()
	return fmt.Sprintf("1.%04d%02d%02d.01.00", t.Year(), t.Month(), t.Day())
}
