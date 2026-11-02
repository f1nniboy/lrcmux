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
func (lrcEncoder) SupportsInstrumental() bool        { return false }

func (lrcEncoder) Encode(w io.Writer, r *lyrics.Result) error {
	bw := bufio.NewWriter(w)

	if r.SyncLevel == lyrics.SyncNone {
		for _, line := range r.Lines {
			fmt.Fprintln(bw, line.Text)
		}
		return bw.Flush()
	}
	for _, line := range r.Lines {
		fmt.Fprintf(bw, "[%s]", formatStamp(line.StartMs))
		switch {
		case r.SyncLevel == lyrics.SyncWord && len(line.Words) > 0:
			for _, word := range line.Words {
				fmt.Fprintf(bw, "<%s>%s", formatStamp(word.StartMs), word.Text)
			}
		case line.Text != "":
			fmt.Fprintf(bw, " %s", line.Text)
		default:
		}
		fmt.Fprintln(bw)
	}
	return bw.Flush()
}
