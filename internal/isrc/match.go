package isrc

import (
	"cmp"
	"math"
	"regexp"
	"slices"

	"github.com/f1nniboy/lrcmux/internal/normalize"
)

// a candidate needs its title and its artist to each agree at least this much,
// while duration and album only rank the candidates that do
const minSimilarity = 0.8

var numberRE = regexp.MustCompile(`\b(?:\d+|[ivx]+)\b`)

// the dash in "x - y" is part of the title ("Garakuta - JUNK"), splits artist
// from title the way youtube uploads do, or follows the title with a version
// ("Title - Remastered 2011"), so every reading is scored against the candidates
func readings(in ResolveInput) []ResolveInput {
	out := []ResolveInput{in}
	if head, tail, ok := normalize.SplitArtistTitle(in.Title); ok {
		split, bare := in, in
		split.Artist, split.Title = head, tail
		bare.Title = head
		out = append(out, split, bare)
	}
	return out
}

type candidate struct {
	track    deezerTrack
	exact    int
	title    float64
	artist   float64
	reading  int
	duration float64
	album    float64
}

func (a candidate) beats(b candidate) bool {
	return cmp.Or(
		cmp.Compare(a.exact, b.exact),
		cmp.Compare(a.title, b.title),
		cmp.Compare(a.artist, b.artist),
		// the title as written beats a reading of it, before duration has a say
		cmp.Compare(b.reading, a.reading),
		cmp.Compare(a.duration, b.duration),
		cmp.Compare(a.album, b.album),
	) > 0
}

func pickBest(tracks []deezerTrack, in ResolveInput, artistID int64) (candidate, bool) {
	var best *candidate
	for _, t := range tracks {
		// an instrumental or karaoke version has no lyrics of its own, yet its
		// title normalizes to the song's
		if normalize.Instrumental(t.Title) || !sameNumbers(t.Title, in.Title) {
			continue
		}
		exact, title := titleScore(t.Title, in.Title)
		c := candidate{track: t, title: title, artist: artistSimilarity(t.Artist, in.Artist, artistID)}
		if c.title < minSimilarity || c.artist < minSimilarity {
			continue
		}
		if exact {
			c.exact = 1
		}
		if in.Duration > 0 {
			c.duration = -math.Abs(float64(t.Duration - in.Duration))
		}
		if in.Album != "" {
			c.album = similarity(normalize.String(t.Album.Title), normalize.String(in.Album))
		}
		if best == nil || c.beats(*best) {
			best = &c
		}
	}
	if best == nil {
		return candidate{}, false
	}
	return *best, true
}

// exact when the full titles agree, scored by the closer of the full titles and
// their cores, so a decoration on one side only, like "Summertime (From 'Porgy &
// Bess')" against "Summertime", still names the song, though below an exact one
func titleScore(have, want string) (exact bool, score float64) {
	full := similarity(normalize.Title(have), normalize.Title(want))
	return full == 1, max(full, similarity(normalize.CoreTitle(have), normalize.CoreTitle(want)))
}

func titleMatches(tracks []deezerTrack, title string) bool {
	return slices.ContainsFunc(tracks, func(t deezerTrack) bool {
		_, score := titleScore(t.Title, title)
		return score >= minSimilarity
	})
}

// numbers that disagree set siblings apart ("Part 1" and "Part 2", "Song I" and
// "Song II"), while numbers only one side has are a decoration ("Live 1999")
func sameNumbers(have, want string) bool {
	a := numberRE.FindAllString(normalize.Title(have), -1)
	b := numberRE.FindAllString(normalize.Title(want), -1)
	return contains(a, b) || contains(b, a)
}

// reports whether every number of sub is in of, counting repeats
func contains(of, sub []string) bool {
	left := slices.Clone(of)
	for _, n := range sub {
		i := slices.Index(left, n)
		if i < 0 {
			return false
		}
		left = slices.Delete(left, i, i+1)
	}
	return true
}

// id is deezer's own pick for want, or 0 when not asked
func artistSimilarity(have deezerArtist, want string, id int64) float64 {
	if (id != 0 && have.ID == id) || normalize.ArtistMatch(have.Name, want) || normalize.ArtistMatch(want, have.Name) {
		return 1
	}
	return similarity(normalize.String(have.Name), normalize.String(want))
}

// share of the longer string left unchanged, so one edit costs a short title as
// much as it should
func similarity(a, b string) float64 {
	ra, rb := []rune(a), []rune(b)
	n := max(len(ra), len(rb))
	if n == 0 {
		return 0
	}
	return 1 - float64(distance(ra, rb))/float64(n)
}

// edit distance that counts swapping two adjacent characters, the commonest
// typo, as one edit where plain levenshtein counts two
func distance(a, b []rune) int {
	prev2, prev, cur := make([]int, len(b)+1), make([]int, len(b)+1), make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
			if i > 1 && j > 1 && a[i-1] == b[j-2] && a[i-2] == b[j-1] {
				cur[j] = min(cur[j], prev2[j-2]+1)
			}
		}
		prev2, prev, cur = prev, cur, prev2
	}
	return prev[len(b)]
}
