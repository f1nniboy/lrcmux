package ratelimit

import "testing"

func TestClientKey(t *testing.T) {
	tests := []struct {
		a, b string
		same bool
	}{
		{"2001:db8:1:2::1", "2001:db8:1:2:ffff::9", true},
		{"2001:db8:1:2::1", "2001:db8:1:3::1", false},
		{"::ffff:192.0.2.1", "192.0.2.1", true},
		{"192.0.2.1", "192.0.2.2", false},
	}
	for _, tc := range tests {
		if same := clientKey(tc.a) == clientKey(tc.b); same != tc.same {
			t.Errorf("%s and %s: same key = %v, want %v", tc.a, tc.b, same, tc.same)
		}
	}
}
