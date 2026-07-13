package lyrics

import (
	"bufio"
	"strconv"
	"strings"
)

func ParsePlain(s string) []Line {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, "\n")
	out := make([]Line, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, Line{Text: t})
		}
	}
	return out
}

// handles both standard LRC and eLRC (word-sync)
func ParseLRC(body string) ([]Line, SyncLevel) {
	lines := make([]Line, 0, 64)
	hasWords := false

	sc := bufio.NewScanner(strings.NewReader(body))
	for sc.Scan() {
		stamps, text := splitLRCStamps(sc.Text())
		if len(stamps) == 0 {
			continue
		}
		words := splitWordStamps(text)
		cleanText := text
		if len(words) > 0 {
			parts := make([]string, len(words))
			for i, w := range words {
				parts[i] = w.Text
			}
			cleanText = strings.Join(parts, " ")
		}
		for _, ms := range stamps {
			l := Line{StartMs: ms, Text: cleanText}
			if len(words) > 0 {
				hasWords = true
				wc := make([]Word, len(words))
				copy(wc, words)
				l.Words = wc
			}
			lines = append(lines, l)
		}
	}

	// fill line EndMs
	for i := 0; i < len(lines)-1; i++ {
		lines[i].EndMs = lines[i+1].StartMs
	}
	if len(lines) > 0 {
		last := &lines[len(lines)-1]
		last.EndMs = last.StartMs + 5000
	}

	// fill word EndMs now that line EndMs are known
	if hasWords {
		for i := range lines {
			ws := lines[i].Words
			for j := 0; j < len(ws)-1; j++ {
				ws[j].EndMs = ws[j+1].StartMs
			}
			if len(ws) > 0 {
				ws[len(ws)-1].EndMs = lines[i].EndMs
			}
		}
	}

	level := SyncLine
	if hasWords {
		level = SyncWord
	}
	return lines, level
}

// consumes leading [mm:ss.xx] timestamps from an LRC line,
// returning the parsed timestamp and the remaining text
func splitLRCStamps(s string) ([]int64, string) {
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
		ms, ok := parseLRCStamp(inner)
		if !ok {
			break
		}
		stamps = append(stamps, ms)
		rest = remaining
	}
	return stamps, strings.TrimSpace(rest)
}

// parses <mm:ss.xx>word tokens from an eLRC line
func splitWordStamps(s string) []Word {
	if !strings.Contains(s, "<") {
		return nil
	}
	var words []Word
	rest := s
	for strings.Contains(rest, "<") {
		inner, after, ok := strings.Cut(rest[strings.Index(rest, "<")+1:], ">")
		if !ok {
			break
		}
		ms, ok := parseLRCStamp(inner)
		if !ok {
			break
		}
		nextStart := strings.Index(after, "<")
		if nextStart == -1 {
			if t := strings.TrimSpace(after); t != "" {
				words = append(words, Word{StartMs: ms, Text: t})
			}
			break
		}
		if t := strings.TrimSpace(after[:nextStart]); t != "" {
			words = append(words, Word{StartMs: ms, Text: t})
		}
		rest = after[nextStart:]
	}
	return words
}

func parseLRCStamp(s string) (int64, bool) {
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
	ms := int64(mm)*60_000 + int64(secs)*1000
	if hasFrac {
		f, err := strconv.Atoi((fracStr + "000")[:3])
		if err != nil {
			return 0, false
		}
		ms += int64(f)
	}
	return ms, true
}
