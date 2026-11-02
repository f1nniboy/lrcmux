package lyrics

import (
	"reflect"
	"testing"
)

func TestClean(t *testing.T) {
	tests := []struct {
		name string
		want []Line
		in   Result
		ok   bool
	}{
		{
			name: "a synced break becomes one empty line",
			in: Result{SyncLevel: SyncLine, Lines: []Line{
				{StartMs: 0, EndMs: 1000, Text: "♪"},
				{StartMs: 1000, EndMs: 2000, Text: "First line"},
				{StartMs: 2000, EndMs: 9000, Text: "♪ ♪"},
				{StartMs: 9000, EndMs: 10000, Text: ""},
				{StartMs: 10000, EndMs: 12000, Text: "Second line"},
				{StartMs: 12000, EndMs: 15000, Text: ""},
			}},
			want: []Line{
				{StartMs: 1000, EndMs: 2000, Text: "First line"},
				{StartMs: 2000, EndMs: 10000},
				{StartMs: 10000, EndMs: 12000, Text: "Second line"},
			},
			ok: true,
		},
		{
			name: "plain text keeps one blank line between stanzas",
			in:   Result{Lines: []Line{{}, {Text: "First line"}, {Text: " "}, {Text: "🎵"}, {}, {Text: "Second line"}, {}}},
			want: []Line{{Text: "First line"}, {}, {Text: "Second line"}},
			ok:   true,
		},
		{
			name: "section markers go, ad-libs stay",
			in:   Result{Lines: []Line{{Text: "[Chorus]"}, {Text: "First line"}, {Text: "(Verse 2: Artist)"}, {Text: "(yeah)"}}},
			want: []Line{{Text: "First line"}, {Text: "(yeah)"}},
			ok:   true,
		},
		{
			name: "words follow their line",
			in:   Result{SyncLevel: SyncWord, Lines: []Line{{Text: "li\u0435 song", Words: []Word{{Text: "li"}, {Text: "\u0435 "}, {Text: "song"}}}}},
			want: []Line{{Text: "lie song", Words: []Word{{Text: "li"}, {Text: "e "}, {Text: "song"}}}},
			ok:   true,
		},
		{
			name: "only breaks is no lyrics",
			in:   Result{SyncLevel: SyncLine, Lines: []Line{{Text: "♪"}, {}}},
			want: []Line{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.in
			if ok := r.Clean(); ok != tc.ok || !reflect.DeepEqual(r.Lines, tc.want) {
				t.Errorf("got %v %+v, want %v %+v", ok, r.Lines, tc.ok, tc.want)
			}
		})
	}
}

func TestNormalizeText(t *testing.T) {
	tests := []struct{ name, in, want string }{
		{"look-alike letters", "I could li\u0435, th\u0435 song", "I could lie, the song"},
		{"real cyrillic stays", "\u043f\u0440\u0438\u0432\u0435\u0442 song", "\u043f\u0440\u0438\u0432\u0435\u0442 song"},
		{"a word of only look-alikes stays", "\u039f\u03b4\u03cc\u03c2 and \u0441\u043e\u0440\u043e\u043a", "\u039f\u03b4\u03cc\u03c2 and \u0441\u043e\u0440\u043e\u043a"},
		{"cjk space stays", "First\u3000line", "First\u3000line"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got, _ := normalizeText(tc.in, false); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
