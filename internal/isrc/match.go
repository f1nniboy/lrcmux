package isrc

import (
	"cmp"
	"math"
	"regexp"
	"slices"

	"github.com/f1nniboy/lrcmux/internal/normalize"
)

// title and artist each need this similarity, duration and album only rank
const minSimilarity = 0.8

var numberRE = regexp.MustCompile(`\b(?:\d+|[ivx]+)\b`)

// "x - y" can be one title, artist and title, or a title and its version
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
		// prefer the title as written over a reinterpretation of it
		cmp.Compare(b.reading, a.reading),
		cmp.Compare(a.duration, b.duration),
		cmp.Compare(a.album, b.album),
	) > 0
}

type artistScore func(t deezerTrack, in ResolveInput) float64

func best(tracks []deezerTrack, reads []ResolveInput, artist artistScore) (candidate, bool) {
	var out candidate
	found := false
	for i, rd := range reads {
		c, ok := pick(tracks, rd, artist)
		c.reading = i
		if ok && (!found || c.beats(out)) {
			out, found = c, true
		}
	}
	return out, found
}

func byName(t deezerTrack, in ResolveInput) float64 {
	score := artistSimilarity(t.Artist.Name, in.Artist)
	for _, a := range t.Contributors {
		score = max(score, artistSimilarity(a.Name, in.Artist))
	}
	return score
}

func anyArtist(deezerTrack, ResolveInput) float64 { return 1 }

func byID(id int64) artistScore {
	return func(t deezerTrack, _ ResolveInput) float64 {
		if t.Artist.ID == id || slices.ContainsFunc(t.Contributors, func(a deezerArtist) bool { return a.ID == id }) {
			return 1
		}
		return 0
	}
}

func pick(tracks []deezerTrack, in ResolveInput, artist artistScore) (candidate, bool) {
	var best *candidate
	for _, t := range tracks {
		// instrumentals normalize to the song's title but have no lyrics
		if normalize.Instrumental(t.Title) || !sameNumbers(t.Title, in.Title) {
			continue
		}
		exact, title := titleScore(t.Title, in.Title)
		c := candidate{track: t, title: title, artist: artist(t, in)}
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

// a decoration on one side only ("Song (From Film)" for "Song") still matches, below exact
func titleScore(have, want string) (exact bool, score float64) {
	full := similarity(normalize.Title(have), normalize.Title(want))
	return full == 1, max(full, similarity(normalize.CoreTitle(have), normalize.CoreTitle(want)))
}

// "Part 1" and "Part 2" are siblings, a number on one side only is a decoration
func sameNumbers(have, want string) bool {
	a := numberRE.FindAllString(normalize.Title(have), -1)
	b := numberRE.FindAllString(normalize.Title(want), -1)
	return contains(a, b) || contains(b, a)
}

// multiset containment
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

func artistSimilarity(have, want string) float64 {
	if normalize.ArtistMatch(have, want) || normalize.ArtistMatch(want, have) {
		return 1
	}
	return similarity(normalize.String(have), normalize.String(want))
}

func similarity(a, b string) float64 {
	ra, rb := []rune(a), []rune(b)
	n := max(len(ra), len(rb))
	if n == 0 {
		return 0
	}
	return 1 - float64(distance(ra, rb))/float64(n)
}

// osa distance, an adjacent swap is one edit
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
