package kugou

import "testing"

func FuzzParseKRC(f *testing.F) {
	f.Add("[1000,2000]<0,500,0>Hello <500,500,0>world")
	f.Add("")
	f.Add("[invalid]")
	f.Add("[0,0]")
	f.Fuzz(func(_ *testing.T, s string) {
		parseKRC(s)
	})
}
