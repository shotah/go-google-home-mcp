package home

import (
	"encoding/json"
	"io"
	"time"
)

// Session is the persisted OAuth + Device Access project state.
type Session struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
	TokenType    string    `json:"token_type,omitempty"`
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret"`
	ProjectID    string    `json:"project_id"`
}

func (s *Session) save(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}

func (s *Session) load(r io.Reader) error {
	return json.NewDecoder(r).Decode(s)
}

func (s *Session) isExpired() bool {
	if s.Expiry.IsZero() {
		return true
	}
	return time.Now().Add(5 * time.Minute).After(s.Expiry)
}

func (s *Session) isAuthenticated() bool {
	return s.AccessToken != "" && s.ProjectID != ""
}
