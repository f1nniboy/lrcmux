package musixmatch

import (
	"encoding/json"
	"strings"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
)

func parseRichsync(body string) []lyrics.Line {
	var entries []struct {
		Lines []struct {
			Content   string  `json:"c"`
			OffsetSec float64 `json:"o"`
		} `json:"l"`
		StartSec float64 `json:"ts"`
		EndSec   float64 `json:"te"`
	}

	if err := json.Unmarshal([]byte(body), &entries); err != nil {
		return nil
	}
	lines := make([]lyrics.Line, 0, len(entries))
	for _, e := range entries {
		lineStart := int64(e.StartSec * 1000)
		lineEnd := int64(e.EndSec * 1000)

		words := make([]lyrics.Word, 0, len(e.Lines))
		for _, w := range e.Lines {
			wStart := int64((e.StartSec + w.OffsetSec) * 1000)
			if len(words) > 0 {
				words[len(words)-1].EndMs = wStart
			}
			words = append(words, lyrics.Word{StartMs: wStart, EndMs: lineEnd, Text: w.Content})
		}

		if len(words) == 0 {
			continue
		}

		var b strings.Builder
		for _, w := range words {
			b.WriteString(w.Text)
		}
		lines = append(lines, lyrics.Line{
			StartMs: lineStart,
			EndMs:   lineEnd,
			Text:    b.String(),
			Words:   words,
		})
	}
	return lines
}

func parseSubtitles(body string) []lyrics.Line {
	var entries []struct {
		Text string `json:"text"`
		Time struct {
			Total float64 `json:"total"`
		} `json:"time"`
	}

	if err := json.Unmarshal([]byte(body), &entries); err != nil {
		return nil
	}
	lines := make([]lyrics.Line, 0, len(entries))
	for _, e := range entries {
		lines = append(lines, lyrics.Line{
			StartMs: int64(e.Time.Total * 1000),
			Text:    e.Text,
		})
	}
	for i := 0; i < len(lines)-1; i++ {
		lines[i].EndMs = lines[i+1].StartMs
	}
	if len(lines) > 0 {
		last := &lines[len(lines)-1]
		last.EndMs = last.StartMs + 5000
	}
	return lines
}
