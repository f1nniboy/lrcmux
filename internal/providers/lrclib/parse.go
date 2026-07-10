package lrclib

import (
	"bufio"
	"strconv"
	"strings"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
)

func parseSynced(body string) []lyrics.Line {
	lines := make([]lyrics.Line, 0, 64)
	sc := bufio.NewScanner(strings.NewReader(body))
	for sc.Scan() {
		raw := sc.Text()
		stamps, text := splitStamps(raw)
		if len(stamps) == 0 {
			continue
		}
		if text == "" {
			continue
		}
		for _, ms := range stamps {
			lines = append(lines, lyrics.Line{StartMs: ms, Text: text})
		}
	}
	for i := 1; i < len(lines); i++ {
		j := i
		for j > 0 && lines[j-1].StartMs > lines[j].StartMs {
			lines[j-1], lines[j] = lines[j], lines[j-1]
			j--
		}
	}
	for i := 0; i < len(lines)-1; i++ {
		lines[i].EndMs = lines[i+1].StartMs
	}
	if len(lines) > 0 {
		last := &lines[len(lines)-1]
		last.EndMs = last.StartMs + 5000
	}
	return lines
}

func splitStamps(s string) ([]int64, string) {
	var stamps []int64
	rest := s
	for {
		rest = strings.TrimLeft(rest, " \t")
		after, ok := strings.CutPrefix(rest, "[")
		if !ok {
			break
		}
		inner, remaining, ok := strings.Cut(after, "]")
		if !ok {
			break
		}
		ms, ok := parseStamp(inner)
		if !ok {
			break
		}
		stamps = append(stamps, ms)
		rest = remaining
	}
	return stamps, strings.TrimSpace(rest)
}

func parseStamp(s string) (int64, bool) {
	mmStr, rest, ok := strings.Cut(s, ":")
	if !ok {
		return 0, false
	}
	mm, err := strconv.Atoi(mmStr)
	if err != nil {
		return 0, false
	}
	secsStr, fracStr, hasFrac := strings.Cut(rest, ".")
	secs, err := strconv.Atoi(secsStr)
	if err != nil {
		return 0, false
	}
	var frac int
	if hasFrac {
		if len(fracStr) > 3 {
			fracStr = fracStr[:3]
		}
		f, err := strconv.Atoi(fracStr)
		if err != nil {
			return 0, false
		}
		switch len(fracStr) {
		case 1:
			frac = f * 100
		case 2:
			frac = f * 10
		case 3:
			frac = f
		default:
		}
	}
	return int64(mm)*60_000 + int64(secs)*1000 + int64(frac), true
}
