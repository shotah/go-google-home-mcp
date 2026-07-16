package home

import (
	"errors"
	"fmt"
)

var (
	ErrNotAuthenticated = errors.New("google-home: not authenticated, run: ghome login")
	ErrSessionExpired   = errors.New("google-home: session expired, re-login required")
)

// APIError is a non-2xx SDM response.
type APIError struct {
	StatusCode int
	Status     string
	Body       []byte
}

func (e *APIError) Error() string {
	if len(e.Body) > 0 {
		return fmt.Sprintf("google-home: %s: %s", e.Status, string(e.Body))
	}
	return "google-home: " + e.Status
}
