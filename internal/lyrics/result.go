package lyrics

import "strings"

// filters out empty lines and bracketed section labels such as
// [Intro], [Verse 1] etc.
func CleanLines(lines []Line) []Line {
	out := lines[:0:len(lines)]
	for _, l := range lines {
		t := strings.TrimSpace(l.Text)
		if t == "" {
			continue
		}
		if strings.HasPrefix(t, "[") && strings.HasSuffix(t, "]") {
			continue
		}
		out = append(out, l)
	}
	return out
}

func (r *Result) Downgrade(target SyncLevel) *Result {
	if r == nil || r.SyncLevel <= target {
		return r
	}
	out := *r
	out.SyncLevel = target
	return &out
}
