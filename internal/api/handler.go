package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/danielgtaylor/huma/v2"

	"github.com/f1nniboy/lrcmux/internal/format"
	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/orchestrator"
	"github.com/f1nniboy/lrcmux/internal/ratelimit"
	"github.com/f1nniboy/lrcmux/internal/utils"
)

type GetLyricsInput struct {
	Artist   string `query:"artist" doc:"Artist name" example:"Rick Astley"`
	Title    string `query:"title" doc:"Song title" example:"Never Gonna Give You Up"`
	Album    string `query:"album" doc:"Album name"`
	Duration int64  `query:"duration" doc:"Track duration in seconds"`
	ISRC     string `query:"isrc" doc:"ISRC of the track, has priority over artist and title"`
	Level    string `query:"level" doc:"Highest sync level to accept, or exact level if strict is set" enum:"word,line,none" default:"word"`
	Format   string `query:"format" doc:"Response format" enum:"lrc,txt,json,srt,vtt" default:"json"`
	Strict   bool   `query:"strict" doc:"Fail instead of falling back to a lower sync level"`
	Fetch    string `query:"fetch" doc:"Cache strategy" enum:"default,cache,force" default:"default"`
}

var responseHeaders = map[string]*huma.Param{
	"X-Source":     {Schema: &huma.Schema{Type: "string"}, Description: "Provider that supplied the lyrics"},
	"X-Sync-Level": {Schema: &huma.Schema{Type: "string"}, Description: "Actual sync level of the returned lyrics, may be lower than requested when `strict=false`"},
	"X-Cache":      {Schema: &huma.Schema{Type: "string"}, Description: "`HIT` if served from cache, `MISS` if freshly fetched from a provider"},
}

func (s *Server) getOp() huma.Operation {
	return huma.Operation{
		OperationID: "get-lyrics",
		Method:      http.MethodGet,
		Path:        "/get",
		Summary:     "Get lyrics for a song",
		Description: "Searches various providers and returns the best available result.",
		Tags:        []string{"Lyrics"},
		Responses: map[string]*huma.Response{
			"200": {
				Description: "Lyrics in the requested format",
				Headers:     responseHeaders,
				Content: map[string]*huma.MediaType{
					"text/plain": {
						Schema: &huma.Schema{Type: "string", Description: "Lyrics in the requested format"},
					},
					"application/json": {
						Schema: s.api.OpenAPI().Components.Schemas.Schema(
							reflect.TypeFor[format.JSONResponse](), true, "LyricsJSON",
						),
					},
				},
			},
		},
	}
}

func (s *Server) handleGet(ctx context.Context, input *GetLyricsInput) (*huma.StreamResponse, error) {
	fetchMode := input.Fetch
	if fetchMode == "" {
		fetchMode = "default"
	}
	if fetchMode == "force" && !utils.IsPrivateIP(clientIP(ctx)) {
		return nil, huma.Error403Forbidden("you can't force-refresh, sorry")
	}

	level, err := lyrics.ParseLevel(input.Level)
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error())
	}

	encoder, err := format.Get(input.Format)
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error())
	}
	if min := encoder.MinLevel(); min > level {
		level = min
	}

	resp, err := s.fetch(ctx, orchestrator.Request{
		Artist:    input.Artist,
		Title:     input.Title,
		Album:     input.Album,
		Duration:  input.Duration,
		ISRC:      input.ISRC,
		Level:     level,
		Strict:    input.Strict,
		FetchMode: fetchMode,
	})
	if err != nil {
		if e, ok := errors.AsType[*ratelimit.LimitError](err); ok {
			return nil, huma.ErrorWithHeaders(
				huma.Error429TooManyRequests("rate limit exceeded"),
				http.Header{"Retry-After": {strconv.Itoa(int(e.RetryAfter.Seconds()))}},
			)
		}
		return nil, s.mapError(err)
	}

	if min := encoder.MinLevel(); resp.Result.SyncLevel < min {
		return nil, huma.Error400BadRequest(fmt.Sprintf("format %q requires %s-synced lyrics", input.Format, min.String()))
	}

	if s.hide {
		resp.Result.Source = lyrics.Source{}
	}

	var buf bytes.Buffer
	if err := encoder.Encode(&buf, resp.Result); err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}

	filename := fmt.Sprintf("%s - %s.%s", utils.SanitizeFilename(resp.Result.Track.Artist), utils.SanitizeFilename(resp.Result.Track.Title), input.Format)

	return &huma.StreamResponse{
		Body: func(ctx huma.Context) {
			ctx.SetHeader("Content-Type", encoder.ContentType())
			ctx.SetHeader("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, filename))
			if !s.hide {
				ctx.SetHeader("X-Source", resp.Result.Source.ID)
			}
			ctx.SetHeader("X-Sync-Level", resp.Result.SyncLevel.String())
			if resp.Cached {
				ctx.SetHeader("X-Cache", "HIT")
			} else {
				ctx.SetHeader("X-Cache", "MISS")
			}
			ctx.SetStatus(http.StatusOK)
			ctx.BodyWriter().Write(buf.Bytes())
		},
	}, nil
}

func (s *Server) fetch(ctx context.Context, req orchestrator.Request) (*orchestrator.Response, error) {
	if req.ISRC == "" && (req.Artist == "" || req.Title == "") {
		return nil, huma.Error400BadRequest("provide either ISRC or both artist and title")
	}

	if s.rl != nil {
		ip := clientIP(ctx)
		req.Charge = func(ctx context.Context) error {
			return s.rl.Allow(ctx, ip)
		}
	}

	req.Artist, req.Title = utils.CleanQuery(req.Artist, req.Title)
	return s.orch.Get(ctx, req)
}

func (s *Server) mapError(err error) error {
	if e, ok := errors.AsType[*huma.ErrorModel](err); ok {
		return e
	}

	switch {
	case errors.Is(err, ratelimit.ErrRateLimited):
		return huma.Error429TooManyRequests("rate limit exceeded")
	case errors.Is(err, orchestrator.ErrNoProviders):
		return huma.Error503ServiceUnavailable("no providers available")
	case errors.Is(err, orchestrator.ErrNotFound):
		return huma.Error404NotFound("no lyrics found for the given query")
	default:
		s.log.Error("provider error", "err", err)
		return huma.Error500InternalServerError("internal error")
	}
}
