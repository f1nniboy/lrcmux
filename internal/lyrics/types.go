package lyrics

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("lyrics not found")

type SyncLevel int

const (
	SyncNone SyncLevel = iota
	SyncLine
	SyncWord
)

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
	Cover    TrackCover `json:"cover"`
	ISRC     string     `json:"isrc"`
	Title    string     `json:"title"`
	Artist   string     `json:"artist"`
	Album    string     `json:"album"`
	Duration int64      `json:"duration"`
}

type Word struct {
	Text    string `json:"text"`
	StartMs int64  `json:"start"`
	EndMs   int64  `json:"end"`
}

type Line struct {
	Text    string `json:"text"`
	Words   []Word `json:"words,omitempty"`
	StartMs int64  `json:"start,omitempty"`
	EndMs   int64  `json:"end,omitempty"`
}

type Source struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

type Result struct {
	Track     Track
	Source    Source
	Lines     []Line
	SyncLevel SyncLevel
}
