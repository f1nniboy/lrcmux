package lyrics

import (
	"strings"
	"unicode"
)

// drops what isn't lyrics and undoes watermarks, false when nothing is left
func (r *Result) Clean() bool {
	if r.Instrumental {
		return true
	}
	for i := range r.Lines {
		l := &r.Lines[i]
		var swapped bool
		l.Text, swapped = normalizeText(l.Text, false)
		for j := range l.Words {
			l.Words[j].Text, _ = normalizeText(l.Words[j].Text, swapped)
		}
	}
	r.Lines = tidy(r.Lines)
	return len(r.Lines) > 0
}

// a blank or "♪" line becomes one empty line spanning the whole break, and section
// markers go
func tidy(lines []Line) []Line {
	out := lines[:0]
	for _, l := range lines {
		text := strings.TrimSpace(l.Text)
		if SectionMarker(text) {
			continue
		}
		if breakText(text) {
			if len(out) == 0 {
				continue
			}
			if prev := &out[len(out)-1]; prev.Text == "" {
				prev.EndMs = l.EndMs
				continue
			}
			l = Line{StartMs: l.StartMs, EndMs: l.EndMs}
		}
		out = append(out, l)
	}
	for len(out) > 0 && out[len(out)-1].Text == "" {
		out = out[:len(out)-1]
	}
	return out
}

func breakText(text string) bool {
	return strings.TrimFunc(text, func(r rune) bool {
		return unicode.IsSpace(r) || strings.ContainsRune("♪♫♬♩🎵🎶", r)
	}) == ""
}

// a transcriber's "[Chorus]" or "(Verse: Artist)", unlike an ad-lib "(yeah)"
func SectionMarker(text string) bool {
	return strings.HasPrefix(text, "[") && strings.HasSuffix(text, "]") ||
		strings.HasPrefix(text, "(") && strings.HasSuffix(text, ")") && strings.Contains(text, ":")
}

// some providers watermark lyrics with look-alike apostrophes, letters and spaces
var apostrophes = strings.NewReplacer("\u2019", "'", "\u2018", "'", "\u2032", "'", "\u02bc", "'")

// cyrillic and greek letters drawn like latin ones
var lookalikes = map[rune]rune{
	'а': 'a', 'е': 'e', 'о': 'o', 'р': 'p', 'с': 'c', 'у': 'y', 'х': 'x', 'і': 'i', 'ј': 'j', 'ѕ': 's',
	'А': 'A', 'В': 'B', 'Е': 'E', 'К': 'K', 'М': 'M', 'Н': 'H', 'О': 'O', 'Р': 'P', 'С': 'C', 'Т': 'T',
	'Х': 'X', 'І': 'I', 'Ј': 'J', 'Ѕ': 'S',
	'ο': 'o', 'Α': 'A', 'Β': 'B', 'Ε': 'E', 'Η': 'H', 'Ι': 'I', 'Κ': 'K', 'Μ': 'M', 'Ν': 'N', 'Ο': 'O',
	'Ρ': 'P', 'Τ': 'T', 'Υ': 'Y', 'Χ': 'X', 'Ζ': 'Z',
}

// a word chunk holds no latin letter to go by, so it follows its line
func normalizeText(s string, lineSwapped bool) (string, bool) {
	s = apostrophes.Replace(s)
	s = strings.Map(func(r rune) rune {
		switch {
		case r == '\u200b', r == '\u2060', r == '\ufeff':
			return -1
		case r == '\u00a0', r >= '\u2000' && r <= '\u200a', r == '\u202f', r == '\u205f':
			return ' '
		default:
			return r
		}
	}, s)
	return latinize(s, lineSwapped)
}

// only inside a word that is otherwise latin, so cyrillic and greek lyrics stay
func latinize(s string, lineSwapped bool) (string, bool) {
	swapped := false
	words := strings.SplitAfter(s, " ")
	for i, w := range words {
		latin, foreign, alike := false, false, false
		for _, r := range w {
			switch {
			case unicode.Is(unicode.Latin, r):
				latin = true
			case lookalikes[r] != 0:
				alike = true
			case unicode.IsLetter(r):
				foreign = true
			default:
			}
		}
		if alike && !foreign && (latin || lineSwapped) {
			words[i] = strings.Map(func(r rune) rune {
				if l, ok := lookalikes[r]; ok {
					return l
				}
				return r
			}, w)
			swapped = true
		}
	}
	return strings.Join(words, ""), swapped
}
