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
	t.Run("single track returned immediately", func(t *testing.T) {
		tracks := []deezerTrack{{ISRC: "only"}}
		got := pickBest(tracks, ResolveInput{})
		if got.ISRC != "only" {
			t.Fatalf("got %q, want %q", got.ISRC, "only")
		}
	})

	t.Run("exact title and artist wins", func(t *testing.T) {
		tracks := []deezerTrack{
			{ISRC: "wrong", Title: "Wrong Song", Artist: deezerArtist{Name: "Artist"}},
			{ISRC: "right", Title: "Right Song", Artist: deezerArtist{Name: "Artist"}},
		}
		got := pickBest(tracks, ResolveInput{Title: "Right Song", Artist: "Artist"})
		if got.ISRC != "right" {
			t.Fatalf("got %q, want %q", got.ISRC, "right")
		}
	})

	t.Run("closer title wins", func(t *testing.T) {
		tracks := []deezerTrack{
			{ISRC: "close", Title: "Right Songx", Artist: deezerArtist{Name: "Artist"}}, // distance 1
			{ISRC: "far", Title: "Right Songxxx", Artist: deezerArtist{Name: "Artist"}}, // distance 3
		}
		got := pickBest(tracks, ResolveInput{Title: "Right Song", Artist: "Artist"})
		if got.ISRC != "close" {
			t.Fatalf("got %q, want %q", got.ISRC, "close")
		}
	})

	t.Run("duration tiebreaker", func(t *testing.T) {
		tracks := []deezerTrack{
			{ISRC: "far", Title: "Song", Artist: deezerArtist{Name: "Artist"}, Duration: 190},
			{ISRC: "close", Title: "Song", Artist: deezerArtist{Name: "Artist"}, Duration: 182},
		}
		got := pickBest(tracks, ResolveInput{Title: "Song", Artist: "Artist", Duration: 180})
		if got.ISRC != "close" {
			t.Fatalf("got %q, want %q", got.ISRC, "close")
		}
	})

	t.Run("album exact match breaks tie", func(t *testing.T) {
		tracks := []deezerTrack{
			{ISRC: "right", Title: "Song", Artist: deezerArtist{Name: "Artist"}, Album: deezerAlbum{Title: "Right Album"}},
			{ISRC: "wrong", Title: "Song", Artist: deezerArtist{Name: "Artist"}, Album: deezerAlbum{Title: "Wrong Album"}},
		}
		got := pickBest(tracks, ResolveInput{Title: "Song", Artist: "Artist", Album: "Right Album"})
		if got.ISRC != "right" {
			t.Fatalf("got %q, want %q", got.ISRC, "right")
		}
	})

	t.Run("album typo still scores higher than wrong album", func(t *testing.T) {
		tracks := []deezerTrack{
			{ISRC: "right", Title: "Song", Artist: deezerArtist{Name: "Artist"}, Album: deezerAlbum{Title: "Right Album"}},
			{ISRC: "wrong", Title: "Song", Artist: deezerArtist{Name: "Artist"}, Album: deezerAlbum{Title: "Totally Different"}},
		}
		got := pickBest(tracks, ResolveInput{Title: "Song", Artist: "Artist", Album: "Rigth Album"}) // typo
		if got.ISRC != "right" {
			t.Fatalf("got %q, want %q", got.ISRC, "right")
		}
	})

	t.Run("featured artist in query matches plain track artist", func(t *testing.T) {
		tracks := []deezerTrack{
			{ISRC: "right", Title: "Song", Artist: deezerArtist{Name: "Artist"}},
			{ISRC: "wrong", Title: "Song", Artist: deezerArtist{Name: "Somebody Else"}},
		}
		got := pickBest(tracks, ResolveInput{Title: "Song", Artist: "Artist feat. Someone"})
		if got.ISRC != "right" {
			t.Fatalf("got %q, want %q", got.ISRC, "right")
		}
	})

	t.Run("featured artist on track matches plain query artist", func(t *testing.T) {
		tracks := []deezerTrack{
			{ISRC: "right", Title: "Song", Artist: deezerArtist{Name: "Artist feat. Someone"}},
			{ISRC: "wrong", Title: "Song", Artist: deezerArtist{Name: "Somebody Else"}},
		}
		got := pickBest(tracks, ResolveInput{Title: "Song", Artist: "Artist"})
		if got.ISRC != "right" {
			t.Fatalf("got %q, want %q", got.ISRC, "right")
		}
	})
}
