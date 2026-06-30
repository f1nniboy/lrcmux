package lyrics

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNotFound = errors.New("lyrics not found")
	ErrTimeout  = errors.New("upstream timeout")
)

type SyncLevel int

const (
	SyncNone SyncLevel = iota
	SyncLine
	SyncWord
)

// Levels lists every sync level from most to least precision.
var Levels = []SyncLevel{SyncWord, SyncLine, SyncNone}

func (s SyncLevel) String() string {
	switch s {
	case SyncWord:
		return "word"
	case SyncLine:
		return "line"
	case SyncNone:
		return "none"
	}
	return "unknown"
}

func (s SyncLevel) Desc() string {
	switch s {
	case SyncWord:
		return "Per-word timestamps, suitable for karaoke-style highlighting"
	case SyncLine:
		return "Per-line timestamps, suitable for scrolling displays"
	case SyncNone:
		return "Unsynced plain text"
	}
	return ""
}

func ParseLevel(s string) (SyncLevel, error) {
	switch s {
	case "":
		return SyncWord, nil
	case "none":
		return SyncNone, nil
	case "line":
		return SyncLine, nil
	case "word":
		return SyncWord, nil
	}
	return SyncNone, fmt.Errorf("invalid sync level %q", s)
}

type Query struct {
	Track Track
}

type TrackCover struct {
	Small  string `json:"small,omitempty"`
	Medium string `json:"medium,omitempty"`
	Big    string `json:"big,omitempty"`
}

type Track struct {
	ISRC     string     `json:"isrc"`
	Title    string     `json:"title"`
	Duration int64      `json:"duration"`
	Artist   string     `json:"artist"`
	Album    string     `json:"album"`
	Cover    TrackCover `json:"cover"`
}

type Word struct {
	StartMs int64  `json:"start"`
	EndMs   int64  `json:"end"`
	Text    string `json:"text"`
}

type Line struct {
	StartMs int64  `json:"start,omitempty"`
	EndMs   int64  `json:"end,omitempty"`
	Text    string `json:"text"`
	Words   []Word `json:"words,omitempty"`
}

type Source struct {
	ID   string
	Name string
}

type Result struct {
	Lines     []Line
	SyncLevel SyncLevel
	Source    Source
	Track     Track
}

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

func Downgrade(r *Result, target SyncLevel) *Result {
	if r == nil || r.SyncLevel <= target {
		return r
	}
	out := *r
	out.SyncLevel = target
	return &out
}
