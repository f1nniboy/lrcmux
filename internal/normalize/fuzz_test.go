package normalize

import (
	"testing"
	"unicode/utf8"
)

func FuzzString(f *testing.F) {
	f.Add("Artist Name")
	f.Add("Héros")
	f.Add("")
	f.Add("\t  spaces  \n")
	f.Fuzz(func(t *testing.T, s string) {
		if !utf8.ValidString(s) {
			return
		}
		// normalizing twice should give the same result
		once := String(s)
		if twice := String(once); once != twice {
			t.Errorf("String not idempotent: String(%q) = %q, String(%q) = %q", s, once, once, twice)
		}
	})
}

func FuzzTitle(f *testing.F) {
	f.Add("Song (feat. Someone)")
	f.Add("Title [Official Video]")
	f.Add("")
	f.Fuzz(func(_ *testing.T, s string) {
		Title(s)
	})
}

func FuzzArtistMatch(f *testing.F) {
	f.Add("Artist", "Artist feat. Someone")
	f.Add("Artist (Ft. Someone)", "Artist")
	f.Add("", "")
	f.Fuzz(func(t *testing.T, a, b string) {
		ArtistMatch(a, b)
		// matching against yourself must always hold if there are artists to match
		if parts := splitArtists(a); len(parts) > 0 {
			if !ArtistMatch(a, a) {
				t.Errorf("ArtistMatch(%q, %q) = false, expected true (self-match)", a, a)
			}
		}
	})
}
