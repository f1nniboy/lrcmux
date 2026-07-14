package musixmatch

import (
	"encoding/json"
	"testing"
)

func FuzzParseRichsync(f *testing.F) {
	f.Add(1.0, 2.0, "Hello", 0.0)
	f.Add(0.0, 0.0, "", 0.0)
	f.Add(1.0, 1.0, "word", 0.5)
	f.Fuzz(func(_ *testing.T, ts, te float64, content string, offset float64) {
		type entry struct {
			Lines []struct {
				Content   string  `json:"c"`
				OffsetSec float64 `json:"o"`
			} `json:"l"`
			StartSec float64 `json:"ts"`
			EndSec   float64 `json:"te"`
		}
		data, _ := json.Marshal([]entry{{
			StartSec: ts,
			EndSec:   te,
			Lines: []struct {
				Content   string  `json:"c"`
				OffsetSec float64 `json:"o"`
			}{{Content: content, OffsetSec: offset}},
		}})
		parseRichsync(string(data))
	})
}

func FuzzParseSubtitles(f *testing.F) {
	f.Add("Hello", 1.0)
	f.Add("", 0.0)
	f.Fuzz(func(_ *testing.T, text string, total float64) {
		type entry struct {
			Text string `json:"text"`
			Time struct {
				Total float64 `json:"total"`
			} `json:"time"`
		}
		data, _ := json.Marshal([]entry{{Text: text, Time: struct {
			Total float64 `json:"total"`
		}{Total: total}}})
		parseSubtitles(string(data))
	})
}
