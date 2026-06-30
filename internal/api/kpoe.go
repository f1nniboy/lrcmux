package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/meta"
	"github.com/f1nniboy/lrcmux/internal/orchestrator"
)

type KpoeInput struct {
	Artist   string `query:"artist"   doc:"Artist name"`
	Title    string `query:"title"    doc:"Song title"`
	Album    string `query:"album"    doc:"Album name"`
	Duration int64  `query:"duration" doc:"Track duration in seconds"`
	ISRC     string `query:"isrc"     doc:"ISRC of the track"`
}

type KpoeOutput struct {
	Body KpoeResponse
}

type KpoeResponse struct {
	KpoeTools      string       `json:"KpoeTools"`
	Type           string       `json:"type"`
	Metadata       KpoeMetadata `json:"metadata"`
	Lyrics         []KpoeLine   `json:"lyrics"`
	Cached         string       `json:"cached"`
	ProcessingTime KpoeTime     `json:"processingTime"`
}

type KpoeMetadata struct {
	Source string `json:"source"`
}

type KpoeLine struct {
	Time     *int64         `json:"time,omitempty"`
	Duration *int64         `json:"duration,omitempty"`
	Text     string         `json:"text"`
	Syllabus []KpoeSyllabus `json:"syllabus"`
	Element  KpoeElement    `json:"element"`
}

type KpoeSyllabus struct {
	Time     int64  `json:"time"`
	Duration int64  `json:"duration"`
	Text     string `json:"text"`
}

type KpoeElement struct {
	Key string `json:"key"`
}

type KpoeTime struct {
	TimeElapsed int64 `json:"timeElapsed"`
}

func (s *Server) kpoeOp() huma.Operation {
	return huma.Operation{
		OperationID: "kpoe-get-lyrics",
		Method:      http.MethodGet,
		Path:        "/api/compat/kpoe/v2/lyrics/get",
		Summary:     "LyricsPlus/KPOE",
		Description: "Drop-in replacement for apps that use the LyricsPlus/KPOE API.",
		Tags:        []string{"Compatibility"},
	}
}

func (s *Server) handleKpoe(ctx context.Context, input *KpoeInput) (*KpoeOutput, error) {
	start := time.Now()

	resp, err := s.fetch(ctx, orchestrator.Request{
		Artist:   input.Artist,
		Title:    input.Title,
		Album:    input.Album,
		Duration: input.Duration,
		ISRC:     input.ISRC,
		Level:    lyrics.SyncWord,
	})
	if err != nil {
		return nil, s.mapError(err)
	}

	elapsed := time.Since(start).Milliseconds()
	synced := resp.Result.SyncLevel >= lyrics.SyncLine

	klines := make([]KpoeLine, 0, len(resp.Result.Lines))
	for i, l := range resp.Result.Lines {
		kl := KpoeLine{
			Text:     l.Text,
			Syllabus: []KpoeSyllabus{},
			Element:  KpoeElement{Key: fmt.Sprintf("L%d", i+1)},
		}
		if synced {
			startMs := l.StartMs
			dur := max(l.EndMs-l.StartMs, 0)
			kl.Time = &startMs
			kl.Duration = &dur
		}
		if len(l.Words) > 0 {
			kl.Syllabus = make([]KpoeSyllabus, len(l.Words))
			for j, w := range l.Words {
				wd := max(w.EndMs-w.StartMs, 0)
				kl.Syllabus[j] = KpoeSyllabus{
					Time:     w.StartMs,
					Duration: wd,
					Text:     w.Text,
				}
			}
		}
		klines = append(klines, kl)
	}

	cached := "None"
	if resp.Cached {
		cached = "Database"
	}

	src := meta.AppDomain
	if !s.hide && resp.Result.Source.ID != "" {
		src = fmt.Sprintf("%s at %s", resp.Result.Source.Name, meta.AppDomain)
	}

	return &KpoeOutput{Body: KpoeResponse{
		KpoeTools:      meta.AppName,
		Type:           kpoeType(resp.Result.SyncLevel),
		Metadata:       KpoeMetadata{Source: src},
		Lyrics:         klines,
		Cached:         cached,
		ProcessingTime: KpoeTime{TimeElapsed: elapsed},
	}}, nil
}

func kpoeType(level lyrics.SyncLevel) string {
	switch level {
	case lyrics.SyncWord:
		return "Word"
	case lyrics.SyncLine:
		return "Line"
	default:
		return "None"
	}
}
