package format

import (
	"encoding/json"
	"io"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
)

type JSONMeta struct {
	Source *lyrics.Source `json:"source,omitempty" doc:"Provider that returned the result"`
	Level  string         `json:"level" doc:"Sync level of the returned lyrics" enum:"word,line,none"`
}

type JSONResponse struct {
	Track lyrics.Track  `json:"track"`
	Meta  JSONMeta      `json:"meta"`
	Lines []lyrics.Line `json:"lines"`
}

type jsonEncoder struct{}

func (jsonEncoder) Levels() (lo, hi lyrics.SyncLevel) { return lyrics.SyncNone, lyrics.SyncWord }
func (jsonEncoder) ContentType() string               { return "application/json; charset=utf-8" }
func (jsonEncoder) Extension() string                 { return "json" }
func (jsonEncoder) Desc() string                      { return "Default, structured lines and metadata" }

func (jsonEncoder) Encode(w io.Writer, r *lyrics.Result) error {
	out := JSONResponse{
		Meta:  JSONMeta{Level: r.SyncLevel.String()},
		Lines: r.Lines,
		Track: r.Track,
	}
	if r.Source.ID != "" {
		out.Meta.Source = &r.Source
	}

	lines := make([]lyrics.Line, 0, len(r.Lines))
	for _, l := range r.Lines {
		if r.SyncLevel != lyrics.SyncNone && l.Text == "" {
			continue
		}
		line := lyrics.Line{Text: l.Text}
		if r.SyncLevel >= lyrics.SyncLine {
			line.StartMs = l.StartMs
			line.EndMs = l.EndMs
		}
		if r.SyncLevel >= lyrics.SyncWord && len(l.Words) > 0 {
			line.Words = make([]lyrics.Word, len(l.Words))
			copy(line.Words, l.Words)
		}
		lines = append(lines, line)
	}
	out.Lines = lines
	enc := json.NewEncoder(w)
	return enc.Encode(out)
}
