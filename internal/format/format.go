package format

import (
	"fmt"
	"io"
	"sort"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
)

type Encoder interface {
	Encode(w io.Writer, r *lyrics.Result) error
	ContentType() string
	MinLevel() lyrics.SyncLevel
	Desc() string
}

type anyLevel struct{}

func (anyLevel) MinLevel() lyrics.SyncLevel { return lyrics.SyncNone }

var registry = map[string]Encoder{
	"lrc":  lrcEncoder{},
	"json": jsonEncoder{},
	"txt":  txtEncoder{},
	"srt":  srtEncoder{},
	"vtt":  vttEncoder{},
}

func All() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func Get(name string) (Encoder, error) {
	if name == "" {
		name = "lrc"
	}
	e, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown format %q", name)
	}
	return e, nil
}

func formatStamp(ms int64) string {
	if ms < 0 {
		ms = 0
	}
	mm := ms / 60_000
	ss := (ms % 60_000) / 1000
	cs := (ms % 1000) / 10
	return fmt.Sprintf("%02d:%02d.%02d", mm, ss, cs)
}

// subStamp formats ms as HH:MM:SS<sep>mmm, used by SRT (sep=',') and VTT (sep='.')
func subStamp(ms int64, sep byte) string {
	if ms < 0 {
		ms = 0
	}
	h := ms / 3_600_000
	m := (ms % 3_600_000) / 60_000
	s := (ms % 60_000) / 1_000
	millis := ms % 1_000
	return fmt.Sprintf("%02d:%02d:%02d%c%03d", h, m, s, sep, millis)
}
