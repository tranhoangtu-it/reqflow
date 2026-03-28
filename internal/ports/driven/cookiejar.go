package driven

import "github.com/ye-kart/reqflow/internal/domain"

// CookieJar is the driven port for storing and retrieving HTTP cookies.
type CookieJar interface {
	// SetCookies stores cookies associated with the given URL.
	SetCookies(url string, cookies []domain.Cookie) error
	// GetCookies returns non-expired cookies that match the given URL.
	GetCookies(url string) ([]domain.Cookie, error)
	// Clear removes all stored cookies.
	Clear() error
	// ClearDomain removes all cookies for the given domain.
	ClearDomain(domain string) error
	// All returns all stored, non-expired cookies.
	All() ([]domain.Cookie, error)
}
