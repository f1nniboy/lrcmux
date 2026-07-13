package format

import (
	"bufio"
	"fmt"
	"io"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
)

type lrcEncoder struct{}

func (lrcEncoder) Levels() (lo, hi lyrics.SyncLevel) { return lyrics.SyncNone, lyrics.SyncWord }
func (lrcEncoder) ContentType() string               { return "text/plain; charset=utf-8" }
func (lrcEncoder) Extension() string                 { return "lrc" }
func (lrcEncoder) Desc() string                      { return "Standard .lrc files for music players" }

func (lrcEncoder) Encode(w io.Writer, r *lyrics.Result) error {
	bw := bufio.NewWriter(w)

	switch r.SyncLevel {
	case lyrics.SyncWord:
		for _, line := range r.Lines {
			fmt.Fprintf(bw, "[%s]", formatStamp(line.StartMs))
			if len(line.Words) == 0 {
				fmt.Fprintln(bw)
				continue
			}
			for _, word := range line.Words {
				fmt.Fprintf(bw, "<%s>%s", formatStamp(word.StartMs), word.Text)
			}
			fmt.Fprintln(bw)
		}
	case lyrics.SyncNone:
		for _, line := range r.Lines {
			fmt.Fprintln(bw, line.Text)
		}
	case lyrics.SyncLine:
		for _, line := range r.Lines {
			if line.Text != "" {
				fmt.Fprintf(bw, "[%s] %s\n", formatStamp(line.StartMs), line.Text)
			} else {
				fmt.Fprintf(bw, "[%s]\n", formatStamp(line.StartMs))
			}
		}
	default:
	}
	return bw.Flush()
}
