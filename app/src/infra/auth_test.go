package infra

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"bwsf/src/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemorySecretStore_APICredentials(t *testing.T) {
	store := NewMemorySecretStore()

	_, err := LoadAPICredentials(store)
	assert.ErrorIs(t, err, ErrSecretNotFound)

	err = SaveAPICredentials(store, APICredentials{
		ClientID:     "user.abc",
		ClientSecret: "secret-value",
	})
	require.NoError(t, err)

	creds, err := LoadAPICredentials(store)
	require.NoError(t, err)
	assert.Equal(t, "user.abc", creds.ClientID)
	assert.Equal(t, "secret-value", creds.ClientSecret)

	require.NoError(t, ClearAPICredentials(store))
	_, err = LoadAPICredentials(store)
	assert.ErrorIs(t, err, ErrSecretNotFound)
}

func TestResolveIdentityBase_Cloud(t *testing.T) {
	base, err := ResolveIdentityBase(config.HostTypeCloud, config.DefaultCloudURL)
	require.NoError(t, err)
	assert.Equal(t, cloudIdentityUS, base)

	base, err = ResolveIdentityBase("cloud", "")
	require.NoError(t, err)
	assert.Equal(t, cloudIdentityUS, base)
}

func TestResolveIdentityBase_SelfHosted(t *testing.T) {
	base, err := ResolveIdentityBase(config.HostTypeSelfhost, "https://vw.example.com/")
	require.NoError(t, err)
	assert.Equal(t, "https://vw.example.com/identity", base)

	base, err = ResolveIdentityBase("selfhosted", "https://vw.example.com/")
	require.NoError(t, err)
	assert.Equal(t, "https://vw.example.com/identity", base)
}

func TestTokenSet_Valid(t *testing.T) {
	assert.False(t, (*TokenSet)(nil).Valid())
	assert.False(t, (&TokenSet{}).Valid())
	assert.True(t, (&TokenSet{AccessToken: "t"}).Valid())
	assert.False(t, (&TokenSet{
		AccessToken: "t",
		ExpiresAt:   time.Now().Add(-time.Minute),
	}).Valid())
	assert.True(t, (&TokenSet{
		AccessToken: "t",
		ExpiresAt:   time.Now().Add(10 * time.Minute),
	}).Valid())
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestIdentityClient_FetchClientCredentialsToken(t *testing.T) {
	var gotBody string
	client := &IdentityClient{
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, http.MethodPost, req.Method)
				assert.True(t, strings.HasSuffix(req.URL.Path, "/connect/token"))
				body, err := io.ReadAll(req.Body)
				require.NoError(t, err)
				gotBody = string(body)

				payload, _ := json.Marshal(map[string]interface{}{
					"access_token": "tok-123",
					"token_type":   "Bearer",
					"expires_in":   3600,
				})
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(string(payload))),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}

	token, err := client.FetchClientCredentialsToken(
		context.Background(),
		"https://identity.example.com",
		"user.cid",
		"csecret",
		DeviceInfo{Identifier: "dev-1", Name: "bwsf", Type: "8"},
	)
	require.NoError(t, err)
	assert.Equal(t, "tok-123", token.AccessToken)
	assert.Contains(t, gotBody, "grant_type=client_credentials")
	assert.Contains(t, gotBody, "client_id=user.cid")
	assert.Contains(t, gotBody, "deviceIdentifier=dev-1")
	assert.Contains(t, gotBody, "deviceName=bwsf")
	assert.Contains(t, gotBody, "deviceType=8")
}

func TestIdentityClient_FetchClientCredentialsToken_Error(t *testing.T) {
	client := &IdentityClient{
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				payload := `{"error":"invalid_client","error_description":"bad key"}`
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(strings.NewReader(payload)),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}

	_, err := client.FetchClientCredentialsToken(
		context.Background(),
		"https://identity.example.com",
		"bad",
		"bad",
		DefaultDeviceInfo("dev"),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 400")
	assert.Contains(t, err.Error(), "bad key")
}

func TestIdentityClient_RefreshAccessToken(t *testing.T) {
	client := &IdentityClient{
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				body, _ := io.ReadAll(req.Body)
				assert.Contains(t, string(body), "grant_type=refresh_token")
				payload := `{"access_token":"new-tok","refresh_token":"new-rt","expires_in":1800}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(payload)),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}

	token, err := client.RefreshAccessToken(
		context.Background(),
		"https://identity.example.com",
		"old-rt",
		"user.cid",
		DefaultDeviceInfo("dev"),
	)
	require.NoError(t, err)
	assert.Equal(t, "new-tok", token.AccessToken)
	assert.Equal(t, "new-rt", token.RefreshToken)
}

func TestApiBwClient_AuthenticateAndLogin(t *testing.T) {
	withTempHome(t)

	store := NewMemorySecretStore()
	require.NoError(t, SaveAPICredentials(store, APICredentials{
		ClientID:     "user.cid",
		ClientSecret: "csecret",
	}))

	identity := &IdentityClient{
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				payload := `{"access_token":"mem-tok","token_type":"Bearer","expires_in":3600}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(payload)),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}

	cfg := testConfig("cloud", "", "a@example.com", "")
	client := NewApiBwClientWithDeps(cfg, store, identity, nil)

	require.NoError(t, client.Authenticate(context.Background()))
	assert.True(t, client.IsAuthenticated())

	tok, err := client.EnsureAccessToken(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "mem-tok", tok)

	client.ClearSession()
	assert.False(t, client.IsAuthenticated())
	require.NoError(t, client.Login("ignored@example.com", "ignored", ""))
	assert.True(t, client.IsAuthenticated())
}

func TestApiBwClient_Authenticate_MissingCredentials(t *testing.T) {
	client := NewApiBwClientWithDeps(testConfig("cloud", "", "", ""), NewMemorySecretStore(), NewIdentityClient(), nil)
	err := client.Authenticate(context.Background())
	assert.ErrorIs(t, err, ErrAPINotAuthenticated)
}

func TestApiBwClient_EnsureAccessToken_Refresh(t *testing.T) {
	withTempHome(t)

	calls := 0
	identity := &IdentityClient{
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				calls++
				body, _ := io.ReadAll(req.Body)
				if strings.Contains(string(body), "grant_type=refresh_token") {
					payload := `{"access_token":"refreshed","refresh_token":"rt2","expires_in":3600}`
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(payload)),
						Header:     make(http.Header),
					}, nil
				}
				t.Fatalf("unexpected body: %s", body)
				return nil, nil
			}),
		},
	}

	store := NewMemorySecretStore()
	_ = SaveAPICredentials(store, APICredentials{ClientID: "user.cid", ClientSecret: "sec"})
	client := NewApiBwClientWithDeps(testConfig("cloud", "", "", ""), store, identity, nil)
	client.mu.Lock()
	client.token = &TokenSet{
		AccessToken:  "old",
		RefreshToken: "rt1",
		ExpiresAt:    time.Now().Add(-time.Minute),
	}
	client.mu.Unlock()

	tok, err := client.EnsureAccessToken(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "refreshed", tok)
	assert.Equal(t, 1, calls)
}

func withTempHome(t *testing.T) {
	t.Helper()
	orig := os.Getenv("HOME")
	tmp := t.TempDir()
	require.NoError(t, os.Setenv("HOME", tmp))
	t.Cleanup(func() {
		_ = os.Setenv("HOME", orig)
	})
}
