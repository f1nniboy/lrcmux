package normalize

import (
	"regexp"
	"slices"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const featRE = `(?:feat|ft)\.?`

const (
	openBrackets  = `\[(【`
	closeBrackets = `\])】`
)

// only strip Latin combining diacritical marks,
// not Japanese dakuten (important)
// or Arabic harakat (could be stripped, not used in metadata)
var latinDiacritics = &unicode.RangeTable{
	R16: []unicode.Range16{{Lo: 0x0300, Hi: 0x036F, Stride: 1}},
}

var smartQuotes = strings.NewReplacer(
	"‘", "'",
	"’", "'",
	"“", "\"",
	"”", "\"",
)

var artistBrackets = strings.NewReplacer(
	"(", "",
	")", "",
	"[", "",
	"]", "",
)

var (
	// splits multi-artist strings on common feature markers
	collaborationRE = regexp.MustCompile(`(?i)\s*[&,×]\s*|\s+(?:` + featRE + `|and|und|et|con|with|vs\.?|x)\s+`)

	// strips feature credits from song titles
	titleFeatureRE = regexp.MustCompile(`(?i)\s*[` + openBrackets + `]` + featRE + `[^` + closeBrackets + `]*[` + closeBrackets + `]|\s+` + featRE + `\s+\S.*$`)

	// strips parenthetical/bracketed video and audio type markers from titles
	videoSuffixRE = regexp.MustCompile(`(?i)\s*[` + openBrackets + `][^` + closeBrackets + `]*\b(?:video|v[ií]deo|videoclip|musikvideo|musik|clip|audio|lyric(?:s)?|letra|paroles|mv|hd|4k|remaster(?:ed)?|official|offiziell(?:es)?|oficial|ufficiale|officiel(?:le)?)\b[^` + closeBrackets + `]*[` + closeBrackets + `]`)

	// strips production credits from titles
	prodRE = regexp.MustCompile(`(?i)\s*[` + openBrackets + `]prod(?:uced)?\b[^` + closeBrackets + `]*[` + closeBrackets + `]|\s+[|]?\s*prod(?:uced)?\b.*$`)

	// strips instrumental/karaoke version markers from titles
	instrumentalRE = regexp.MustCompile(`(?i)\s*[` + openBrackets + `][^` + closeBrackets + `]*\b(?:instrumental|karaoke)\b[^` + closeBrackets + `]*[` + closeBrackets + `]`)

	// matches the artist–title separator in combined title strings
	artistTitleSepRE = regexp.MustCompile(`\s[-–—~]\s`)
)

func String(s string) string {
	s = smartQuotes.Replace(s)
	t := transform.Chain(norm.NFKD, runes.Remove(runes.In(latinDiacritics)), norm.NFKC)
	if out, _, err := transform.String(t, s); err == nil {
		s = out
	}
	s = strings.ToLower(s)
	return strings.Join(strings.Fields(s), " ")
}

func stripMarkers(s string) string {
	s = videoSuffixRE.ReplaceAllString(s, "")
	s = prodRE.ReplaceAllString(s, "")
	s = instrumentalRE.ReplaceAllString(s, "")
	return s
}

func Title(s string) string {
	s = stripMarkers(s)
	s = titleFeatureRE.ReplaceAllString(s, "")
	return String(s)
}

func artist(s string) string {
	s = String(s)
	s, _ = strings.CutSuffix(s, " - topic")
	return s
}

func splitArtists(s string) []string {
	s = artistBrackets.Replace(s)
	parts := collaborationRE.Split(s, -1)

	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if n := artist(p); n != "" {
			out = append(out, n)
		}
	}
	return out
}

func primaryArtist(s string) string {
	if parts := splitArtists(s); len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func ArtistMatch(a, b string) bool {
	as := splitArtists(a)
	for _, want := range splitArtists(b) {
		if slices.Contains(as, want) {
			return true
		}
	}
	return false
}

func Match(queryTitle, queryArtist, resultTitle, resultArtist string) bool {
	return Title(resultTitle) == Title(queryTitle) && ArtistMatch(resultArtist, queryArtist)
}

func Query(inputArtist, inputTitle string) (cleanArtist, cleanTitle string) {
	inputTitle = stripMarkers(inputTitle)

	// extract artist from title (e.g. YouTube videos)
	if loc := artistTitleSepRE.FindStringIndex(inputTitle); loc != nil {
		inputArtist = inputTitle[:loc[0]]
		inputTitle = inputTitle[loc[1]:]
	}

	return primaryArtist(inputArtist), Title(inputTitle)
}
