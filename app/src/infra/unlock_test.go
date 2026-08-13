package infra

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"bwsf/src/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiBwClient_Unlock_SuccessAndClearSession(t *testing.T) {
	withTempHome(t)

	store := NewMemorySecretStore()
	_ = SaveAPICredentials(store, APICredentials{ClientID: "user.cid", ClientSecret: "sec"})
	identity := &IdentityClient{
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				payload := `{"access_token":"tok","expires_in":3600}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(payload)),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}
	crypto := &MockCryptoSession{}
	client := NewApiBwClientWithDeps(
		&config.Config{HostType: "cloud", Email: "a@example.com", Backend: config.BackendAPI},
		store,
		identity,
		crypto,
	)

	require.NoError(t, client.Authenticate(context.Background()))
	_, err := client.ListItemsInFolder("f")
	assert.ErrorIs(t, err, ErrAPINotUnlocked)

	require.NoError(t, client.Unlock("master-pass"))
	assert.True(t, client.IsUnlocked())
	assert.Contains(t, crypto.Calls, "UnlockWithPassword")

	crypto.Folders = []VaultFolder{{ID: "f1", Name: "dotenvs"}}
	id, err := client.GetDotenvsFolderID()
	require.NoError(t, err)
	assert.Equal(t, "f1", id)

	client.ClearSession()
	assert.False(t, client.IsAuthenticated())
	assert.False(t, client.IsUnlocked())
	assert.Contains(t, crypto.Calls, "Lock")

	_ = SaveAPICredentials(store, APICredentials{ClientID: "user.cid", ClientSecret: "sec"})
	require.NoError(t, client.Authenticate(context.Background()))
	assert.True(t, client.IsAuthenticated())
}

func TestApiBwClient_Unlock_EmptyPassword(t *testing.T) {
	client := NewApiBwClientWithDeps(
		&config.Config{HostType: "cloud", Email: "a@example.com"},
		NewMemorySecretStore(),
		NewIdentityClient(),
		&MockCryptoSession{},
	)
	err := client.Unlock("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestApiBwClient_Unlock_MissingEmail(t *testing.T) {
	withTempHome(t)
	store := NewMemorySecretStore()
	_ = SaveAPICredentials(store, APICredentials{ClientID: "user.cid", ClientSecret: "sec"})
	identity := &IdentityClient{
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				payload := `{"access_token":"tok","expires_in":3600}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(payload)),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}
	client := NewApiBwClientWithDeps(&config.Config{HostType: "cloud"}, store, identity, &MockCryptoSession{})
	require.NoError(t, client.Authenticate(context.Background()))
	err := client.Unlock("mp")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email is required")
}

func TestApiBwClient_Unlock_CryptoFailure(t *testing.T) {
	withTempHome(t)
	store := NewMemorySecretStore()
	_ = SaveAPICredentials(store, APICredentials{ClientID: "user.cid", ClientSecret: "sec"})
	identity := &IdentityClient{
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				payload := `{"access_token":"tok","expires_in":3600}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(payload)),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}
	crypto := &MockCryptoSession{UnlockErr: fmt.Errorf("bad password")}
	client := NewApiBwClientWithDeps(
		&config.Config{HostType: "cloud", Email: "a@example.com"},
		store,
		identity,
		crypto,
	)
	require.NoError(t, client.Authenticate(context.Background()))
	err := client.Unlock("wrong")
	assert.Error(t, err)
	assert.False(t, client.IsUnlocked())
}
