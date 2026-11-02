package genius

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
)

var payloadRE = regexp.MustCompile(`JSON\.parse\('((?:[^'\\]|\\.)*)'\)`)

func parseEmbed(script string) ([]lyrics.Line, error) {
	m := payloadRE.FindStringSubmatch(script)
	if m == nil {
		return nil, errors.New("embed payload not found")
	}
	var html string
	if err := json.Unmarshal([]byte(unquoteJS(m[1])), &html); err != nil {
		return nil, fmt.Errorf("embed payload: %w", err)
	}

	// every <br> is followed by a newline
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(strings.ReplaceAll(html, "\n", "")))
	if err != nil {
		return nil, fmt.Errorf("embed html: %w", err)
	}
	body := doc.Find(".rg_embed_body")
	if body.Length() == 0 {
		return nil, errors.New("embed lyrics not found")
	}
	body.Find("br").ReplaceWithHtml("\n")

	var lines []lyrics.Line
	for l := range strings.SplitSeq(body.Text(), "\n") {
		lines = append(lines, lyrics.Line{Text: strings.TrimSpace(l)})
	}
	return lines, nil
}

// a \u escape is left for the json inside
func unquoteJS(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 == len(s) {
			b.WriteByte(s[i])
			continue
		}
		i++
		switch s[i] {
		case 'n':
			b.WriteByte('\n')
		case 'r':
			b.WriteByte('\r')
		case 't':
			b.WriteByte('\t')
		case 'u':
			b.WriteString(`\u`)
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}
