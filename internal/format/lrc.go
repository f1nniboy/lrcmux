package format

import (
	"bufio"
	"fmt"
	"io"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
)

type lrcEncoder struct{ anyLevel }

func (lrcEncoder) ContentType() string { return "text/plain; charset=utf-8" }
func (lrcEncoder) Desc() string        { return "Standard .lrc files for music players" }

func (lrcEncoder) Encode(w io.Writer, r *lyrics.Result) error {
	bw := bufio.NewWriter(w)
	switch r.SyncLevel {
	case lyrics.SyncNone:
		for _, line := range r.Lines {
			fmt.Fprintln(bw, line.Text)
		}
	case lyrics.SyncLine:
		for _, line := range r.Lines {
			fmt.Fprintf(bw, "[%s] %s\n", formatStamp(line.StartMs), line.Text)
		}
	case lyrics.SyncWord:
		for _, line := range r.Lines {
			fmt.Fprintf(bw, "[%s]", formatStamp(line.StartMs))
			if len(line.Words) == 0 {
				fmt.Fprintf(bw, " %s\n", line.Text)
				continue
			}
			for _, word := range line.Words {
				fmt.Fprintf(bw, "<%s>%s ", formatStamp(word.StartMs), word.Text)
			}
			fmt.Fprintln(bw)
		}
	}
	return bw.Flush()
}
