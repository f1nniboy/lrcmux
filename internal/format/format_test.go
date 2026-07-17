package format

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
)

var track = lyrics.Track{
	Artist:   "Test Artist",
	Title:    "Test Song",
	Album:    "Test Album",
	Duration: 180,
	ISRC:     "ABCD12345678",
	Cover: lyrics.TrackCover{
		Big:    "https://example.com",
		Medium: "https://example.com",
		Small:  "https://example.com",
	},
}

var source = lyrics.Source{ID: "test", Name: "Test", URL: "https://example.com"}

var result = &lyrics.Result{
	SyncLevel: lyrics.SyncWord,
	Track:     track,
	Source:    source,
	Lines: []lyrics.Line{
		{
			StartMs: 0, EndMs: 3000, Text: "Hello world",
			Words: []lyrics.Word{
				{StartMs: 0, EndMs: 2000, Text: "Hello "},
				{StartMs: 2000, EndMs: 3000, Text: "world"},
			},
		},
		{StartMs: 3000, EndMs: 5000, Text: ""},
		{
			StartMs: 5000, EndMs: 7000, Text: "Goodbye",
			Words: []lyrics.Word{
				{StartMs: 5000, EndMs: 7000, Text: "Goodbye"},
			},
		},
	},
}

var update = flag.Bool("update", false, "rewrite golden files")

func TestEncoders(t *testing.T) {
	for key, enc := range registry {
		lo, hi := enc.Levels()
		for _, level := range lyrics.Levels {
			if level < lo || level > hi {
				continue
			}
			level := level
			t.Run(key+"/"+level.String(), func(t *testing.T) {
				var buf bytes.Buffer
				if err := enc.Encode(&buf, result.Downgrade(level)); err != nil {
					t.Fatalf("encode: %v", err)
				}
				path := filepath.Join("test", level.String()+"."+key)
				if *update {
					if err := os.MkdirAll("test", 0755); err != nil {
						t.Fatalf("mkdir: %v", err)
					}
					if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
						t.Fatalf("write golden: %v", err)
					}
					return
				}
				want, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("missing golden file %s (run with -update to create)", path)
				}
				if !bytes.Equal(buf.Bytes(), want) {
					edits := myers.ComputeEdits(span.URIFromPath(path), string(want), buf.String())
					t.Errorf("output mismatch:\n%s", gotextdiff.ToUnified("want", "got", string(want), edits))
				}
			})
		}
	}
}
