package musixmatch

import (
	"encoding/json"
	"strings"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
)

type richsyncEntry struct {
	StartSec float64 `json:"ts"`
	EndSec   float64 `json:"te"`
	Lines    []struct {
		Content   string  `json:"c"`
		OffsetSec float64 `json:"o"`
	} `json:"l"`
}

func parseRichsync(body string) []lyrics.Line {
	var entries []richsyncEntry
	if err := json.Unmarshal([]byte(body), &entries); err != nil {
		return nil
	}
	lines := make([]lyrics.Line, 0, len(entries))
	for _, e := range entries {
		lineStart := int64(e.StartSec * 1000)
		lineEnd := int64(e.EndSec * 1000)

		words := make([]lyrics.Word, 0, len(e.Lines))
		var parts []string
		for i, w := range e.Lines {
			text := strings.TrimSpace(w.Content)
			if text == "" {
				continue
			}
			wStart := int64((e.StartSec + w.OffsetSec) * 1000)
			var wEnd int64
			if i+1 < len(e.Lines) {
				wEnd = int64((e.StartSec + e.Lines[i+1].OffsetSec) * 1000)
			} else {
				wEnd = lineEnd
			}
			words = append(words, lyrics.Word{StartMs: wStart, EndMs: wEnd, Text: text})
			parts = append(parts, text)
		}

		text := strings.Join(parts, " ")
		if text == "" {
			continue
		}
		lines = append(lines, lyrics.Line{
			StartMs: lineStart,
			EndMs:   lineEnd,
			Text:    text,
			Words:   words,
		})
	}
	return lines
}

type subtitleEntry struct {
	Text string `json:"text"`
	Time struct {
		Total float64 `json:"total"`
	} `json:"time"`
}

func parseSubtitles(body string) []lyrics.Line {
	var entries []subtitleEntry
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

func parseLyrics(body string) []lyrics.Line {
	body = strings.ReplaceAll(body, "\r\n", "\n")
	body = strings.TrimRight(body, "\n")
	if body == "" {
		return nil
	}
	parts := strings.Split(body, "\n")
	lines := make([]lyrics.Line, 0, len(parts))
	for _, p := range parts {
		lines = append(lines, lyrics.Line{Text: p})
	}
	return lines
}
