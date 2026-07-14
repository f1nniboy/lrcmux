package lyrics

import "testing"

func FuzzParseLRC(f *testing.F) {
	f.Add("[01:23.45]Hello world")
	f.Add("[00:00.00]<00:00.00>Hello <00:01.00>world")
	f.Add("[99:59.99]")
	f.Add("")
	f.Add("[invalid]text")
	f.Add("[01:23.45][02:34.56]multi-stamp line")
	f.Fuzz(func(_ *testing.T, s string) {
		ParseLRC(s)
	})
}

func FuzzParsePlain(f *testing.F) {
	f.Add("Hello\nworld")
	f.Add("")
	f.Add("\r\nwindows\r\nline endings\r\n")
	f.Fuzz(func(_ *testing.T, s string) {
		ParsePlain(s)
	})
}
