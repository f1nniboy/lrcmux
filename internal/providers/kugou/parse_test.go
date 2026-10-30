package kugou

import (
	"slices"
	"testing"
)

func TestParseKRC(t *testing.T) {
	tests := []struct {
		name string
		krc  string
		want []string
	}{
		{
			name: "section markers",
			krc: "[0,1000]<0,1000,0>[Intro]\n" +
				"[1000,2000]<0,1000,0>First <1000,1000,0>line\n" +
				"[3000,1000]<0,1000,0>[Verse 1: Artist]\n" +
				"[4000,1000]<0,1000,0>(Chorus: Artist)\n" +
				"[5000,1000]<0,1000,0>(yeah)",
			want: []string{"First line", "(yeah)"},
		},
		{
			name: "interleaved translation",
			krc: "[0,1000]<0,1000,0>作<1000,0,0>词<1000,0,0>：<1000,0,0>某人\n" +
				"[1000,1000]<0,1000,0>作<1000,0,0>曲<1000,0,0>：<1000,0,0>某人\n" +
				"[2000,1000]<0,500,0>First <500,500,0>line\n" +
				"[3000,60]<0,30,0>第<30,30,0>一<60,0,0>行\n" +
				"[4000,1000]<0,500,0>Second <500,500,0>line\n" +
				"[5000,60]<0,30,0>第<30,30,0>二<60,0,0>行\n" +
				"[5500,500]<0,500,0>（间奏）\n" +
				"[6000,1000]<0,1000,0>Third\n" +
				"[7000,60]<0,30,0>第<30,30,0>三<60,0,0>行",
			want: []string{"First line", "Second line", "Third"},
		},
		{
			name: "sung chinese between ad-libs",
			krc: "[0,1000]<0,1000,0>第一行\n" +
				"[1000,1000]<0,1000,0>(Oh)\n" +
				"[2000,1000]<0,1000,0>第二行\n" +
				"[3000,1000]<0,1000,0>(Oh)",
			want: []string{"第一行", "(Oh)", "第二行", "(Oh)"},
		},
		{
			name: "sung chinese",
			krc: "[0,1000]<0,1000,0>Oh\n" +
				"[1000,1000]<0,1000,0>第一行\n" +
				"[2000,1000]<0,1000,0>第二行\n" +
				"[3000,1000]<0,1000,0>Oh",
			want: []string{"Oh", "第一行", "第二行", "Oh"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got []string
			for _, l := range parseKRC(tc.krc, "") {
				got = append(got, l.Text)
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
