package cookiejar

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ye-kart/reqflow/internal/domain"
	"github.com/ye-kart/reqflow/internal/ports/driven"
)

// Compile-time check that Jar satisfies the driven.CookieJar interface.
var _ driven.CookieJar = (*Jar)(nil)

// Jar is a persistent cookie jar that stores cookies in memory and can
// serialize them to/from a JSON file.
type Jar struct {
	mu      sync.RWMutex
	cookies []domain.Cookie
}

// New creates a new empty Jar.
func New() *Jar {
	return &Jar{}
}

// SetCookies stores cookies associated with the given URL.
func (j *Jar) SetCookies(_ string, cookies []domain.Cookie) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	for _, c := range cookies {
		j.setLocked(c)
	}
	return nil
}

// setLocked adds or replaces a cookie. Must be called with j.mu held for writing.
func (j *Jar) setLocked(c domain.Cookie) {
	for i, existing := range j.cookies {
		if existing.Name == c.Name && existing.Domain == c.Domain && existing.Path == c.Path {
			j.cookies[i] = c
			return
		}
	}
	j.cookies = append(j.cookies, c)
}

// GetCookies returns non-expired cookies that match the given URL's
// domain and path.
func (j *Jar) GetCookies(rawURL string) ([]domain.Cookie, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	j.mu.RLock()
	defer j.mu.RUnlock()

	var result []domain.Cookie
	for _, c := range j.cookies {
		if c.IsExpired() {
			continue
		}
		if !domainMatches(c.Domain, u.Hostname()) {
			continue
		}
		if !pathMatches(c.Path, u.Path) {
			continue
		}
		result = append(result, c)
	}
	return result, nil
}

// Clear removes all stored cookies.
func (j *Jar) Clear() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.cookies = nil
	return nil
}

// ClearDomain removes all cookies for the given domain.
func (j *Jar) ClearDomain(targetDomain string) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	var kept []domain.Cookie
	for _, c := range j.cookies {
		if !domainMatches(c.Domain, targetDomain) {
			kept = append(kept, c)
		}
	}
	j.cookies = kept
	return nil
}

// All returns all stored, non-expired cookies.
func (j *Jar) All() ([]domain.Cookie, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	var result []domain.Cookie
	for _, c := range j.cookies {
		if !c.IsExpired() {
			result = append(result, c)
		}
	}
	return result, nil
}

// Load reads cookies from a JSON file at the given path.
// If the file does not exist, Load is a no-op and returns nil.
func (j *Jar) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var cookies []domain.Cookie
	if err := json.Unmarshal(data, &cookies); err != nil {
		return err
	}

	j.mu.Lock()
	defer j.mu.Unlock()

	// Only load non-expired cookies.
	for _, c := range cookies {
		if !c.IsExpired() {
			j.cookies = append(j.cookies, c)
		}
	}
	return nil
}

// Save writes all non-expired cookies to a JSON file at the given path.
// It creates parent directories as needed.
func (j *Jar) Save(path string) error {
	j.mu.RLock()
	var toSave []domain.Cookie
	for _, c := range j.cookies {
		if !c.IsExpired() {
			toSave = append(toSave, c)
		}
	}
	j.mu.RUnlock()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(toSave, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

// domainMatches checks if the cookie domain matches the request host.
func domainMatches(cookieDomain, host string) bool {
	cookieDomain = strings.ToLower(strings.TrimPrefix(cookieDomain, "."))
	host = strings.ToLower(host)

	if cookieDomain == host {
		return true
	}
	// Subdomain match: host ends with "."+cookieDomain
	return strings.HasSuffix(host, "."+cookieDomain)
}

// pathMatches checks if the request path starts with the cookie path.
func pathMatches(cookiePath, reqPath string) bool {
	if cookiePath == "" || cookiePath == "/" {
		return true
	}
	if reqPath == cookiePath {
		return true
	}
	return strings.HasPrefix(reqPath, cookiePath+"/") || strings.HasPrefix(reqPath, cookiePath)
}
