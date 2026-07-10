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

func TestCleanQuery(t *testing.T) {
	cases := []struct {
		artist, title, wantArtist, wantTitle string
	}{
		{"Artist", "Title", "Artist", "Title"},
		{"Uploader", "Performer - Song ft. Other", "Performer", "Song"},
		{"Artist", "Artist - Title", "Artist", "Title"},
		{"", "Artist - Title", "", "Artist - Title"},
		{"Uploader", "Performer - Song ft. Other (Official Video)", "Performer", "Song"},
		{"Artist", "Artist - Title (Offizielles Musikvideo)", "Artist", "Title"},
		{"", "Title (Vídeo Oficial)", "", "Title"},
		{"Artist", "Artist - Title (prod by Someone)", "Artist", "Title"},
		{"", "Title (prod. Producer)", "", "Title"},
		{"", "Title (produced by Producer)", "", "Title"},
	}
	for _, c := range cases {
		gotArtist, gotTitle := CleanQuery(c.artist, c.title)
		if gotArtist != c.wantArtist || gotTitle != c.wantTitle {
			t.Errorf("CleanQuery(%q, %q) = (%q, %q), want (%q, %q)",
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
		{"Artxst", []string{"artxst"}},
		{"A & B, C feat D", []string{"a", "b", "c", "d"}},
	}
	for _, c := range cases {
		got := SplitArtists(c.in)
		if len(got) != len(c.want) {
			t.Errorf("SplitArtists(%q) = %v, want %v", c.in, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("SplitArtists(%q)[%d] = %q, want %q", c.in, i, got[i], c.want[i])
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
		if got := PrimaryArtist(c.in); got != c.want {
			t.Errorf("PrimaryArtist(%q) = %q, want %q", c.in, got, c.want)
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
		{"Artist", "Artist and Another", true},
		{"Artist", "Artist und Another", true},
		{"Another", "Artist et Another", true},
		{"Another", "Artist x Another", true},
	}
	for _, c := range cases {
		if got := ArtistMatch(c.sourceArtist, c.inputArtist); got != c.want {
			t.Errorf("ArtistMatch(%q, %q) = %v, want %v", c.sourceArtist, c.inputArtist, got, c.want)
		}
	}
}
