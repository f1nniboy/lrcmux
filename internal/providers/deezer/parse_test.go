package deezer

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
)

const (
	wordLines = `"synchronizedWordByWordLines":[` +
		`{"words":[{"start":1000,"end":1500,"word":"First"},{"start":1500,"end":2000,"word":"line"}]},` +
		`{"words":[{"start":3000,"end":4000,"word":"Second"}]}]`
	syncedLines = `"synchronizedLines":[{"line":"First line","milliseconds":1100,"duration":900},{"line":"Second","milliseconds":3000,"duration":1000}]`
	plainText   = `"text":"First line\r\nSecond\r\n\r\nThird"`
)

var (
	wordWant = []lyrics.Line{
		{Text: "First line", StartMs: 1000, EndMs: 2000, Words: []lyrics.Word{
			{Text: "First ", StartMs: 1000, EndMs: 1500},
			{Text: "line", StartMs: 1500, EndMs: 2000},
		}},
		{Text: "Second", StartMs: 3000, EndMs: 4000, Words: []lyrics.Word{{Text: "Second", StartMs: 3000, EndMs: 4000}}},
	}
	lineWant  = []lyrics.Line{{Text: "First line", StartMs: 1100, EndMs: 2000}, {Text: "Second", StartMs: 3000, EndMs: 4000}}
	plainWant = []lyrics.Line{{Text: "First line"}, {Text: "Second"}, {Text: ""}, {Text: "Third"}}
)

func TestParse(t *testing.T) {
	// with durationSlack, a duration of 2 s allows syncs to end at 4000 ms
	cases := []struct {
		err      error
		name     string
		body     string
		want     []lyrics.Line
		duration int64
		level    lyrics.SyncLevel
	}{
		{
			name:     "words join back into their line",
			body:     wordLines + "," + syncedLines + "," + plainText,
			want:     wordWant,
			duration: 200,
			level:    lyrics.SyncWord,
		},
		{
			name:     "line sync without word lines",
			body:     `"synchronizedWordByWordLines":null,` + syncedLines + "," + plainText,
			want:     lineWant,
			duration: 200,
			level:    lyrics.SyncLine,
		},
		{
			name:     "plain from text",
			body:     `"synchronizedWordByWordLines":[],"synchronizedLines":null,` + plainText,
			want:     plainWant,
			duration: 200,
			level:    lyrics.SyncNone,
		},
		{
			name:     "words past the limit fall back to lines",
			body:     `"synchronizedWordByWordLines":[{"words":[{"start":1000,"end":4001,"word":"First"}]}],` + syncedLines,
			want:     lineWant,
			duration: 2,
			level:    lyrics.SyncLine,
		},
		{
			name:     "lines past the limit fall back to plain",
			body:     syncedLines + "," + plainText,
			want:     plainWant,
			duration: 1,
			level:    lyrics.SyncNone,
		},
		{
			name: "blank words and word lines are skipped",
			body: `"synchronizedWordByWordLines":[{"words":[]},{"words":[{"start":0,"end":0,"word":""}]},` +
				`{"words":[{"start":1000,"end":1500,"word":"First"},{"start":1500,"end":1500,"word":" "},{"start":1500,"end":2000,"word":"line"}]},` +
				`{"words":[{"start":3000,"end":4000,"word":"Second"}]}]`,
			want:     wordWant,
			duration: 200,
			level:    lyrics.SyncWord,
		},
		{
			name: "nothing left after the blanks",
			body: `"text":" \r\n ","synchronizedWordByWordLines":[{"words":[{"start":0,"end":0,"word":""}]}],` +
				`"synchronizedLines":[{"line":"","milliseconds":0,"duration":0}]`,
			duration: 200,
			err:      lyrics.ErrNotFound,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var l lyricsData
			if err := json.Unmarshal([]byte("{"+c.body+"}"), &l); err != nil {
				t.Fatal(err)
			}
			res, err := parse(l, c.duration)
			if c.err != nil || err != nil {
				if !errors.Is(err, c.err) {
					t.Errorf("err = %v, want %v", err, c.err)
				}
				return
			}
			if res.SyncLevel != c.level || !reflect.DeepEqual(res.Lines, c.want) {
				t.Errorf("got %s %+v, want %s %+v", res.SyncLevel, res.Lines, c.level, c.want)
			}
		})
	}
}
