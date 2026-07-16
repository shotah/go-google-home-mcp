package home

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	sdmBaseURL = "https://smartdevicemanagement.googleapis.com/v1"
	oauthScope = "https://www.googleapis.com/auth/sdm.service"
)

// SessionPersister writes updated auth after a successful token refresh.
type SessionPersister func(*Client) error

// Options configures the Google Home / SDM client.
type Options struct {
	HTTPClient *http.Client
	// RedirectURL used during interactive OAuth (default http://127.0.0.1:8787/oauth/callback).
	RedirectURL string
}

// Client talks to the Nest Smart Device Management API.
type Client struct {
	opts             Options
	http             *http.Client
	session          *Session
	sessionPersister SessionPersister
}

// New creates an SDM client.
func New(opts Options) *Client {
	hc := opts.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: 30 * time.Second}
	}
	if opts.RedirectURL == "" {
		opts.RedirectURL = "http://127.0.0.1:8787/oauth/callback"
	}
	return &Client{
		opts:    opts,
		http:    hc,
		session: &Session{},
	}
}

// SetSessionPersister registers a callback after token refresh.
func (c *Client) SetSessionPersister(fn SessionPersister) {
	c.sessionPersister = fn
}

// SaveSession persists OAuth state.
func (c *Client) SaveSession(w io.Writer) error {
	return c.session.save(w)
}

// LoadSession restores OAuth state.
func (c *Client) LoadSession(r io.Reader) error {
	return c.session.load(r)
}

// ProjectID returns the Device Access project id.
func (c *Client) ProjectID() string {
	return c.session.ProjectID
}

// RedirectURL returns the OAuth redirect URI used for interactive login.
func (c *Client) RedirectURL() string {
	return c.opts.RedirectURL
}

// SetCredentials stores OAuth client + project used for login/refresh.
func (c *Client) SetCredentials(clientID, clientSecret, projectID string) {
	c.session.ClientID = clientID
	c.session.ClientSecret = clientSecret
	c.session.ProjectID = projectID
}

func (c *Client) persistSession() error {
	if c.sessionPersister == nil {
		return nil
	}
	return c.sessionPersister(c)
}

func (c *Client) ensureAuth(ctx context.Context) error {
	if !c.session.isAuthenticated() {
		return ErrNotAuthenticated
	}
	if c.session.isExpired() {
		return c.RefreshToken(ctx)
	}
	return nil
}

func (c *Client) enterprisePath() string {
	return "enterprises/" + c.session.ProjectID
}

func (c *Client) doJSON(ctx context.Context, method, path string, body any, out any) error {
	if err := c.ensureAuth(ctx); err != nil {
		return err
	}

	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	resp, err := c.doOnce(ctx, method, path, bodyBytes)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		_, _ = io.Copy(io.Discard, resp.Body)
		if err := c.RefreshToken(ctx); err != nil {
			return err
		}
		resp, err = c.doOnce(ctx, method, path, bodyBytes)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{StatusCode: resp.StatusCode, Status: resp.Status, Body: raw}
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

func (c *Client) doOnce(ctx context.Context, method, path string, bodyBytes []byte) (*http.Response, error) {
	url := path
	if !strings.HasPrefix(path, "http") {
		url = sdmBaseURL + "/" + strings.TrimPrefix(path, "/")
	}
	var bodyReader io.Reader = http.NoBody
	if bodyBytes != nil {
		bodyReader = bytes.NewReader(bodyBytes)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.session.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	return c.http.Do(req)
}

// ListDevices returns all devices linked to the Device Access project.
func (c *Client) ListDevices(ctx context.Context) ([]Device, error) {
	var resp listDevicesResponse
	path := c.enterprisePath() + "/devices"
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Devices, nil
}

// GetDevice fetches a single device by resource name or bare device id.
func (c *Client) GetDevice(ctx context.Context, deviceName string) (*Device, error) {
	path := normalizeDeviceName(c.session.ProjectID, deviceName)
	var dev Device
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &dev); err != nil {
		return nil, err
	}
	return &dev, nil
}

// ListStructures returns authorized structures (homes).
func (c *Client) ListStructures(ctx context.Context) ([]Structure, error) {
	var resp listStructuresResponse
	path := c.enterprisePath() + "/structures"
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Structures, nil
}

// ExecuteCommand runs an SDM trait command on a device.
func (c *Client) ExecuteCommand(ctx context.Context, deviceName, command string, params map[string]any) error {
	path := normalizeDeviceName(c.session.ProjectID, deviceName) + ":executeCommand"
	req := ExecuteCommandRequest{Command: command, Params: params}
	return c.doJSON(ctx, http.MethodPost, path, req, nil)
}

func normalizeDeviceName(projectID, deviceName string) string {
	if strings.HasPrefix(deviceName, "enterprises/") {
		return deviceName
	}
	id := strings.TrimPrefix(deviceName, "devices/")
	return fmt.Sprintf("enterprises/%s/devices/%s", projectID, id)
}

// SummarizeDevice builds a compact listing row.
func SummarizeDevice(d Device) DeviceSummary {
	sum := DeviceSummary{
		Name: d.Name,
		Type: strings.TrimPrefix(d.Type, "sdm.devices.types."),
		ID:   deviceID(d.Name),
	}
	if len(d.ParentRelations) > 0 {
		sum.Room = d.ParentRelations[0].DisplayName
	}
	if info, ok := d.Traits["sdm.devices.traits.Info"].(map[string]any); ok {
		if n, ok := info["customName"].(string); ok {
			sum.CustomName = n
		}
	}
	if conn, ok := d.Traits["sdm.devices.traits.Connectivity"].(map[string]any); ok {
		if s, ok := conn["status"].(string); ok {
			sum.Online = s
		}
	}
	if temp, ok := d.Traits["sdm.devices.traits.Temperature"].(map[string]any); ok {
		if v, ok := temp["ambientTemperatureCelsius"].(float64); ok {
			sum.Temperature = v
		}
	}
	if mode, ok := d.Traits["sdm.devices.traits.ThermostatMode"].(map[string]any); ok {
		if m, ok := mode["mode"].(string); ok {
			sum.Mode = m
		}
	}
	return sum
}

func deviceID(name string) string {
	parts := strings.Split(name, "/")
	if len(parts) == 0 {
		return name
	}
	return parts[len(parts)-1]
}
