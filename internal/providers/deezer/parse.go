package deezer

import (
	"strings"
	"time"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
)

// a correct sync can end slightly past the track
const durationSlack = 2 * time.Second

type lyricsData struct {
	Text      string       `json:"text"`
	WordLines []wordLine   `json:"synchronizedWordByWordLines"`
	Lines     []syncedLine `json:"synchronizedLines"`
}

type wordLine struct {
	Words []word `json:"words"`
}

type word struct {
	Word  string `json:"word"`
	Start int64  `json:"start"`
	End   int64  `json:"end"`
}

type syncedLine struct {
	Line         string `json:"line"`
	Milliseconds int64  `json:"milliseconds"`
	Duration     int64  `json:"duration"`
}

// a sync running past the track drops a level, since a deezer result often ends the search
func parse(l lyricsData, duration int64) (*lyrics.Result, error) {
	limit := (time.Duration(duration)*time.Second + durationSlack).Milliseconds()
	fits := func(lines []lyrics.Line) bool {
		return len(lines) > 0 && lines[len(lines)-1].EndMs <= limit
	}
	if lines := wordSynced(l.WordLines); fits(lines) {
		return &lyrics.Result{Lines: lines, SyncLevel: lyrics.SyncWord}, nil
	}
	if lines := lineSynced(l.Lines); fits(lines) {
		return &lyrics.Result{Lines: lines, SyncLevel: lyrics.SyncLine}, nil
	}
	if lines := lyrics.ParsePlain(l.Text); len(lines) > 0 {
		return &lyrics.Result{Lines: lines, SyncLevel: lyrics.SyncNone}, nil
	}
	return nil, lyrics.ErrNotFound
}

// a word holds no whitespace, so all but a line's last get a trailing space and
// the words join back into the line
func wordSynced(in []wordLine) []lyrics.Line {
	out := make([]lyrics.Line, 0, len(in))
	for _, wl := range in {
		words := make([]lyrics.Word, 0, len(wl.Words))
		for _, w := range wl.Words {
			if t := strings.TrimSpace(w.Word); t != "" {
				words = append(words, lyrics.Word{Text: t, StartMs: w.Start, EndMs: w.End})
			}
		}
		if len(words) == 0 {
			continue
		}
		var text strings.Builder
		for i := range words {
			if i < len(words)-1 {
				words[i].Text += " "
			}
			text.WriteString(words[i].Text)
		}
		out = append(out, lyrics.Line{
			Text:    text.String(),
			Words:   words,
			StartMs: words[0].StartMs,
			EndMs:   words[len(words)-1].EndMs,
		})
	}
	return out
}

func lineSynced(in []syncedLine) []lyrics.Line {
	out := make([]lyrics.Line, 0, len(in))
	for _, l := range in {
		if t := strings.TrimSpace(l.Line); t != "" {
			out = append(out, lyrics.Line{Text: t, StartMs: l.Milliseconds, EndMs: l.Milliseconds + l.Duration})
		}
	}
	return out
}
