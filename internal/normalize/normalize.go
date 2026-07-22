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

// shared pattern for feature markers
const featRE = `(?:feat|ft)\.?`

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
	titleFeatureRE = regexp.MustCompile(`(?i)\s*[(\[]` + featRE + `[^)\]]*[)\]]|\s+` + featRE + `\s+\S.*$`)

	// strips parenthetical/bracketed video and audio type markers from titles
	videoSuffixRE = regexp.MustCompile(`(?i)\s*[\[(【][^\])】]*\b(?:video|v[ií]deo|videoclip|musikvideo|musik|clip|audio|lyric(?:s)?|letra|paroles|mv|hd|4k|remaster(?:ed)?|official|offiziell(?:es)?|oficial|ufficiale|officiel(?:le)?)\b[^\])】]*[\])】]`)

	// strips production credits from titles
	prodRE = regexp.MustCompile(`(?i)\s*[\[(]prod(?:uced)?\b[^\])]*[\])]|\s+[|]?\s*prod(?:uced)?\b.*$`)

	// matches the artist–title separator in combined title strings
	artistTitleSepRE = regexp.MustCompile(`\s[-–—~]\s`)
)

func String(s string) string {
	s = smartQuotes.Replace(s)
	s = strings.TrimSpace(s)
	t := transform.Chain(norm.NFKD, runes.Remove(runes.In(latinDiacritics)), norm.NFKC)
	if out, _, err := transform.String(t, s); err == nil {
		s = out
	}
	s = strings.ToLower(s)
	return strings.Join(strings.Fields(s), " ")
}

func Title(s string) string {
	s = videoSuffixRE.ReplaceAllString(s, "")
	s = prodRE.ReplaceAllString(s, "")
	s = titleFeatureRE.ReplaceAllString(s, "")
	return String(s)
}

func artist(s string) string {
	s = String(s)
	s, _ = strings.CutSuffix(s, " - topic")
	return strings.TrimSpace(s)
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
	inputTitle = videoSuffixRE.ReplaceAllString(inputTitle, "")
	inputTitle = prodRE.ReplaceAllString(inputTitle, "")
	inputTitle = strings.TrimSpace(inputTitle)

	// extract artist from title (e.g. YouTube videos)
	if loc := artistTitleSepRE.FindStringIndex(inputTitle); loc != nil {
		inputArtist = strings.TrimSpace(inputTitle[:loc[0]])
		inputTitle = strings.TrimSpace(inputTitle[loc[1]:])
	}

	inputTitle = titleFeatureRE.ReplaceAllString(inputTitle, "")
	return primaryArtist(inputArtist), String(inputTitle)
}
