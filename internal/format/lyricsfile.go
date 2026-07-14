package format

import (
	"io"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
)

// https://github.com/tranxuanthang/lrcget/blob/main/LYRICSFILE_CONCEPT.md

type lfEncoder struct{}

func (lfEncoder) Levels() (lo, hi lyrics.SyncLevel) {
	return lyrics.SyncNone, lyrics.SyncWord
}
func (lfEncoder) ContentType() string { return "text/yaml; charset=utf-8" }
func (lfEncoder) Extension() string   { return "yaml" }
func (lfEncoder) Desc() string        { return "Used by LRCLIB/LRCGET" }

type lfMeta struct {
	Title      string `yaml:"title"`
	Artist     string `yaml:"artist"`
	Album      string `yaml:"album,omitempty"`
	DurationMs int64  `yaml:"duration_ms,omitempty"`
}

type lfWord struct {
	Text    string `yaml:"text"`
	StartMs int64  `yaml:"start_ms"`
	EndMs   int64  `yaml:"end_ms,omitempty"`
}

type lfLine struct {
	Text    string   `yaml:"text"`
	Words   []lfWord `yaml:"words,omitempty"`
	StartMs int64    `yaml:"start_ms,omitempty"`
	EndMs   int64    `yaml:"end_ms,omitempty"`
}

//nolint:govet // fieldalignment
type lfDoc struct {
	Version  string   `yaml:"version"`
	Metadata lfMeta   `yaml:"metadata"`
	Lines    []lfLine `yaml:"lines,omitempty"`
	Plain    string   `yaml:"plain,omitempty"`
}

func (lfEncoder) Encode(w io.Writer, r *lyrics.Result) error {
	doc := lfDoc{
		Version: "1.0",
		Metadata: lfMeta{
			Title:      r.Track.Title,
			Artist:     r.Track.Artist,
			Album:      r.Track.Album,
			DurationMs: r.Track.Duration * 1000,
		},
	}

	var plainLines []string
	for _, l := range r.Lines {
		plainLines = append(plainLines, l.Text)
	}
	doc.Plain = strings.Join(plainLines, "\n")

	if r.SyncLevel >= lyrics.SyncLine {
		doc.Lines = make([]lfLine, 0, len(r.Lines))
		for _, l := range r.Lines {
			if l.Text == "" {
				continue
			}
			line := lfLine{
				Text:    l.Text,
				StartMs: l.StartMs,
				EndMs:   l.EndMs,
			}
			if r.SyncLevel == lyrics.SyncWord {
				line.Words = make([]lfWord, len(l.Words))
				for i, word := range l.Words {
					line.Words[i] = lfWord{
						Text:    word.Text,
						StartMs: word.StartMs,
						EndMs:   word.EndMs,
					}
				}
			}
			doc.Lines = append(doc.Lines, line)
		}
	}

	return yaml.NewEncoder(w).Encode(doc)
}
