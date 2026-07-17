package kugou

import (
	"bufio"
	"html"
	"regexp"
	"strconv"
	"strings"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
)

var (
	reLine = regexp.MustCompile(`^\[(\d+),(\d+)\](.*)$`)
	reWord = regexp.MustCompile(`<(\d+),(\d+),\d+>([^<]*)`)
)

func isCreditLine(text string) bool {
	return strings.ContainsAny(text, ":：")
}

func stripMetadata(lines []lyrics.Line) []lyrics.Line {
	// scan backwards for the last credit line,
	// then cut out all lines before that
	limit := min(30, len(lines))
	for i := limit - 1; i >= 0; i-- {
		if isCreditLine(lines[i].Text) {
			return lines[i+1:]
		}
	}
	// no credit lines, drop first line if it looks like a title separator
	// TODO: can this really happen?
	if len(lines) > 0 && strings.Contains(lines[0].Text, "-") {
		return lines[1:]
	}
	return lines
}

func parseKRC(text string) []lyrics.Line {
	out := make([]lyrics.Line, 0, 64)
	sc := bufio.NewScanner(strings.NewReader(text))
	for sc.Scan() {
		m := reLine.FindStringSubmatch(sc.Text())
		if m == nil {
			continue
		}
		lineStart, _ := strconv.ParseInt(m[1], 10, 64)
		lineDur, _ := strconv.ParseInt(m[2], 10, 64)
		lineEnd := lineStart + lineDur

		wordMatches := reWord.FindAllStringSubmatch(m[3], -1)
		lineWords := make([]lyrics.Word, 0, len(wordMatches))
		for _, w := range wordMatches {
			t := html.UnescapeString(w[3])
			if t == "" {
				continue
			}

			// kugou is inconsistent with trailing spaces, normalize to at most one
			if trimmed := strings.TrimRight(t, " "); trimmed != t {
				t = trimmed + " "
			}

			offset, _ := strconv.ParseInt(w[1], 10, 64)
			dur, _ := strconv.ParseInt(w[2], 10, 64)
			lineWords = append(lineWords, lyrics.Word{
				StartMs: lineStart + offset,
				EndMs:   lineStart + offset + dur,
				Text:    t,
			})
		}

		var text string
		if len(lineWords) > 0 {
			// sometimes the last word in a line has a trailing space
			last := &lineWords[len(lineWords)-1]
			last.Text = strings.TrimSuffix(last.Text, " ")

			var b strings.Builder
			for _, w := range lineWords {
				b.WriteString(w.Text)
			}
			text = b.String()
		}

		out = append(out, lyrics.Line{
			StartMs: lineStart,
			EndMs:   lineEnd,
			Text:    text,
			Words:   lineWords,
		})
	}
	return stripMetadata(out)
}
