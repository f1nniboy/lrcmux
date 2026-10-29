package musixmatch

import (
	"testing"
	"time"
)

func TestSign(t *testing.T) {
	const raw = "https://www.musixmatch.com/ws/1.1/track.lyrics.get?app_id=community-app-v1.0&format=json&track_isrc=XX0000000000"
	// reference from the implementation verified against the live api
	const want = raw + "&signature=ag%2BtwtauIB3A8iJGO8xnjA8umuKJ%2BC45cdrX9jPMDOM%3D&signature_protocol=sha256"

	// 01:00 on jan 3 at utc+2 is still jan 2 in utc, which is what gets signed
	now := time.Date(2026, 1, 3, 1, 0, 0, 0, time.FixedZone("", 2*60*60))
	if got := sign(raw, "0123456789abcdef0123456789abcdef", now); got != want {
		t.Errorf("sign() =\n%s\nwant\n%s", got, want)
	}
}

func TestExtractKey(t *testing.T) {
	js := `a=Buffer.from("not a key".split(""));` +
		`b=Buffer.from("=YWZkNmYhlDO3YTN0MjMxAjZlR2YiFWO4cjN1QzMyEDM".split("").reverse().join(""),"base64")`
	if key, ok := extractKey(js); !ok || key != "0123456789abcdef0123456789abcdef" {
		t.Errorf("extractKey() = %q, %v", key, ok)
	}
	if _, ok := extractKey(`c=Buffer.from("x".split(""))`); ok {
		t.Error("extractKey() found a key where there is none")
	}
}
