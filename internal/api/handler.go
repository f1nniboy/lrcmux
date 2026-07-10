package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/f1nniboy/lrcmux/internal/format"
	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/normalize"
	"github.com/f1nniboy/lrcmux/internal/orchestrator"
	"github.com/f1nniboy/lrcmux/internal/ratelimit"
)

type GetLyricsInput struct {
	Artist   string `query:"artist" doc:"Artist name" example:"Rick Astley"`
	Title    string `query:"title" doc:"Song title" example:"Never Gonna Give You Up"`
	Album    string `query:"album" doc:"Album name"`
	ISRC     string `query:"isrc" doc:"ISRC of the track, has priority over artist and title"`
	Level    string `query:"level" doc:"Highest sync level to accept, or exact level if strict is set" enum:"word,line,none" default:"word"`
	Format   string `query:"format" doc:"Response format" enum:"lrc,txt,json,srt,vtt,lyricsfile" default:"json"`
	Fetch    string `query:"fetch" doc:"Cache strategy" enum:"default,cache,force" default:"default"`
	Duration int64  `query:"duration" doc:"Track duration in seconds"`
	Strict   bool   `query:"strict" doc:"Fail instead of falling back to a lower sync level"`
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

func (s *Server) handleGet(ctx context.Context, input *GetLyricsInput) (resp *huma.StreamResponse, herr error) {
	internal := isPrivateIP(clientIP(ctx))
	if s.metrics != nil && !internal {
		start := time.Now()
		defer func() {
			status := 200
			if herr != nil {
				if he, ok := errors.AsType[*huma.ErrorModel](herr); ok {
					status = he.Status
				} else {
					status = 500
				}
			}
			labels := []string{input.Format, input.Level, strconv.Itoa(status)}
			s.metrics.HTTPRequests.WithLabelValues(labels...).Inc()
			s.metrics.HTTPLatency.WithLabelValues(labels...).Observe(time.Since(start).Seconds())
		}()
	}

	fetchMode := input.Fetch
	if fetchMode == "" {
		fetchMode = "default"
	}
	if fetchMode == "force" && !internal {
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

	minLevel, maxLevel := encoder.Levels()
	if minLevel > level {
		level = minLevel
	}
	if maxLevel < level {
		level = maxLevel
	}

	lyricsResp, err := s.fetch(ctx, orchestrator.Request{
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
		return nil, s.mapError(err)
	}

	if lyricsResp.Result.SyncLevel < minLevel {
		return nil, huma.Error400BadRequest(fmt.Sprintf("format %q requires %s-synced lyrics", input.Format, minLevel.String()))
	}

	if s.metrics != nil && !internal {
		cacheResult := "miss"
		if lyricsResp.Cached {
			cacheResult = "hit"
		}
		s.metrics.CacheOps.WithLabelValues(cacheResult).Inc()
	}

	if s.cfg.Provider.Hide {
		lyricsResp.Result.Source = lyrics.Source{}
	}

	var buf bytes.Buffer
	if err := encoder.Encode(&buf, lyricsResp.Result); err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}

	filename := fmt.Sprintf("%s - %s.%s", sanitizeFilename(lyricsResp.Result.Track.Artist), sanitizeFilename(lyricsResp.Result.Track.Title), encoder.Extension())

	return &huma.StreamResponse{
		Body: func(ctx huma.Context) {
			ctx.SetHeader("Content-Type", encoder.ContentType())
			ctx.SetHeader("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, filename))
			if !s.cfg.Provider.Hide {
				ctx.SetHeader("X-Source", lyricsResp.Result.Source.ID)
			}
			ctx.SetHeader("X-Sync-Level", lyricsResp.Result.SyncLevel.String())
			if lyricsResp.Cached {
				ctx.SetHeader("X-Cache", "HIT")
			} else {
				ctx.SetHeader("X-Cache", "MISS")
			}
			if lyricsResp.TTL > 0 && fetchMode != "force" {
				ctx.SetHeader("Cache-Control", fmt.Sprintf("public, max-age=%d", int(lyricsResp.TTL.Seconds())))
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

	if s.rate != nil {
		req.Charge = func(ctx context.Context) error {
			return s.rate.Charge(ctx, clientIP(ctx))
		}
	}

	req.Artist, req.Title = normalize.CleanQuery(req.Artist, req.Title)
	return s.orch.Get(ctx, req)
}

func (s *Server) mapError(err error) error {
	if e, ok := errors.AsType[*huma.ErrorModel](err); ok {
		return e
	}

	switch {
	case errors.Is(err, ratelimit.ErrRateLimited):
		e, _ := errors.AsType[*ratelimit.LimitError](err)
		return huma.ErrorWithHeaders(huma.Error429TooManyRequests(e.Error()), http.Header{
			"Retry-After": {strconv.Itoa(int(e.RetryAfter.Seconds()))},
		})
	case errors.Is(err, orchestrator.ErrNoProviders):
		return huma.Error503ServiceUnavailable(err.Error())
	case errors.Is(err, orchestrator.ErrNotFound):
		e := huma.Error404NotFound(err.Error())
		if s.cfg.Cache.TTL.Miss.Duration > 0 {
			return huma.ErrorWithHeaders(e, http.Header{"Cache-Control": {fmt.Sprintf("public, max-age=%d", int(s.cfg.Cache.TTL.Miss.Duration.Seconds()))}})
		}
		return e
	default:
		s.log.Error("provider error", "err", err)
		return huma.Error500InternalServerError("internal error")
	}
}
