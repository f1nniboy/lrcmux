package format

import (
	"bufio"
	"fmt"
	"io"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
)

type txtEncoder struct{}

func (txtEncoder) Levels() (min, max lyrics.SyncLevel) { return lyrics.SyncNone, lyrics.SyncNone }
func (txtEncoder) ContentType() string                 { return "text/plain; charset=utf-8" }
func (txtEncoder) Desc() string                        { return "Plain unsynced text" }

func (txtEncoder) Encode(w io.Writer, r *lyrics.Result) error {
	bw := bufio.NewWriter(w)
	for _, line := range r.Lines {
		fmt.Fprintln(bw, line.Text)
	}
	return bw.Flush()
}
