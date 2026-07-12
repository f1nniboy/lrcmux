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
			name: "word beats line",
			a:    result("a", lyrics.SyncWord, "x"),
			b:    result("b", lyrics.SyncLine, "x", "y", "z"),
		},
		{
			name: "word beats none",
			a:    result("a", lyrics.SyncWord),
			b:    result("b", lyrics.SyncNone, "x", "y"),
		},
		{
			name: "line beats none",
			a:    result("a", lyrics.SyncLine, "x"),
			b:    result("b", lyrics.SyncNone, "x", "y", "z"),
		},
		{
			name: "more lines wins at equal sync",
			a:    result("a", lyrics.SyncNone, "a", "b", "c", "d"),
			b:    result("b", lyrics.SyncNone, "x", "y"),
		},
		{
			name: "higher sync beats more lines",
			a:    result("a", lyrics.SyncWord, "x"),
			b:    result("b", lyrics.SyncNone, "x", "y", "z"),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if rankResult(c.a, c.b) <= 0 {
				t.Errorf("rankResult(%s, %s) <= 0, want > 0", c.a.Source.ID, c.b.Source.ID)
			}
		})
	}

	t.Run("clean line beats censored word", func(t *testing.T) {
		a := result("clean", lyrics.SyncLine, "hello world")
		b := result("dirty", lyrics.SyncWord, "hello **** world")
		if rankResult(a, b) <= 0 {
			t.Error("clean line-synced should outrank censored word-synced")
		}
	})

	t.Run("otherwise equal: alphabetical source ID wins", func(t *testing.T) {
		a := result("aaa", lyrics.SyncLine, "x", "y")
		b := result("zzz", lyrics.SyncLine, "x", "y")
		if rankResult(a, b) <= 0 {
			t.Error("expected 'aaa' to outrank 'zzz' on alpha tiebreak")
		}
	})
}

func TestPick(t *testing.T) {
	o := testOrch()

	t.Run("no result meets required level", func(t *testing.T) {
		results := []*lyrics.Result{
			result("a", lyrics.SyncNone, "x"),
			result("b", lyrics.SyncLine, "x"),
		}
		if got := o.pick(results, lyrics.SyncWord); got != nil {
			t.Errorf("expected nil, got %s", got.Source.ID)
		}
	})

	t.Run("exact level match accepted", func(t *testing.T) {
		results := []*lyrics.Result{
			result("a", lyrics.SyncLine, "x"),
		}
		if got := o.pick(results, lyrics.SyncLine); got == nil {
			t.Error("exact level match should be accepted")
		}
	})

	t.Run("best ranked result selected", func(t *testing.T) {
		results := []*lyrics.Result{
			result("word", lyrics.SyncWord, "x"),
			result("line", lyrics.SyncLine, "x", "y", "z"),
			result("none", lyrics.SyncNone, "x"),
		}
		got := o.pick(results, lyrics.SyncNone)
		if got == nil || got.Source.ID != "word" {
			t.Errorf("expected word-sync to win, got %v", got)
		}
	})
}

func TestSatisfies(t *testing.T) {
	t.Run("uncensored at level: satisfies", func(t *testing.T) {
		r := result("a", lyrics.SyncWord, "clean")
		if !satisfies(r, lyrics.SyncWord) {
			t.Error("clean word result should satisfy word target")
		}
	})

	t.Run("censored at level: does not satisfy", func(t *testing.T) {
		r := result("a", lyrics.SyncWord, "f**k")
		if satisfies(r, lyrics.SyncWord) {
			t.Error("censored word result should not satisfy word target")
		}
	})

	t.Run("uncensored above level: satisfies", func(t *testing.T) {
		r := result("a", lyrics.SyncWord, "abc")
		if !satisfies(r, lyrics.SyncLine) {
			t.Error("word result should satisfy line target")
		}
	})

	t.Run("uncensored below level: does not satisfy", func(t *testing.T) {
		r := result("a", lyrics.SyncLine, "abc")
		if satisfies(r, lyrics.SyncWord) {
			t.Error("line result should not satisfy word target")
		}
	})
}
