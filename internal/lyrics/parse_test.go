package lyrics

import (
	"testing"
)

func TestParsePlain(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		lines := ParsePlain("Hello\nWorld")
		if len(lines) != 2 || lines[0].Text != "Hello" || lines[1].Text != "World" {
			t.Errorf("got %v", lines)
		}
	})

	t.Run("CRLF normalized", func(t *testing.T) {
		lines := ParsePlain("Hello\r\nWorld")
		if len(lines) != 2 || lines[0].Text != "Hello" || lines[1].Text != "World" {
			t.Errorf("expected 2 lines, got %v", lines)
		}
	})

	t.Run("empty string returns nil", func(t *testing.T) {
		if ParsePlain("") != nil {
			t.Error("expected nil")
		}
	})

	t.Run("whitespace-only lines preserved as empty", func(t *testing.T) {
		lines := ParsePlain("Hello\n   \nWorld")
		if len(lines) != 3 || lines[1].Text != "" {
			t.Errorf("expected 3 lines with line 2 being empty, got %v", lines)
		}
	})
}

func TestParseLRC(t *testing.T) {
	t.Run("basic line sync returns SyncLine", func(t *testing.T) {
		lines, level := ParseLRC("[00:01.00]Line one\n[00:03.50]Line two\n[00:06.00]Line three")
		if level != SyncLine {
			t.Errorf("expected SyncLine, got %s", level)
		}
		if len(lines) != 3 {
			t.Fatalf("expected 3 lines, got %d", len(lines))
		}
		if lines[0].StartMs != 1000 || lines[0].Text != "Line one" {
			t.Errorf("line 0: got %+v", lines[0])
		}
		if lines[0].Words != nil {
			t.Error("line-sync line should have no Words")
		}
	})

	t.Run("line EndMs filled from next line start", func(t *testing.T) {
		lines, _ := ParseLRC("[00:01.00]A\n[00:04.00]B")
		if lines[0].EndMs != 4000 {
			t.Errorf("expected EndMs=4000, got %d", lines[0].EndMs)
		}
	})

	t.Run("multiple stamps on one line produce multiple lines", func(t *testing.T) {
		lines, _ := ParseLRC("[00:01.00][00:05.00]Chorus")
		if len(lines) != 2 {
			t.Fatalf("expected 2 lines, got %d", len(lines))
		}
		if lines[0].StartMs != 1000 || lines[1].StartMs != 5000 {
			t.Errorf("wrong start times: %d %d", lines[0].StartMs, lines[1].StartMs)
		}
		if lines[0].Text != "Chorus" || lines[1].Text != "Chorus" {
			t.Error("both lines should have same text")
		}
	})

	t.Run("metadata lines skipped", func(t *testing.T) {
		input := "[ar:Artist]\n[ti:Title]\n[00:01.00]Hello"
		lines, _ := ParseLRC(input)
		if len(lines) != 1 || lines[0].Text != "Hello" {
			t.Errorf("expected 1 lyric line, got %d: %v", len(lines), lines)
		}
	})

	t.Run("empty timestamps preserved", func(t *testing.T) {
		lines, _ := ParseLRC("[00:01.00]Text\n[00:02.00]\n[00:03.00]Another one")
		if len(lines) != 3 {
			t.Fatalf("expected 3 lines, got %v", lines)
		}
		if lines[1].Text != "" || lines[1].StartMs != 2000 {
			t.Errorf("expected empty-text line at 2000ms, got %+v", lines[0])
		}
		if lines[0].Text != "Text" || lines[0].StartMs != 1000 {
			t.Errorf("expected Text line at 1000ms, got %+v", lines[1])
		}
	})

	t.Run("empty body returns empty slice", func(t *testing.T) {
		lines, level := ParseLRC("")
		if len(lines) != 0 {
			t.Errorf("expected empty, got %v", lines)
		}
		if level != SyncLine {
			t.Errorf("expected SyncLine for empty input, got %s", level)
		}
	})

	t.Run("eLRC with words returns correct level", func(t *testing.T) {
		input := "[00:01.00]<00:01.00>Never <00:02.00>gonna <00:03.00>give\n[00:05.00]<00:05.00>you <00:06.00>up"
		lines, level := ParseLRC(input)
		if level != SyncWord {
			t.Errorf("expected SyncWord, got %s", level)
		}
		if len(lines) != 2 {
			t.Fatalf("expected 2 lines, got %d", len(lines))
		}
		if lines[0].Text != "Never gonna give" {
			t.Errorf("wrong text: %q", lines[0].Text)
		}
		if len(lines[0].Words) != 3 {
			t.Fatalf("expected 3 words, got %d", len(lines[0].Words))
		}
	})

	t.Run("word StartMs parsed correctly", func(t *testing.T) {
		lines, _ := ParseLRC("[00:01.00]<00:01.20>Hello <00:02.40>world")
		ws := lines[0].Words
		if ws[0].StartMs != 1200 {
			t.Errorf("word 0 StartMs: want 1200, got %d", ws[0].StartMs)
		}
		if ws[1].StartMs != 2400 {
			t.Errorf("word 1 StartMs: want 2400, got %d", ws[1].StartMs)
		}
	})

	t.Run("word EndMs: intermediate word ends at next word start", func(t *testing.T) {
		lines, _ := ParseLRC("[00:00.00]<00:00.00>A <00:01.00>B <00:02.00>C")
		ws := lines[0].Words
		if ws[0].EndMs != 1000 {
			t.Errorf("word 0 EndMs: want 1000, got %d", ws[0].EndMs)
		}
		if ws[1].EndMs != 2000 {
			t.Errorf("word 1 EndMs: want 2000, got %d", ws[1].EndMs)
		}
	})

	t.Run("last word EndMs equals line EndMs", func(t *testing.T) {
		lines, _ := ParseLRC("[00:00.00]<00:00.00>Only <00:01.00>two\n[00:05.00]<00:05.00>Next line")
		ws := lines[0].Words
		if ws[len(ws)-1].EndMs != lines[0].EndMs {
			t.Errorf("last word EndMs %d != line EndMs %d", ws[len(ws)-1].EndMs, lines[0].EndMs)
		}
		if lines[0].EndMs != 5000 {
			t.Errorf("line EndMs should be 5000, got %d", lines[0].EndMs)
		}
	})

	t.Run("fractional timestamp precision", func(t *testing.T) {
		// 2-digit (centiseconds)
		lines, _ := ParseLRC("[01:23.45]Text")
		if lines[0].StartMs != 83450 {
			t.Errorf("2-digit frac: want 83450, got %d", lines[0].StartMs)
		}
		// 3-digit (milliseconds)
		lines, _ = ParseLRC("[01:23.456]Text")
		if lines[0].StartMs != 83456 {
			t.Errorf("3-digit frac: want 83456, got %d", lines[0].StartMs)
		}
	})

	t.Run("stamp without fraction parsed correctly", func(t *testing.T) {
		lines, _ := ParseLRC("[00:01]Text")
		if len(lines) != 1 || lines[0].StartMs != 1000 {
			t.Errorf("want 1 line at 1000ms, got %v", lines)
		}
	})

	t.Run("stamp without colon skipped", func(t *testing.T) {
		lines, _ := ParseLRC("[invalid]Text\n[00:01.00]Good")
		if len(lines) != 1 || lines[0].Text != "Good" {
			t.Errorf("expected only the valid line, got %v", lines)
		}
	})

	t.Run("unclosed bracket skipped", func(t *testing.T) {
		lines, _ := ParseLRC("[00:01.00 Text\n[00:02.00]Good")
		if len(lines) != 1 || lines[0].Text != "Good" {
			t.Errorf("expected only the valid line, got %v", lines)
		}
	})

	t.Run("malformed word stamp falls back gracefully", func(t *testing.T) {
		lines, level := ParseLRC("[00:01.00]<bad>word")
		if level != SyncLine {
			t.Errorf("expected SyncLine, got %s", level)
		}
		if len(lines) != 1 || lines[0].Words != nil {
			t.Errorf("expected 1 line with no words, got %v", lines)
		}
	})

	t.Run("non-numeric seconds skipped", func(t *testing.T) {
		lines, _ := ParseLRC("[00:ab.00]Bad\n[00:01.00]Good")
		if len(lines) != 1 || lines[0].Text != "Good" {
			t.Errorf("expected only valid line, got %v", lines)
		}
	})

	t.Run("non-numeric fraction skipped", func(t *testing.T) {
		lines, _ := ParseLRC("[00:01.xy]Bad\n[00:02.00]Good")
		if len(lines) != 1 || lines[0].Text != "Good" {
			t.Errorf("expected only valid line, got %v", lines)
		}
	})

	t.Run("unclosed word stamp angle bracket falls back to no words", func(t *testing.T) {
		lines, level := ParseLRC("[00:01.00]<notclosed")
		if level != SyncLine {
			t.Errorf("expected SyncLine, got %s", level)
		}
		if len(lines) != 1 || lines[0].Words != nil {
			t.Errorf("expected 1 line with no words, got %v", lines)
		}
	})

	t.Run("trailing word stamp with no text produces no extra word", func(t *testing.T) {
		lines, _ := ParseLRC("[00:01.00]<00:01.00>Hello <00:02.00>")
		if len(lines[0].Words) != 1 || lines[0].Words[0].Text != "Hello" {
			t.Errorf("expected 1 word 'Hello', got %v", lines[0].Words)
		}
	})
}
