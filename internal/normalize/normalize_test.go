package normalize

import (
	"testing"
)

func TestString(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Artist Name", "artist name"},
		{"  spaces  ", "spaces"},
		{"Héros", "heros"},
		{"naïve café", "naive cafe"},
		{"'smart' “quotes”", "'smart' \"quotes\""},
		{"collapse   whitespace", "collapse whitespace"},
		{"UPPER", "upper"},
	}
	for _, c := range cases {
		if got := String(c.in); got != c.want {
			t.Errorf("String(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestArtist(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Artist", "artist"},
		{"Artist - Topic", "artist"},
		{"Artist - TOPIC", "artist"},
		{"Héros - Topic", "heros"},
	}
	for _, c := range cases {
		if got := artist(c.in); got != c.want {
			t.Errorf("artist(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestTitle(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Song Title", "song title"},
		{"Title (feat. Someone)", "title"},
		{"Title (ft. Someone)", "title"},
		{"Title [feat. Someone]", "title"},
		{"Title ft. Someone", "title"},
		{"Title feat. Someone Else", "title"},
		{"Title", "title"},
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
		{"Artist", "Title", "artist", "title"},
		{"Artist - Topic", "Title", "artist", "title"},
		{"Artist feat. Someone", "Title", "artist", "title"},
		{"Uploader", "Artist - Title", "artist", "title"},
		{"Uploader", "Artist – Title", "artist", "title"},
		{"Uploader", "Artist ~ Title", "artist", "title"},
		{"Uploader", "Artist - Song ft. Other", "artist", "song"},
		{"Artist", "Artist - Title", "artist", "title"},
		{"Uploader", "Artist - Song ft. Other (Official Video)", "artist", "song"},
		{"Artist", "Artist - Title (Offizielles Musikvideo)", "artist", "title"},
		{"", "Title (Official Video)", "", "title"},
		{"Artist", "Artist - Title (prod by Someone)", "artist", "title"},
		{"", "Title (prod. Producer)", "", "title"},
		{"", "Title (produced by Producer)", "", "title"},
		{"Uploader", "Artist - Title (OFFICIAL VIDEO) prod. by Producer", "artist", "title"},
		{"Uploader", "Artist - Title (OFFICIAL VIDEO) | prod. Producer", "artist", "title"},
		{"Artist", "Title 【MV】", "artist", "title"},
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
		{"Artist & Another", []string{"artist", "another"}},
		{"Artist, Another", []string{"artist", "another"}},
		{"Artist feat. Another", []string{"artist", "another"}},
		{"Artist ft. Another", []string{"artist", "another"}},
		{"Artist feat Another", []string{"artist", "another"}},
		{"Artist and Another", []string{"artist", "another"}},
		{"Artist und Another", []string{"artist", "another"}},
		{"Artist et Another", []string{"artist", "another"}},
		{"Artist x Another", []string{"artist", "another"}},
		{"Artist con Another", []string{"artist", "another"}},
		{"Artist with Another", []string{"artist", "another"}},
		{"Artist vs Another", []string{"artist", "another"}},
		{"Artist vs. Another", []string{"artist", "another"}},
		{"Artist × Another", []string{"artist", "another"}},
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

func TestPrimaryArtist(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Artist", "artist"},
		{"Artist & Another", "artist"},
		{"Artist feat. Another", "artist"},
		{"", ""},
	}
	for _, c := range cases {
		if got := primaryArtist(c.in); got != c.want {
			t.Errorf("primaryArtist(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestMatch(t *testing.T) {
	if !Match("Song", "Artist", "Song", "Artist") {
		t.Error("identical title and artist should match")
	}
	if Match("Song", "Artist", "Other Song", "Artist") {
		t.Error("different title should not match")
	}
	if Match("Song", "Artist", "Song", "Other Artist") {
		t.Error("different artist should not match")
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
		{"Artist", "Artist and Another", true},
		{"Artist", "Artist und Another", true},
		{"Another", "Artist et Another", true},
		{"Another", "Artist x Another", true},
		{"Artist", "Artist (Ft. Someone)", true},
		{"Someone", "Artist (Ft. Someone)", true},
	}
	for _, c := range cases {
		if got := ArtistMatch(c.sourceArtist, c.inputArtist); got != c.want {
			t.Errorf("ArtistMatch(%q, %q) = %v, want %v", c.sourceArtist, c.inputArtist, got, c.want)
		}
	}
}
