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
	if len(lines) == 0 {
		return lines
	}
	lines = lines[1:]
	for len(lines) > 0 && isCreditLine(lines[0].Text) {
		lines = lines[1:]
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

		wordMatches := reWord.FindAllStringSubmatch(m[3], -1)
		if len(wordMatches) == 0 {
			continue
		}

		lineWords := make([]lyrics.Word, 0, len(wordMatches))
		var parts []string
		for _, w := range wordMatches {
			t := strings.TrimSpace(html.UnescapeString(w[3]))
			if t == "" {
				continue
			}
			offset, _ := strconv.ParseInt(w[1], 10, 64)
			dur, _ := strconv.ParseInt(w[2], 10, 64)
			lineWords = append(lineWords, lyrics.Word{
				StartMs: lineStart + offset,
				EndMs:   lineStart + offset + dur,
				Text:    t,
			})
			parts = append(parts, t)
		}

		lineText := strings.Join(parts, " ")
		if lineText == "" {
			continue
		}

		out = append(out, lyrics.Line{
			StartMs: lineStart,
			EndMs:   lineStart + lineDur,
			Text:    lineText,
			Words:   lineWords,
		})
	}
	return stripMetadata(out)
}
