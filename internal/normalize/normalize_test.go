package normalize

import (
	"testing"
)

func TestString(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Héros", "heros"},
		{"naïve café", "naive cafe"},
		{"'smart' “quotes”", "'smart' \"quotes\""},
		{"Straße", "strasse"},
	}
	for _, c := range cases {
		if got := String(c.in); got != c.want {
			t.Errorf("String(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestTitle(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Title (feat. Someone)", "title"},
		{"Title（feat. Someone）", "title"},
		{"Title [feat. Someone]", "title"},
		{"Title ft. Someone", "title"},
		{"Title feat. Someone Else", "title"},
		{"Title (Instrumental)", "title"},
		{"Title (Karaoke)", "title"},
	}
	for _, c := range cases {
		if got := Title(c.in); got != c.want {
			t.Errorf("Title(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestQuery(t *testing.T) {
	cases := []struct {
		artist, title, wantArtist, wantTitle string
	}{
		{"Artist - Topic", "Title", "artist", "title"},
		{"Artist feat. Someone", "Title", "artist", "title"},
		{"Uploader", "Artist - Title", "uploader", "artist - title"},
		{"Uploader", "Artist - Song ft. Other", "uploader", "artist - song"},
		{"Uploader", "Artist - Song ft. Other (Official Video)", "uploader", "artist - song"},
		{"Artist", "Artist - Title (Offizielles Musikvideo)", "artist", "artist - title"},
		{"", "Title (Official Video)", "", "title"},
		{"", "Title (prod. Producer)", "", "title"},
		{"", "Title (produced by Producer)", "", "title"},
		{"Uploader", "Artist - Title (OFFICIAL VIDEO) | prod. Producer", "uploader", "artist - title"},
		{"Artist", "Title 【Official MV】", "artist", "title"},
	}
	for _, c := range cases {
		gotArtist, gotTitle := Query(c.artist, c.title)
		if gotArtist != c.wantArtist || gotTitle != c.wantTitle {
			t.Errorf("Query(%q, %q) = (%q, %q), want (%q, %q)",
				c.artist, c.title, gotArtist, gotTitle, c.wantArtist, c.wantTitle)
		}
	}
}

func TestSplitArtists(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"Artist", []string{"artist"}},
		{"Artist×Another", []string{"artist", "another"}},
		{"Artxst", []string{"artxst"}},
		{"A & B, C feat D", []string{"a", "b", "c", "d"}},
	}
	for _, c := range cases {
		got := splitArtists(c.in)
		if len(got) != len(c.want) {
			t.Errorf("splitArtists(%q) = %v, want %v", c.in, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("splitArtists(%q)[%d] = %q, want %q", c.in, i, got[i], c.want[i])
			}
		}
	}
}

func TestSplitArtistTitle(t *testing.T) {
	cases := []struct {
		in, artist, title string
		ok                bool
	}{
		{"artist - title", "artist", "title", true},
		{"artist – title", "artist", "title", true},
		{"artist - title - more", "artist", "title - more", true},
		{"artist-title", "", "", false},
		{"artist  -  title", "artist", "title", true},
		{" - title", "", "", false},
		{"artist - ", "", "", false},
		{"title", "", "", false},
	}
	for _, c := range cases {
		artist, title, ok := SplitArtistTitle(c.in)
		if artist != c.artist || title != c.title || ok != c.ok {
			t.Errorf("SplitArtistTitle(%q) = (%q, %q, %v), want (%q, %q, %v)", c.in, artist, title, ok, c.artist, c.title, c.ok)
		}
	}
}

func TestInstrumental(t *testing.T) {
	for in, want := range map[string]bool{
		"Title":                   false,
		"Title (Instrumental)":    true,
		"Title [Karaoke Version]": true,
		"Instrumental":            false,
	} {
		if got := Instrumental(in); got != want {
			t.Errorf("Instrumental(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestCoreTitle(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Title", "title"},
		{"Title (From Something)", "title"},
		{"Title [Something] (Another)", "title"},
		{"Title（Something）", "title"},
		{"(Something)", ""},
	}
	for _, c := range cases {
		if got := CoreTitle(c.in); got != c.want {
			t.Errorf("CoreTitle(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestArtistMatch(t *testing.T) {
	cases := []struct {
		sourceArtist, inputArtist string
		want                      bool
	}{
		{"Artist", "Artist", true},
		{"Artist", "Artist feat. Someone", true},
		{"Someone", "Artist feat. Someone", true},
		{"Someone Else", "Artist", false},
		{"Artist & Friends", "Artist", true},
		{"Artist", "Artist (Ft. Someone)", true},
		{"Someone", "Artist (Ft. Someone)", true},
		{"Artist (DEU)", "Artist", true},
		{"Artist [UK]", "Artist", true},
		{"Artist (feat. Someone)", "Someone", true},
		{"(Artist)", "(Artist)", true},
		{"Artist （DEU）", "Artist", true},
		{"The Artist", "Artist", true},
		{"Artist", "The Artist", true},
		{"Theartist", "Artist", false},
		{"Artist(DEU)", "Artist", true},
	}
	for _, c := range cases {
		if got := ArtistMatch(c.sourceArtist, c.inputArtist); got != c.want {
			t.Errorf("ArtistMatch(%q, %q) = %v, want %v", c.sourceArtist, c.inputArtist, got, c.want)
		}
	}
}
