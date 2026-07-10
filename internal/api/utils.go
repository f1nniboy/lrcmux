package api

import (
	"net"
	"strings"
)

func isPrivateIP(ip string) bool {
	p := net.ParseIP(ip)
	return p != nil && (p.IsLoopback() || p.IsPrivate())
}

func sanitizeFilename(s string) string {
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
