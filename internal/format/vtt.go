package format

import (
	"bufio"
	"fmt"
	"io"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
)

type vttEncoder struct{}

func (vttEncoder) Levels() (min, max lyrics.SyncLevel) { return lyrics.SyncLine, lyrics.SyncLine }
func (vttEncoder) ContentType() string                 { return "text/vtt; charset=utf-8" }
func (vttEncoder) Desc() string                        { return "WebVTT for browser-based players" }

func (vttEncoder) Encode(w io.Writer, r *lyrics.Result) error {
	bw := bufio.NewWriter(w)
	fmt.Fprint(bw, "WEBVTT\n\n")
	for _, line := range r.Lines {
		fmt.Fprintf(bw, "%s --> %s\n%s\n\n", subStamp(line.StartMs, '.'), subStamp(line.EndMs, '.'), line.Text)
	}
	return bw.Flush()
}
