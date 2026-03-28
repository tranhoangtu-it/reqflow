package domain

import "time"

// Cookie represents an HTTP cookie with metadata for domain/path matching.
type Cookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Domain   string    `json:"domain"`
	Path     string    `json:"path"`
	Expires  time.Time `json:"expires"`
	Secure   bool      `json:"secure"`
	HTTPOnly bool      `json:"http_only"`
}

// IsExpired reports whether the cookie has expired.
// A zero Expires time means the cookie has no expiry (session cookie).
func (c Cookie) IsExpired() bool {
	if c.Expires.IsZero() {
		return false
	}
	return time.Now().After(c.Expires)
}
