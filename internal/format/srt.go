package format

import (
	"bufio"
	"fmt"
	"io"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
)

type srtEncoder struct{}

func (srtEncoder) Levels() (lo, hi lyrics.SyncLevel) { return lyrics.SyncLine, lyrics.SyncLine }
func (srtEncoder) ContentType() string               { return "text/plain; charset=utf-8" }
func (srtEncoder) Extension() string                 { return "srt" }
func (srtEncoder) Desc() string                      { return "Subtitles for video editors and media players" }

func (srtEncoder) Encode(w io.Writer, r *lyrics.Result) error {
	bw := bufio.NewWriter(w)
	n := 0
	for _, line := range r.Lines {
		if line.Text == "" {
			continue
		}
		n++
		fmt.Fprintf(bw, "%d\n%s --> %s\n%s\n\n", n, subStamp(line.StartMs, ','), subStamp(line.EndMs, ','), line.Text)
	}
	return bw.Flush()
}
