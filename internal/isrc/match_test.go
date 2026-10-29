package isrc

import (
	"math"
	"slices"
	"testing"
)

func TestSimilarity(t *testing.T) {
	tests := []struct {
		a, b string
		want float64
	}{
		{"song", "song", 1},
		{"song", "sonx", 0.75},                    // one edit in four
		{"right song", "right songxx", 10.0 / 12}, // two edits in twelve
		{"right song", "right sogn", 0.9},         // a swap is one edit
		{"song", "xxxx", 0},
		{"", "", 0},
	}
	for _, tc := range tests {
		if got := similarity(tc.a, tc.b); math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("similarity(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestDistance(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"song", "song", 0},
		{"song", "sogn", 1}, // adjacent swap
		{"kitten", "sitting", 3},
		{"", "song", 4},
		{"song", "", 4},
		{"abc", "ca", 3}, // a swap can't be edited again
		{"äöü", "äüö", 1},
	}
	for _, tc := range tests {
		if got := distance([]rune(tc.a), []rune(tc.b)); got != tc.want {
			t.Errorf("distance(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestPickBest(t *testing.T) {
	tests := []struct {
		name   string
		want   string
		in     ResolveInput
		tracks []deezerTrack
	}{
		{
			name:   "lone matching candidate",
			tracks: []deezerTrack{{ISRC: "only", Title: "Song", Artist: deezerArtist{Name: "Artist"}}},
			in:     ResolveInput{Title: "Song", Artist: "Artist"},
			want:   "only",
		},
		{
			name: "no similar candidate",
			tracks: []deezerTrack{
				{ISRC: "different", Title: "Completely Different Track", Artist: deezerArtist{Name: "Artist"}},
			},
			in: ResolveInput{Title: "Right Song", Artist: "Artist"},
		},
		{
			name: "exact title and artist wins",
			tracks: []deezerTrack{
				{ISRC: "wrong", Title: "Wrong Song", Artist: deezerArtist{Name: "Artist"}},
				{ISRC: "right", Title: "Right Song", Artist: deezerArtist{Name: "Artist"}},
			},
			in:   ResolveInput{Title: "Right Song", Artist: "Artist"},
			want: "right",
		},
		{
			name: "closer title wins",
			tracks: []deezerTrack{
				{ISRC: "close", Title: "Right Songx", Artist: deezerArtist{Name: "Artist"}}, // distance 1
				{ISRC: "far", Title: "Right Songxxx", Artist: deezerArtist{Name: "Artist"}}, // distance 3
			},
			in:   ResolveInput{Title: "Right Song", Artist: "Artist"},
			want: "close",
		},
		{
			name: "duration tiebreaker",
			tracks: []deezerTrack{
				{ISRC: "far", Title: "Song", Artist: deezerArtist{Name: "Artist"}, Duration: 190},
				{ISRC: "close", Title: "Song", Artist: deezerArtist{Name: "Artist"}, Duration: 182},
			},
			in:   ResolveInput{Title: "Song", Artist: "Artist", Duration: 180},
			want: "close",
		},
		{
			name: "title match wins over duration match on wrong track",
			tracks: []deezerTrack{
				{ISRC: "right", Title: "Right Song", Artist: deezerArtist{Name: "Artist"}, Duration: 180},
				{ISRC: "wrong", Title: "Wrong Song", Artist: deezerArtist{Name: "Artist"}, Duration: 120},
			},
			in:   ResolveInput{Title: "Right Song", Artist: "Artist", Duration: 120},
			want: "right",
		},
		{
			name: "album exact match breaks tie",
			tracks: []deezerTrack{
				{ISRC: "right", Title: "Song", Artist: deezerArtist{Name: "Artist"}, Album: deezerAlbum{Title: "Right Album"}},
				{ISRC: "wrong", Title: "Song", Artist: deezerArtist{Name: "Artist"}, Album: deezerAlbum{Title: "Wrong Album"}},
			},
			in:   ResolveInput{Title: "Song", Artist: "Artist", Album: "Right Album"},
			want: "right",
		},
		{
			name: "album typo still scores higher than wrong album",
			tracks: []deezerTrack{
				{ISRC: "right", Title: "Song", Artist: deezerArtist{Name: "Artist"}, Album: deezerAlbum{Title: "Right Album"}},
				{ISRC: "wrong", Title: "Song", Artist: deezerArtist{Name: "Artist"}, Album: deezerAlbum{Title: "Totally Different"}},
			},
			in:   ResolveInput{Title: "Song", Artist: "Artist", Album: "Rigth Album"}, // typo
			want: "right",
		},
		{
			name: "featured artist in query matches plain track artist",
			tracks: []deezerTrack{
				{ISRC: "right", Title: "Song", Artist: deezerArtist{Name: "Artist"}},
				{ISRC: "wrong", Title: "Song", Artist: deezerArtist{Name: "Somebody Else"}},
			},
			in:   ResolveInput{Title: "Song", Artist: "Artist feat. Someone"},
			want: "right",
		},
		{
			name: "featured artist on track matches plain query artist",
			tracks: []deezerTrack{
				{ISRC: "right", Title: "Song", Artist: deezerArtist{Name: "Artist feat. Someone"}},
				{ISRC: "wrong", Title: "Song", Artist: deezerArtist{Name: "Somebody Else"}},
			},
			in:   ResolveInput{Title: "Song", Artist: "Artist"},
			want: "right",
		},
		{
			name: "exact title and duration with an unrelated artist is not a match",
			tracks: []deezerTrack{
				{ISRC: "wrong", Title: "Song", Artist: deezerArtist{Name: "Somebody Else"}, Duration: 180},
			},
			in: ResolveInput{Title: "Song", Artist: "Artist", Duration: 180},
		},
		{
			name: "same artist, different song is not a match",
			tracks: []deezerTrack{
				{ISRC: "wrong", Title: "Another Title", Artist: deezerArtist{Name: "Artist"}},
			},
			in: ResolveInput{Title: "Song", Artist: "Artist"},
		},
		{
			name: "one edit is too much for a short title",
			tracks: []deezerTrack{
				{ISRC: "wrong", Title: "Sonx", Artist: deezerArtist{Name: "Artist"}},
			},
			in: ResolveInput{Title: "Song", Artist: "Artist"},
		},
		{
			name: "typo in a longer title still matches when the artist agrees",
			tracks: []deezerTrack{
				{ISRC: "right", Title: "Right Songx", Artist: deezerArtist{Name: "Artist"}}, // distance 1
			},
			in:   ResolveInput{Title: "Right Song", Artist: "Artist"},
			want: "right",
		},
		{
			name: "a decoration on one side only still matches",
			tracks: []deezerTrack{
				{ISRC: "right", Title: "Song (From Something)", Artist: deezerArtist{Name: "Artist"}},
			},
			in:   ResolveInput{Title: "Song", Artist: "Artist"},
			want: "right",
		},
		{
			name: "a typo in a decorated title still matches",
			tracks: []deezerTrack{
				{ISRC: "right", Title: "Right Song (From Something)", Artist: deezerArtist{Name: "Artist"}},
			},
			in:   ResolveInput{Title: "Right Sogn", Artist: "Artist"},
			want: "right",
		},
		{
			name: "an instrumental version never stands in for the song",
			tracks: []deezerTrack{
				{ISRC: "instrumental", Title: "Song (Instrumental)", Artist: deezerArtist{Name: "Artist"}, Duration: 180},
				{ISRC: "song", Title: "Song", Artist: deezerArtist{Name: "Artist"}, Duration: 200},
			},
			in:   ResolveInput{Title: "Song", Artist: "Artist", Duration: 180},
			want: "song",
		},
		{
			name: "an instrumental version alone is not a match",
			tracks: []deezerTrack{
				{ISRC: "instrumental", Title: "Song (Karaoke Version)", Artist: deezerArtist{Name: "Artist"}},
			},
			in: ResolveInput{Title: "Song", Artist: "Artist"},
		},
		{
			name: "a sibling with another number is not a match",
			tracks: []deezerTrack{
				{ISRC: "wrong", Title: "Song Part 2", Artist: deezerArtist{Name: "Artist"}},
				{ISRC: "roman", Title: "Song II", Artist: deezerArtist{Name: "Artist"}},
			},
			in: ResolveInput{Title: "Song Part 1", Artist: "Artist"},
		},
		{
			name: "a roman numeral sibling is not a match",
			tracks: []deezerTrack{
				{ISRC: "wrong", Title: "Song II", Artist: deezerArtist{Name: "Artist"}},
			},
			in: ResolveInput{Title: "Song I", Artist: "Artist"},
		},
		{
			name: "a number on one side only is a decoration",
			tracks: []deezerTrack{
				{ISRC: "right", Title: "I Song (Something 1999)", Artist: deezerArtist{Name: "Artist"}},
			},
			in:   ResolveInput{Title: "I Song", Artist: "Artist"},
			want: "right",
		},
		{
			name: "the decorated song beats a near sibling",
			tracks: []deezerTrack{
				{ISRC: "sibling", Title: "Another Titles", Artist: deezerArtist{Name: "Artist"}},
				{ISRC: "song", Title: "Another Title (Something)", Artist: deezerArtist{Name: "Artist"}},
			},
			in:   ResolveInput{Title: "Another Title", Artist: "Artist"},
			want: "song",
		},
		{
			name: "an exact title beats a decorated one",
			tracks: []deezerTrack{
				{ISRC: "decorated", Title: "Song (Something Else)", Artist: deezerArtist{Name: "Artist"}},
				{ISRC: "exact", Title: "Song", Artist: deezerArtist{Name: "Artist"}},
			},
			in:   ResolveInput{Title: "Song", Artist: "Artist"},
			want: "exact",
		},
		{
			name: "a cover by another artist is not a match",
			tracks: []deezerTrack{
				{ISRC: "cover", Title: "Song (Karaoke Version Originally Performed by Artist)", Artist: deezerArtist{Name: "Another Artist"}},
			},
			in: ResolveInput{Title: "Song", Artist: "Artist"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := pickBest(tc.tracks, tc.in, 0)
			if tc.want == "" {
				if ok {
					t.Fatalf("got %q, ok=true, want no match", got.track.ISRC)
				}
				return
			}
			if !ok || got.track.ISRC != tc.want {
				t.Fatalf("got %q, ok=%v, want %q, ok=true", got.track.ISRC, ok, tc.want)
			}
		})
	}
}

func TestReadings(t *testing.T) {
	if got := readings(ResolveInput{Artist: "artist", Title: "song"}); len(got) != 1 {
		t.Errorf("a plain title has %d readings, want 1", len(got))
	}
	got := readings(ResolveInput{Artist: "uploader", Title: "artist - song"})
	want := []ResolveInput{
		{Artist: "uploader", Title: "artist - song"},
		{Artist: "artist", Title: "song"},
		{Artist: "uploader", Title: "artist"},
	}
	if !slices.Equal(got, want) {
		t.Errorf("readings = %+v, want %+v", got, want)
	}
	if got := readings(ResolveInput{Artist: "artist", Title: "- song"}); len(got) != 1 {
		t.Errorf("a dash with nothing before it has %d readings, want 1", len(got))
	}
}
