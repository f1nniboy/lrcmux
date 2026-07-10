package normalize

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// shared pattern for "feat", "feat.", "ft", "ft."
const featRE = `(?:feat|ft)\.?`

var smartQuotes = strings.NewReplacer(
	"‘", "'",
	"’", "'",
	"“", "\"",
	"”", "\"",
)

var (
	// splits multi-artist strings on "&", "," and feat/ft markers,
	// plus the conjunctions "and" (English), "und" (German), "et" (French),
	// and the "x" marker
	collaborationRE = regexp.MustCompile(`(?i)\s*[&,]\s*|\s+(?:` + featRE + `|and|und|et|x)\s+`)

	// strips feature credits from song titles: both parenthetical
	// "(feat. X)" / "[ft. X]" and bare suffix " ft. X ...".
	titleFeatureRE = regexp.MustCompile(`(?i)\s*[(\[]` + featRE + `[^)\]]*[)\]]|\s+` + featRE + `\s+\S.*$`)

	// strips parenthetical/bracketed video and audio type markers from titles
	// in several languages, e.g. "(Official Video)", "(Offizielles Musikvideo)", "(Vídeo Oficial)",
	// "(Lyric Video)", "(4K Remastered)"
	videoSuffixRE = regexp.MustCompile(`(?i)\s*[\[(][^\])]*\b(?:video|v[ií]deo|videoclip|musikvideo|musik|clip|audio|lyric(?:s)?|letra|paroles|mv|hd|4k|remaster(?:ed)?|official|offiziell(?:es)?|oficial|ufficiale|officiel(?:le)?)\b[^\])]*[\])]`)

	// strips production credits from titles, e.g. "(prod by someone)", "(prod. someone)"
	prodRE = regexp.MustCompile(`(?i)\s*[\[(]prod(?:uced)?\b[^\])]*[\])]`)
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

func SplitArtists(s string) []string {
	parts := collaborationRE.Split(s, -1)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if n := String(p); n != "" {
			out = append(out, n)
		}
	}
	return out
}

func PrimaryArtist(s string) string {
	if parts := SplitArtists(s); len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func ArtistMatch(a, b string) bool {
	na := String(a)
	for _, part := range SplitArtists(b) {
		if strings.Contains(na, part) {
			return true
		}
	}
	return false
}

func CleanQuery(artist, title string) (cleanArtist, cleanTitle string) {
	title = videoSuffixRE.ReplaceAllString(title, "")
	title = prodRE.ReplaceAllString(title, "")
	title = strings.TrimSpace(title)
	if artist != "" {
		if idx := strings.Index(title, " - "); idx != -1 {
			artist = strings.TrimSpace(title[:idx])
			title = strings.TrimSpace(title[idx+3:])
		}
	}
	title = titleFeatureRE.ReplaceAllString(title, "")
	return artist, strings.TrimSpace(title)
}
