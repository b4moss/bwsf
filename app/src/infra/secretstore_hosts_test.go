package infra

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"bwsf/src/config"

	"github.com/bnema/bitwarden-go-sdk/bitwarden"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHostScopedAPICredentials(t *testing.T) {
	store := NewMemorySecretStore()
	require.NoError(t, SaveAPICredentials(store, "work", APICredentials{ClientID: "w.id", ClientSecret: "w.sec"}))
	assert.True(t, store.Has("hosts/work/api_client_id"))
	assert.True(t, store.Has("hosts/work/api_client_secret"))
	assert.False(t, store.Has(keyAPIClientID))

	creds, err := LoadAPICredentials(store, "work")
	require.NoError(t, err)
	assert.Equal(t, "w.id", creds.ClientID)

	require.NoError(t, SaveAPICredentials(store, "other", APICredentials{ClientID: "o.id", ClientSecret: "o.sec"}))
	require.NoError(t, SetVaultUnlock(store, "work", "keep-me"))
	require.NoError(t, ClearAPICredentials(store, "work"))
	_, err = LoadAPICredentials(store, "work")
	assert.ErrorIs(t, err, ErrSecretNotFound)
	blob, err := GetVaultUnlock(store, "work")
	require.NoError(t, err)
	assert.Equal(t, "keep-me", blob)
	creds, err = LoadAPICredentials(store, "other")
	require.NoError(t, err)
	assert.Equal(t, "o.id", creds.ClientID)
}

func TestFlatKeyMigrationForDefaultOnly(t *testing.T) {
	store := NewMemorySecretStore()
	require.NoError(t, store.Set(keyAPIClientID, "flat.id"))
	require.NoError(t, store.Set(keyAPIClientSecret, "flat.sec"))

	creds, err := LoadAPICredentials(store, config.DefaultHostID)
	require.NoError(t, err)
	assert.Equal(t, "flat.id", creds.ClientID)
	assert.True(t, store.Has("hosts/default/api_client_id"))
	assert.False(t, store.Has(keyAPIClientID))

	store2 := NewMemorySecretStore()
	require.NoError(t, store2.Set(keyAPIClientID, "flat.id"))
	require.NoError(t, store2.Set(keyAPIClientSecret, "flat.sec"))
	_, err = LoadAPICredentials(store2, "work")
	assert.ErrorIs(t, err, ErrSecretNotFound)
	assert.True(t, store2.Has(keyAPIClientID))
}

func TestFlatIncompletePair(t *testing.T) {
	store := NewMemorySecretStore()
	require.NoError(t, store.Set(keyAPIClientID, "flat.id"))
	_, err := LoadAPICredentials(store, config.DefaultHostID)
	assert.ErrorIs(t, err, ErrSecretNotFound)
	assert.False(t, store.Has("hosts/default/api_client_id"))
}

func TestVaultUnlockCRUD(t *testing.T) {
	store := NewMemorySecretStore()
	require.Error(t, SetVaultUnlock(store, "work", ""))
	require.NoError(t, SetVaultUnlock(store, "work", "v1:blob"))
	got, err := GetVaultUnlock(store, "work")
	require.NoError(t, err)
	assert.Equal(t, "v1:blob", got)

	require.NoError(t, SaveAPICredentials(store, "work", APICredentials{ClientID: "id", ClientSecret: "sec"}))
	require.NoError(t, DeleteVaultUnlock(store, "work"))
	_, err = GetVaultUnlock(store, "work")
	assert.ErrorIs(t, err, ErrSecretNotFound)
	_, err = LoadAPICredentials(store, "work")
	require.NoError(t, err)
}

func TestHostIDRequired(t *testing.T) {
	store := NewMemorySecretStore()
	require.Error(t, SaveAPICredentials(store, "", APICredentials{ClientID: "a", ClientSecret: "b"}))
	_, err := LoadAPICredentials(store, "  ")
	require.Error(t, err)
	require.Error(t, ClearAPICredentials(store, ""))
}

func TestUnlockBlobRoundTrip(t *testing.T) {
	blob, err := encodeUnlockBlob(bitwarden.SessionMaterial{
		AccountID: "acct-1",
		UserKey:   []byte("user-key-bytes"),
		Tokens: bitwarden.TokenSet{
			AccountID:   "acct-1",
			AccessToken: []byte("access"),
			TokenType:   "Bearer",
		},
	})
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(blob, "v1:"))
	material, err := decodeUnlockBlob(blob)
	require.NoError(t, err)
	assert.Equal(t, "acct-1", material.AccountID)
	assert.Equal(t, []byte("user-key-bytes"), material.UserKey)
}

func TestUnlockBlobInvalid(t *testing.T) {
	_, err := decodeUnlockBlob("")
	assert.Error(t, err)
	_, err = decodeUnlockBlob("v99:abc")
	assert.Error(t, err)
	_, err = decodeUnlockBlob("not-a-blob")
	assert.Error(t, err)
}

func TestApiBwClient_UnlockPersistsVaultUnlockAndClearSessionKeepsIt(t *testing.T) {
	withTempHome(t)
	cfg := testConfig("cloud", "", "a@example.com", "")
	store := NewMemorySecretStore()
	require.NoError(t, SaveAPICredentials(store, cfg.DefaultHost().ID, APICredentials{ClientID: "user.cid", ClientSecret: "sec"}))

	crypto := &MockCryptoSession{ExportBlob: "v1:opaque"}
	client := NewApiBwClientWithDepsForHost(cfg, cfg.DefaultHost(), store, &IdentityClient{
		HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"access_token":"tok","expires_in":3600}`)), Header: make(http.Header)}, nil
		})},
	}, crypto)

	require.NoError(t, client.Unlock("master-pass"))
	blob, err := GetVaultUnlock(store, cfg.DefaultHost().ID)
	require.NoError(t, err)
	assert.Equal(t, "v1:opaque", blob)
	assert.NotContains(t, blob, "master-pass")

	client.ClearSession()
	assert.False(t, client.IsUnlocked())
	blob, err = GetVaultUnlock(store, cfg.DefaultHost().ID)
	require.NoError(t, err)
	assert.Equal(t, "v1:opaque", blob)

	require.NoError(t, client.LockVaultSession())
	_, err = GetVaultUnlock(store, cfg.DefaultHost().ID)
	assert.ErrorIs(t, err, ErrSecretNotFound)
	_, err = LoadAPICredentials(store, cfg.DefaultHost().ID)
	require.NoError(t, err)
}

func TestApiBwClient_TryRestoreVaultUnlock(t *testing.T) {
	withTempHome(t)
	cfg := testConfig("cloud", "", "a@example.com", "")
	store := NewMemorySecretStore()
	require.NoError(t, SetVaultUnlock(store, cfg.DefaultHost().ID, "good-blob"))

	crypto := &MockCryptoSession{}
	client := NewApiBwClientWithDepsForHost(cfg, cfg.DefaultHost(), store, NewIdentityClient(), crypto)
	ok, err := client.TryRestoreVaultUnlock()
	require.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, client.IsUnlocked())
	assert.Contains(t, crypto.Calls, "RestoreUnlockBlob")
}

func TestApiBwClient_TryRestoreInvalidDiscards(t *testing.T) {
	withTempHome(t)
	cfg := testConfig("cloud", "", "a@example.com", "")
	store := NewMemorySecretStore()
	require.NoError(t, SetVaultUnlock(store, cfg.DefaultHost().ID, "bad"))

	crypto := &MockCryptoSession{RestoreErr: fmt.Errorf("corrupt")}
	client := NewApiBwClientWithDepsForHost(cfg, cfg.DefaultHost(), store, NewIdentityClient(), crypto)
	ok, err := client.TryRestoreVaultUnlock()
	require.NoError(t, err)
	assert.False(t, ok)
	_, err = GetVaultUnlock(store, cfg.DefaultHost().ID)
	assert.ErrorIs(t, err, ErrSecretNotFound)
}

func TestApiBwClient_UnlockFailureKeepsExistingBlob(t *testing.T) {
	withTempHome(t)
	cfg := testConfig("cloud", "", "a@example.com", "")
	store := NewMemorySecretStore()
	require.NoError(t, SaveAPICredentials(store, cfg.DefaultHost().ID, APICredentials{ClientID: "user.cid", ClientSecret: "sec"}))
	require.NoError(t, SetVaultUnlock(store, cfg.DefaultHost().ID, "existing"))

	crypto := &MockCryptoSession{UnlockErr: fmt.Errorf("bad password")}
	client := NewApiBwClientWithDepsForHost(cfg, cfg.DefaultHost(), store, &IdentityClient{
		HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"access_token":"tok","expires_in":3600}`)), Header: make(http.Header)}, nil
		})},
	}, crypto)
	err := client.Unlock("wrong")
	assert.Error(t, err)
	blob, err := GetVaultUnlock(store, cfg.DefaultHost().ID)
	require.NoError(t, err)
	assert.Equal(t, "existing", blob)
}
