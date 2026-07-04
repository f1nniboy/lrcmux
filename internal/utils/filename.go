package utils

import (
	"strings"
)

// sanitizes the given string for use in a Content-Disposition filename
// parameter
func SanitizeFilename(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '"', '\\', '/', ':', '*', '?', '<', '>', '|':
			b.WriteByte('-')
		default:
			if r < 0x20 || r == 0x7f {
				continue
			}
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}
