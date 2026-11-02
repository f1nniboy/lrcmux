package orchestrator

import (
	"testing"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
)

func TestRankResult(t *testing.T) {
	cases := []struct {
		a    *lyrics.Result
		b    *lyrics.Result
		name string
	}{
		{
			name: "any lyrics beat an instrumental flag",
			a:    result("a", lyrics.SyncLine, "a*a"),
			b:    &lyrics.Result{Source: lyrics.Source{ID: "b"}, Instrumental: true},
		},
		{
			name: "higher sync beats more lines",
			a:    result("a", lyrics.SyncWord, "x"),
			b:    result("b", lyrics.SyncLine, "x", "y", "z"),
		},
		{
			name: "more lines wins at equal sync",
			a:    result("a", lyrics.SyncNone, "a", "b", "c", "d"),
			b:    result("b", lyrics.SyncNone, "x", "y"),
		},
		{
			name: "less censored beats higher sync",
			a:    result("a", lyrics.SyncLine, "a*a"),
			b:    result("b", lyrics.SyncWord, "a*a b*b c*c"),
		},
		{
			name: "otherwise the source id",
			a:    result("aaa", lyrics.SyncLine, "x", "y"),
			b:    result("zzz", lyrics.SyncLine, "x", "y"),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if rankResult(c.a, c.b) <= 0 {
				t.Errorf("rankResult(%s, %s) <= 0, want > 0", c.a.Source.ID, c.b.Source.ID)
			}
		})
	}
}

func TestPick(t *testing.T) {
	o := testOrch()
	results := []*lyrics.Result{result("none", lyrics.SyncNone, "x"), result("line", lyrics.SyncLine, "x")}
	if got := o.pick(results, lyrics.SyncWord); got != nil {
		t.Errorf("got %s below the required level", got.Source.ID)
	}
	if got := o.pick(results, lyrics.SyncLine); got == nil || got.Source.ID != "line" {
		t.Errorf("got %v, want line at its exact level", got)
	}
}
