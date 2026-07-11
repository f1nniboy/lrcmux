package normalize

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// shared pattern for feature markers
const featRE = `(?:feat|ft)\.?`

var smartQuotes = strings.NewReplacer(
	"‘", "'",
	"’", "'",
	"“", "\"",
	"”", "\"",
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
	artistTitleSepRE = regexp.MustCompile(`\s[-–~]\s`)
)

func String(s string) string {
	s = strings.TrimSpace(s)
	s = smartQuotes.Replace(s)
	s = strings.ToLower(s)
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	if out, _, err := transform.String(t, s); err == nil {
		s = out
	}
	return strings.Join(strings.Fields(s), " ")
}

func Title(s string) string {
	return String(titleFeatureRE.ReplaceAllString(s, ""))
}

func artist(s string) string {
	s = String(s)
	s, _ = strings.CutSuffix(s, " - topic")
	return strings.TrimSpace(s)
}

func splitArtists(s string) []string {
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
	aNorm := artist(a)
	for _, part := range splitArtists(b) {
		if strings.Contains(aNorm, part) {
			return true
		}
	}
	return false
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
