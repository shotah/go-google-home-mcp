package home

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const nestPartnerAuthBase = "https://nestservices.google.com/partnerconnections"

// AuthCodeURL builds the Nest Device Access partner OAuth URL.
func (c *Client) AuthCodeURL(state string) (string, error) {
	if c.session.ClientID == "" || c.session.ProjectID == "" {
		return "", fmt.Errorf("google-home: client_id and project_id are required")
	}
	v := url.Values{}
	v.Set("access_type", "offline")
	v.Set("prompt", "consent")
	v.Set("client_id", c.session.ClientID)
	v.Set("response_type", "code")
	v.Set("scope", oauthScope)
	v.Set("redirect_uri", c.opts.RedirectURL)
	if state != "" {
		v.Set("state", state)
	}
	return fmt.Sprintf("%s/%s/auth?%s", nestPartnerAuthBase, c.session.ProjectID, v.Encode()), nil
}

func (c *Client) oauth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     c.session.ClientID,
		ClientSecret: c.session.ClientSecret,
		RedirectURL:  c.opts.RedirectURL,
		Scopes:       []string{oauthScope},
		Endpoint:     google.Endpoint,
	}
}

// ExchangeCode swaps an authorization code for tokens and persists them.
func (c *Client) ExchangeCode(ctx context.Context, code string) error {
	cfg := c.oauth2Config()
	tok, err := cfg.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("google-home: token exchange failed: %w", err)
	}
	return c.applyToken(tok)
}

// RefreshToken refreshes the access token and persists the session.
func (c *Client) RefreshToken(ctx context.Context) error {
	if c.session.RefreshToken == "" {
		return fmt.Errorf("%w: missing refresh token", ErrSessionExpired)
	}
	cfg := c.oauth2Config()
	src := cfg.TokenSource(ctx, &oauth2.Token{
		RefreshToken: c.session.RefreshToken,
		Expiry:       c.session.Expiry,
	})
	tok, err := src.Token()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSessionExpired, err)
	}
	if tok.RefreshToken == "" {
		tok.RefreshToken = c.session.RefreshToken
	}
	return c.applyToken(tok)
}

func (c *Client) applyToken(tok *oauth2.Token) error {
	c.session.AccessToken = tok.AccessToken
	if tok.RefreshToken != "" {
		c.session.RefreshToken = tok.RefreshToken
	}
	c.session.Expiry = tok.Expiry
	c.session.TokenType = tok.TokenType
	if err := c.persistSession(); err != nil {
		return fmt.Errorf("token updated but failed to persist session: %w", err)
	}
	return nil
}

// LoginInteractive opens a local callback server and waits for the OAuth redirect.
// The caller should print AuthCodeURL to the user so they can open it in a browser.
func (c *Client) LoginInteractive(ctx context.Context) (authURL string, wait func() error, err error) {
	state := fmt.Sprintf("ghome-%d", time.Now().UnixNano())
	authURL, err = c.AuthCodeURL(state)
	if err != nil {
		return "", nil, err
	}

	u, err := url.Parse(c.opts.RedirectURL)
	if err != nil {
		return "", nil, err
	}
	addr := u.Host
	if addr == "" {
		addr = "127.0.0.1:8787"
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc(u.Path, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("error") != "" {
			errCh <- fmt.Errorf("oauth error: %s", r.URL.Query().Get("error"))
			_, _ = io.WriteString(w, "Authorization failed. You can close this tab.")
			return
		}
		if r.URL.Query().Get("state") != state {
			errCh <- fmt.Errorf("oauth state mismatch")
			_, _ = io.WriteString(w, "State mismatch. You can close this tab.")
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("oauth missing code")
			_, _ = io.WriteString(w, "Missing code. You can close this tab.")
			return
		}
		codeCh <- code
		_, _ = io.WriteString(w, "Google Home authorized. You can close this tab and return to the terminal.")
	})

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return "", nil, fmt.Errorf("listen on %s: %w", addr, err)
	}
	srv := &http.Server{Handler: mux}

	go func() { _ = srv.Serve(ln) }()

	wait = func() error {
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = srv.Shutdown(shutdownCtx)
		}()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			return err
		case code := <-codeCh:
			return c.ExchangeCode(ctx, code)
		}
	}
	return authURL, wait, nil
}

// ExchangeCodeManual is for paste-the-code flows when a browser redirect is unavailable.
func (c *Client) ExchangeCodeManual(ctx context.Context, code string) error {
	return c.ExchangeCode(ctx, strings.TrimSpace(code))
}

// DebugTokenJSON returns a redacted session snapshot for troubleshooting.
func (c *Client) DebugTokenJSON() ([]byte, error) {
	type view struct {
		HasAccessToken  bool      `json:"has_access_token"`
		HasRefreshToken bool      `json:"has_refresh_token"`
		Expiry          time.Time `json:"expiry"`
		ProjectID       string    `json:"project_id"`
		ClientID        string    `json:"client_id"`
	}
	return json.MarshalIndent(view{
		HasAccessToken:  c.session.AccessToken != "",
		HasRefreshToken: c.session.RefreshToken != "",
		Expiry:          c.session.Expiry,
		ProjectID:       c.session.ProjectID,
		ClientID:        c.session.ClientID,
	}, "", "  ")
}
