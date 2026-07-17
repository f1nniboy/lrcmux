package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/f1nniboy/lrcmux/internal/format"
	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/orchestrator"
)

type LrclibInput struct {
	TrackName  string  `query:"track_name"  doc:"Track name"`
	ArtistName string  `query:"artist_name" doc:"Artist name"`
	AlbumName  string  `query:"album_name"  doc:"Album name"`
	ISRC       string  `query:"isrc"        doc:"ISRC of the track"`
	Duration   float64 `query:"duration"    doc:"Track duration in seconds"`
}

type LrclibOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         LrclibResponse
}

type LrclibResponse struct {
	SyncedLyrics *string `json:"syncedLyrics"`
	TrackName    string  `json:"trackName"`
	ArtistName   string  `json:"artistName"`
	AlbumName    string  `json:"albumName"`
	PlainLyrics  string  `json:"plainLyrics"`
	ID           int     `json:"id"`
	Duration     float64 `json:"duration"`
	Instrumental bool    `json:"instrumental"`
}

func (s *Server) lrclibOp() huma.Operation {
	return huma.Operation{
		OperationID: "lrclib-get-lyrics",
		Method:      http.MethodGet,
		Path:        "/compat/lrclib/api/get",
		Summary:     "LRCLIB",
		Tags:        []string{"Compatibility"},
	}
}

type LrclibSearchInput struct {
	TrackName  string `query:"track_name"  doc:"Track name"`
	ArtistName string `query:"artist_name" doc:"Artist name"`
	AlbumName  string `query:"album_name"  doc:"Album name"`
}

type LrclibSearchOutput struct {
	Body []LrclibResponse
}

func (s *Server) lrclibSearchOp() huma.Operation {
	return huma.Operation{
		OperationID: "lrclib-search-lyrics",
		Method:      http.MethodGet,
		Path:        "/compat/lrclib/api/search",
		Summary:     "LRCLIB",
		Tags:        []string{"Compatibility"},
	}
}

func lrclibResponse(result *orchestrator.Response) LrclibResponse {
	txtEnc, _ := format.Get("txt")
	lrcEnc, _ := format.Get("lrc")

	var plain bytes.Buffer
	txtEnc.Encode(&plain, result.Result)

	var syncedLyrics *string
	if result.Result.SyncLevel >= lyrics.SyncLine {
		var synced bytes.Buffer
		lrcEnc.Encode(&synced, result.Result)
		s := synced.String()
		syncedLyrics = &s
	}

	track := result.Result.Track
	return LrclibResponse{
		TrackName:    track.Title,
		ArtistName:   track.Artist,
		AlbumName:    track.Album,
		Duration:     float64(track.Duration),
		PlainLyrics:  plain.String(),
		SyncedLyrics: syncedLyrics,
	}
}

func (s *Server) handleLrclibSearch(ctx context.Context, input *LrclibSearchInput) (*LrclibSearchOutput, error) {
	result, err := s.fetch(ctx, orchestrator.Request{
		Artist: input.ArtistName,
		Title:  input.TrackName,
		Album:  input.AlbumName,
		Level:  lyrics.SyncLine,
	})
	if err != nil {
		if errors.Is(err, orchestrator.ErrNotFound) {
			return &LrclibSearchOutput{Body: []LrclibResponse{}}, nil
		}
		return nil, s.mapError(err)
	}
	return &LrclibSearchOutput{Body: []LrclibResponse{lrclibResponse(result)}}, nil
}

func (s *Server) handleLrclib(ctx context.Context, input *LrclibInput) (*LrclibOutput, error) {
	result, err := s.fetch(ctx, orchestrator.Request{
		Artist:   input.ArtistName,
		Title:    input.TrackName,
		Album:    input.AlbumName,
		Duration: int64(input.Duration),
		ISRC:     input.ISRC,
		Level:    lyrics.SyncLine,
	})
	if err != nil {
		return nil, s.mapError(err)
	}
	out := &LrclibOutput{Body: lrclibResponse(result)}
	if result.TTL > 0 {
		out.CacheControl = fmt.Sprintf("public, max-age=%d", int(result.TTL.Seconds()))
	}
	return out, nil
}
