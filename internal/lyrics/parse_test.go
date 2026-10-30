package lyrics

import (
	"reflect"
	"testing"
)

func TestParseLRC(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		want  []Line
		level SyncLevel
	}{
		{
			name:  "line sync",
			in:    "[00:01.00]First line\n[00:04.00]\n[00:06.50]Second line",
			level: SyncLine,
			want:  []Line{{StartMs: 1000, EndMs: 4000, Text: "First line"}, {StartMs: 4000, EndMs: 6500}, {StartMs: 6500, EndMs: 11500, Text: "Second line"}},
		},
		{
			name:  "a line with several stamps",
			in:    "[00:01.00][00:05.00]Chorus\n[00:03.00]Verse",
			level: SyncLine,
			want:  []Line{{StartMs: 1000, EndMs: 3000, Text: "Chorus"}, {StartMs: 3000, EndMs: 5000, Text: "Verse"}, {StartMs: 5000, EndMs: 10000, Text: "Chorus"}},
		},
		{
			name:  "word sync",
			in:    "[00:00.00]<00:00.20>First <00:01.00>line <00:02.00>\n[00:05.00]<00:05.00>Next",
			level: SyncWord,
			want: []Line{
				{EndMs: 5000, Text: "First line", Words: []Word{{StartMs: 200, EndMs: 1000, Text: "First "}, {StartMs: 1000, EndMs: 5000, Text: "line"}}},
				{StartMs: 5000, EndMs: 10000, Text: "Next", Words: []Word{{StartMs: 5000, EndMs: 10000, Text: "Next"}}},
			},
		},
		{
			name:  "metadata and malformed stamps",
			in:    "[ar:Artist]\n[invalid]Bad\n[00:01.00 Bad\n[00:ab.00]Bad\n[00:01.xy]Bad\n[00:02.00]Good",
			level: SyncLine,
			want:  []Line{{StartMs: 2000, EndMs: 7000, Text: "Good"}},
		},
		{
			name:  "malformed word stamps",
			in:    "[00:01.00]<bad>word\n[00:02.00]<notclosed",
			level: SyncLine,
			want:  []Line{{StartMs: 1000, EndMs: 2000, Text: "<bad>word"}, {StartMs: 2000, EndMs: 7000, Text: "<notclosed"}},
		},
		{
			name:  "fractions",
			in:    "[01:23.45]First\n[01:23.456]Second\n[01:24]Third",
			level: SyncLine,
			want:  []Line{{StartMs: 83450, EndMs: 83456, Text: "First"}, {StartMs: 83456, EndMs: 84000, Text: "Second"}, {StartMs: 84000, EndMs: 89000, Text: "Third"}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lines, level := ParseLRC(tc.in)
			if level != tc.level || !reflect.DeepEqual(lines, tc.want) {
				t.Errorf("got %s %+v\nwant %s %+v", level, lines, tc.level, tc.want)
			}
		})
	}
}
