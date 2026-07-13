package netease

import (
	"bufio"
	"regexp"
	"strconv"
	"strings"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
)

var creditRE = regexp.MustCompile(`^\S{1,6}\s*[：:]\s*\S`)

func parseYRC(body string) []lyrics.Line {
	lines := make([]lyrics.Line, 0, 64)
	sc := bufio.NewScanner(strings.NewReader(body))
	for sc.Scan() {
		lineStart, lineEnd, rest, ok := parseYRCHeader(sc.Text())
		if !ok {
			continue
		}
		words, text := parseYRCWords(rest)
		lines = append(lines, lyrics.Line{StartMs: lineStart, EndMs: lineEnd, Text: text, Words: words})
	}
	return lines
}

//nolint:revive
func parseYRCHeader(s string) (start, end int64, rest string, ok bool) {
	after, cut := strings.CutPrefix(strings.TrimSpace(s), "[")
	if !cut {
		return
	}
	inner, remaining, cut := strings.Cut(after, "]")
	if !cut {
		return
	}
	startStr, durStr, cut := strings.Cut(inner, ",")
	if !cut {
		return
	}
	startMs, startErr := strconv.ParseInt(startStr, 10, 64)
	durMs, durErr := strconv.ParseInt(strings.SplitN(durStr, ",", 2)[0], 10, 64)
	if startErr != nil || durErr != nil {
		return
	}
	return startMs, startMs + durMs, remaining, true
}

func parseYRCWords(s string) ([]lyrics.Word, string) {
	var words []lyrics.Word
	rest := s
	for {
		idx := strings.Index(rest, "(")
		if idx == -1 {
			break
		}
		inner, after, ok := strings.Cut(rest[idx+1:], ")")
		if !ok {
			break
		}
		parts := strings.SplitN(inner, ",", 3)
		if len(parts) < 2 {
			break
		}
		wStart, startErr := strconv.ParseInt(parts[0], 10, 64)
		wDur, durErr := strconv.ParseInt(parts[1], 10, 64)
		if startErr != nil || durErr != nil {
			break
		}
		nextIdx := strings.Index(after, "(")
		var raw string
		if nextIdx == -1 {
			raw = after
		} else {
			raw = after[:nextIdx]
			rest = after[nextIdx:]
		}
		words = append(words, lyrics.Word{StartMs: wStart, EndMs: wStart + wDur, Text: raw})
		if nextIdx == -1 {
			break
		}
	}
	var b strings.Builder
	for _, w := range words {
		b.WriteString(w.Text)
	}
	return words, b.String()
}

// detects the "lyrics yet to be released" filler text
func hasPlaceholder(lines []lyrics.Line) bool {
	for _, l := range lines {
		if strings.Contains(l.Text, "yet to be released") {
			return true
		}
	}
	return false
}

func cleanLines(lines []lyrics.Line) []lyrics.Line {
	out := lines[:0:len(lines)]
	for _, l := range lines {
		l.Text = uncensor(halfWidth(l.Text))
		if l.StartMs == 0 || creditRE.MatchString(strings.TrimSpace(l.Text)) {
			continue
		}
		for j := range l.Words {
			l.Words[j].Text = uncensor(halfWidth(l.Words[j].Text))
		}
		out = append(out, l)
	}
	return out
}

// converts full-width chars to normal equivalents
func halfWidth(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r >= 0xFF01 && r <= 0xFF5E:
			b.WriteRune(r - 0xFEE0)
		case r == 0x3000:
			b.WriteByte(' ')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// fix netease's simple censor method, by replacing '!' with 'i' if
// the character is NOT the last character in a line
func uncensor(s string) string {
	end := strings.LastIndexFunc(s, func(r rune) bool { return r != ' ' && r != '\t' })
	b := []byte(s)
	for i := range b {
		if b[i] == '!' && i != end {
			b[i] = 'i'
		}
	}
	return string(b)
}
