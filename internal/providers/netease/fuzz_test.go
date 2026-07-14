package netease

import "testing"

func FuzzParseYRC(f *testing.F) {
	f.Add("[1000,2000]（1000,500,0）Hello （1500,500,0）world")
	f.Add("")
	f.Add("[invalid]")
	f.Add("[0,0](0,0,0)")
	f.Fuzz(func(_ *testing.T, s string) {
		parseYRC(s)
	})
}
