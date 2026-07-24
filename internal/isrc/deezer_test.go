package isrc

import (
	"testing"
)

func TestDistScore(t *testing.T) {
	tests := []struct {
		a, b string
		n    int
		want float64
	}{
		{"song", "song", 5, 5},     // exact
		{"song", "sonx", 5, 4},     // distance 1
		{"song", "sxxx", 5, 2},     // distance 3
		{"song", "xxxxxxxx", 5, 0}, // clamped at 0
		{"", "", 3, 3},             // both empty
	}
	for _, tc := range tests {
		got := distScore(tc.a, tc.b, tc.n)
		if got != tc.want {
			t.Errorf("distScore(%q, %q, %d) = %v, want %v", tc.a, tc.b, tc.n, got, tc.want)
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
			name:   "single track returned immediately",
			tracks: []deezerTrack{{ISRC: "only"}},
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := pickBest(tc.tracks, tc.in)
			if tc.want == "" {
				if ok {
					t.Fatalf("got %q, ok=true, want no match", got.ISRC)
				}
				return
			}
			if !ok || got.ISRC != tc.want {
				t.Fatalf("got %q, ok=%v, want %q, ok=true", got.ISRC, ok, tc.want)
			}
		})
	}
}
