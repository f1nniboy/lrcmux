package format

import (
	"bufio"
	"fmt"
	"io"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
)

type srtEncoder struct{}

func (srtEncoder) ContentType() string { return "text/plain; charset=utf-8" }
func (srtEncoder) Desc() string        { return "Subtitles for video editors and media players" }

func (srtEncoder) MinLevel() lyrics.SyncLevel { return lyrics.SyncLine }

func (srtEncoder) Encode(w io.Writer, r *lyrics.Result) error {
	bw := bufio.NewWriter(w)
	for i, line := range r.Lines {
		fmt.Fprintf(bw, "%d\n%s --> %s\n%s\n\n", i+1, subStamp(line.StartMs, ','), subStamp(line.EndMs, ','), line.Text)
	}
	return bw.Flush()
}
