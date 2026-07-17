package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/meta"
	"github.com/f1nniboy/lrcmux/internal/orchestrator"
)

type KpoeInput struct {
	Artist   string `query:"artist"   doc:"Artist name"`
	Title    string `query:"title"    doc:"Song title"`
	Album    string `query:"album"    doc:"Album name"`
	ISRC     string `query:"isrc"     doc:"ISRC of the track"`
	Duration int64  `query:"duration" doc:"Track duration in seconds"`
}

type KpoeOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         KpoeResponse
}

type KpoeResponse struct {
	KpoeTools string       `json:"KpoeTools"`
	Type      string       `json:"type"`
	Metadata  KpoeMetadata `json:"metadata"`
	Cached    string       `json:"cached"`
	Lyrics    []KpoeLine   `json:"lyrics"`
}

type KpoeMetadata struct {
	Source string `json:"source"`
}

type KpoeLine struct {
	Time     *int64         `json:"time,omitempty"`
	Duration *int64         `json:"duration,omitempty"`
	Text     string         `json:"text"`
	Element  KpoeElement    `json:"element"`
	Syllabus []KpoeSyllabus `json:"syllabus"`
}

type KpoeSyllabus struct {
	Text     string `json:"text"`
	Time     int64  `json:"time"`
	Duration int64  `json:"duration"`
}

type KpoeElement struct {
	Key string `json:"key"`
}

func (s *Server) kpoeOp() huma.Operation {
	return huma.Operation{
		OperationID: "kpoe-get-lyrics",
		Method:      http.MethodGet,
		Path:        "/compat/kpoe/v2/lyrics/get",
		Summary:     "LyricsPlus/KPOE",
		Tags:        []string{"Compatibility"},
	}
}

func (s *Server) handleKpoe(ctx context.Context, input *KpoeInput) (*KpoeOutput, error) {
	result, err := s.fetch(ctx, orchestrator.Request{
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

	synced := result.Result.SyncLevel >= lyrics.SyncLine

	klines := make([]KpoeLine, 0, len(result.Result.Lines))
	for i, l := range result.Result.Lines {
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

	src := fmt.Sprintf("%s at %s", result.Result.Source.Name, meta.AppDomain)

	out := &KpoeOutput{Body: KpoeResponse{
		KpoeTools: meta.AppName,
		Type:      kpoeType(result.Result.SyncLevel),
		Metadata:  KpoeMetadata{Source: src},
		Lyrics:    klines,
		Cached:    "None",
	}}
	if result.TTL > 0 {
		out.CacheControl = fmt.Sprintf("public, max-age=%d", int(result.TTL.Seconds()))
	}
	return out, nil
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
