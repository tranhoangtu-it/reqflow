package domain

// Collection represents a named group of HTTP requests, folders, and shared
// configuration such as variables, auth, and default headers.
type Collection struct {
	Name        string
	Description string
	Version     string
	Requests    []SavedRequest
	Folders     []Folder
	Variables   []Variable
	Auth        *AuthConfig
	Headers     []Header
}

// Folder represents a logical grouping of requests within a collection.
// Folders can be nested arbitrarily deep.
type Folder struct {
	Name        string
	Description string
	Requests    []SavedRequest
	Folders     []Folder
	Variables   []Variable
	Auth        *AuthConfig
	Headers     []Header
}

// SavedRequest represents a named HTTP request stored in a collection.
type SavedRequest struct {
	Name        string
	Description string
	Config      RequestConfig
}
