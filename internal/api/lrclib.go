package api

import (
	"bytes"
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/f1nniboy/lrcmux/internal/format"
	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/ratelimit"
)

type LrclibInput struct {
	TrackName  string  `query:"track_name"  doc:"Track name"`
	ArtistName string  `query:"artist_name" doc:"Artist name"`
	AlbumName  string  `query:"album_name"  doc:"Album name"`
	Duration   float64 `query:"duration"    doc:"Track duration in seconds"`
	ISRC       string  `query:"isrc"        doc:"ISRC of the track"`
}

type LrclibOutput struct {
	Body LrclibResponse
}

type LrclibResponse struct {
	TrackName    string  `json:"trackName"`
	ArtistName   string  `json:"artistName"`
	AlbumName    string  `json:"albumName"`
	Duration     float64 `json:"duration"`
	Instrumental bool    `json:"instrumental"`
	PlainLyrics  string  `json:"plainLyrics"`
	SyncedLyrics string  `json:"syncedLyrics"`
}

func (s *Server) lrclibOp() huma.Operation {
	return huma.Operation{
		OperationID: "lrclib-get-lyrics",
		Method:      http.MethodGet,
		Path:        "/api/compat/lrclib/api/get",
		Summary:     "LRCLIB",
		Description: "Drop-in replacement for apps that use the LRCLIB API.",
		Tags:        []string{"Compatibility"},
	}
}

func (s *Server) handleLrclib(ctx context.Context, input *LrclibInput) (*LrclibOutput, error) {
	resp, err := s.fetch(ctx, fetchParams{
		Artist:   input.ArtistName,
		Title:    input.TrackName,
		Album:    input.AlbumName,
		Duration: int64(input.Duration),
		ISRC:     input.ISRC,
		Level:    lyrics.SyncLine,
	})
	if err != nil {
		if errors.Is(err, ratelimit.ErrRateLimited) {
			return nil, huma.Error429TooManyRequests("rate limit exceeded")
		}
		return nil, huma.Error404NotFound("failed to find specified track")
	}

	txtEnc, _ := format.Get("txt")
	lrcEnc, _ := format.Get("lrc")

	var plain bytes.Buffer
	txtEnc.Encode(&plain, resp.Result)

	var synced bytes.Buffer
	if resp.Result.SyncLevel >= lyrics.SyncLine {
		lrcEnc.Encode(&synced, resp.Result)
	}

	return &LrclibOutput{Body: LrclibResponse{
		TrackName:    input.TrackName,
		ArtistName:   input.ArtistName,
		AlbumName:    input.AlbumName,
		Duration:     input.Duration,
		Instrumental: false,
		PlainLyrics:  plain.String(),
		SyncedLyrics: synced.String(),
	}}, nil
}
