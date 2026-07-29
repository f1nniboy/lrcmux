package providers

import (
	"errors"
	"net/http"
	"strconv"
	"time"
)

var ErrRateLimited = errors.New("provider rate limited")

type RateLimitedError struct {
	RetryAfter time.Duration
}

func (e *RateLimitedError) Error() string { return ErrRateLimited.Error() }
func (e *RateLimitedError) Unwrap() error { return ErrRateLimited }

// Retry-After is either an integer number of seconds or a HTTP-date
func ParseRetryAfter(v string) time.Duration {
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}

func RateLimitedFrom(resp *http.Response) error {
	if d := ParseRetryAfter(resp.Header.Get("Retry-After")); d > 0 {
		return &RateLimitedError{RetryAfter: d}
	}
	return ErrRateLimited
}
