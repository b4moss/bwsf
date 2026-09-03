package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"
)

const (
	cloudIdentityUS = "https://identity.bitwarden.com"

	// Device metadata required by Vaultwarden (and harmless on Bitwarden cloud).
	// Numeric types match official clients / spike Scenario B.
	deviceName = "bwsf"
)

// DeviceInfo is sent on Identity /connect/token requests.
type DeviceInfo struct {
	Identifier string
	Name       string
	Type       string
}

// DefaultDeviceType returns a Bitwarden deviceType for the current OS.
func DefaultDeviceType() string {
	switch runtime.GOOS {
	case "darwin":
		return "7" // MacOsDesktop
	case "windows":
		return "6" // WindowsDesktop
	default:
		return "8" // LinuxDesktop
	}
}

// DefaultDeviceInfo builds device metadata with the given stable identifier.
func DefaultDeviceInfo(identifier string) DeviceInfo {
	return DeviceInfo{
		Identifier: identifier,
		Name:       deviceName,
		Type:       DefaultDeviceType(),
	}
}

// TokenSet holds OAuth tokens kept in process memory only.
type TokenSet struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresAt    time.Time
}

// Valid reports whether the access token is present and not expired (60s skew).
func (t *TokenSet) Valid() bool {
	if t == nil || t.AccessToken == "" {
		return false
	}
	if t.ExpiresAt.IsZero() {
		return true
	}
	return time.Now().Before(t.ExpiresAt.Add(-60 * time.Second))
}

// HTTPDoer is the subset of http.Client used by IdentityClient (for mocks).
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// IdentityClient talks to Bitwarden Identity /connect/token.
type IdentityClient struct {
	HTTPClient HTTPDoer
}

// NewIdentityClient creates an IdentityClient using http.DefaultClient.
func NewIdentityClient() *IdentityClient {
	return &IdentityClient{HTTPClient: http.DefaultClient}
}

type tokenJSON struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
}

// FetchClientCredentialsToken obtains an access token with a Personal API Key.
func (c *IdentityClient) FetchClientCredentialsToken(
	ctx context.Context,
	identityBase, clientID, clientSecret string,
	device DeviceInfo,
) (*TokenSet, error) {
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("scope", "api")
	applyDeviceFields(form, device)
	return c.postToken(ctx, identityBase, form)
}

// RefreshAccessToken refreshes an access token when a refresh_token is available.
func (c *IdentityClient) RefreshAccessToken(
	ctx context.Context,
	identityBase, refreshToken, clientID string,
	device DeviceInfo,
) (*TokenSet, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	if clientID != "" {
		form.Set("client_id", clientID)
	}
	applyDeviceFields(form, device)
	return c.postToken(ctx, identityBase, form)
}

func applyDeviceFields(form url.Values, device DeviceInfo) {
	if device.Identifier != "" {
		form.Set("deviceIdentifier", device.Identifier)
	}
	if device.Name != "" {
		form.Set("deviceName", device.Name)
	}
	if device.Type != "" {
		form.Set("deviceType", device.Type)
	}
}

func (c *IdentityClient) postToken(ctx context.Context, identityBase string, form url.Values) (*TokenSet, error) {
	if c == nil || c.HTTPClient == nil {
		return nil, fmt.Errorf("identity client is not configured")
	}
	tokenURL := strings.TrimRight(identityBase, "/") + "/connect/token"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build connect/token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST connect/token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read connect/token response: %w", err)
	}

	var parsed tokenJSON
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("decode connect/token JSON (HTTP %d): %w", resp.StatusCode, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := parsed.ErrorDesc
		if msg == "" {
			msg = parsed.Error
		}
		if msg == "" {
			msg = truncate(string(body), 300)
		}
		return nil, fmt.Errorf("connect/token HTTP %d: %s", resp.StatusCode, msg)
	}
	if parsed.AccessToken == "" {
		return nil, fmt.Errorf("connect/token response missing access_token")
	}

	token := &TokenSet{
		AccessToken:  parsed.AccessToken,
		RefreshToken: parsed.RefreshToken,
		TokenType:    parsed.TokenType,
	}
	if parsed.ExpiresIn > 0 {
		token.ExpiresAt = time.Now().Add(time.Duration(parsed.ExpiresIn) * time.Second)
	}
	return token, nil
}

// ResolveIdentityBase returns the Identity base URL for cloud or self-hosted.
func ResolveIdentityBase(hostType, selfhostedURL string) (string, error) {
	if hostType == "selfhosted" || strings.TrimSpace(selfhostedURL) != "" {
		normalized, err := normalizeServerURL(selfhostedURL)
		if err != nil {
			return "", err
		}
		return strings.TrimRight(normalized, "/") + "/identity", nil
	}
	return cloudIdentityUS, nil
}

func normalizeServerURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("self-hosted URL is required")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse server URL: %w", err)
	}
	if !u.IsAbs() || u.Host == "" {
		return "", fmt.Errorf("server URL must be absolute with host")
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return "", fmt.Errorf("server URL must use http or https")
	}
	u.Path = strings.TrimRight(u.Path, "/")
	u.RawQuery = ""
	u.Fragment = ""
	return u.Scheme + "://" + u.Host + u.Path, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
