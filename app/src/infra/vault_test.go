package infra

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"bwsf/src/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newUnlockedAPIClient(t *testing.T, cfg *config.Config, crypto *MockCryptoSession) *ApiBwClient {
	t.Helper()
	withTempHome(t)
	if cfg == nil {
		cfg = testConfig("cloud", "", "a@example.com", "")
	}
	store := NewMemorySecretStore()
	_ = SaveAPICredentials(store, config.DefaultHostID, APICredentials{ClientID: "user.cid", ClientSecret: "sec"})
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
	client := NewApiBwClientWithDeps(cfg, store, identity, crypto)
	require.NoError(t, client.Authenticate(context.Background()))
	require.NoError(t, client.Unlock("master-pass"))
	return client
}

func TestApiBwClient_GetDotenvsFolderID_Success(t *testing.T) {
	crypto := &MockCryptoSession{
		Folders: []VaultFolder{{ID: "f1", Name: "dotenvs"}},
	}
	client := newUnlockedAPIClient(t, nil, crypto)

	id, err := client.GetDotenvsFolderID()
	require.NoError(t, err)
	assert.Equal(t, "f1", id)
	assert.Contains(t, crypto.Calls, "Sync")
}

func TestApiBwClient_GetDotenvsFolderID_CustomFolderName(t *testing.T) {
	crypto := &MockCryptoSession{
		Folders: []VaultFolder{{ID: "f9", Name: "envnotes"}},
	}
	client := newUnlockedAPIClient(t, testConfig("cloud", "", "a@example.com", "envnotes"), crypto)

	id, err := client.GetDotenvsFolderID()
	require.NoError(t, err)
	assert.Equal(t, "f9", id)
}

func TestApiBwClient_GetDotenvsFolderID_NotFound(t *testing.T) {
	crypto := &MockCryptoSession{}
	client := newUnlockedAPIClient(t, nil, crypto)
	_, err := client.GetDotenvsFolderID()
	assert.ErrorIs(t, err, ErrAPIFolderNotFound)
}

func TestApiBwClient_GetDotenvsFolderID_Duplicate(t *testing.T) {
	crypto := &MockCryptoSession{
		Folders: []VaultFolder{
			{ID: "f1", Name: "dotenvs"},
			{ID: "f2", Name: "dotenvs"},
		},
	}
	client := newUnlockedAPIClient(t, nil, crypto)
	_, err := client.GetDotenvsFolderID()
	assert.ErrorIs(t, err, ErrAPIDuplicateFolder)
}

func TestApiBwClient_DotenvsFolderExists(t *testing.T) {
	crypto := &MockCryptoSession{Folders: []VaultFolder{{ID: "f1", Name: "dotenvs"}}}
	client := newUnlockedAPIClient(t, nil, crypto)
	ok, err := client.DotenvsFolderExists()
	require.NoError(t, err)
	assert.True(t, ok)

	crypto2 := &MockCryptoSession{}
	client2 := newUnlockedAPIClient(t, nil, crypto2)
	ok, err = client2.DotenvsFolderExists()
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestApiBwClient_CreateDotenvsFolder(t *testing.T) {
	crypto := &MockCryptoSession{}
	client := newUnlockedAPIClient(t, nil, crypto)
	require.NoError(t, client.CreateDotenvsFolder())
	ok, err := client.DotenvsFolderExists()
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Contains(t, crypto.Calls, "CreateFolder:dotenvs")
}

func TestApiBwClient_CreateDotenvsFolder_AlreadyExists(t *testing.T) {
	crypto := &MockCryptoSession{Folders: []VaultFolder{{ID: "f1", Name: "dotenvs"}}}
	client := newUnlockedAPIClient(t, nil, crypto)
	err := client.CreateDotenvsFolder()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestApiBwClient_ListItemsInFolder_SecureNotesOnly(t *testing.T) {
	crypto := &MockCryptoSession{
		Items: []VaultItem{
			{ID: "n1", Name: "proj", FolderID: "f1", Type: "secure_note", Notes: "{}"},
			{ID: "l1", Name: "login", FolderID: "f1", Type: "login"},
			{ID: "n2", Name: "other", FolderID: "f2", Type: "secure_note"},
			{ID: "n3", Name: "gone", FolderID: "f1", Type: "secure_note", Deleted: true},
		},
	}
	client := newUnlockedAPIClient(t, nil, crypto)
	items, err := client.ListItemsInFolder("f1")
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "proj", items[0].Name)
}

func TestApiBwClient_GetItemByName_Unique(t *testing.T) {
	crypto := &MockCryptoSession{
		Items: []VaultItem{
			{ID: "n1", Name: "proj", FolderID: "f1", Type: "secure_note", Notes: `{"a":1}`},
		},
	}
	client := newUnlockedAPIClient(t, nil, crypto)
	item, err := client.GetItemByName("f1", "proj")
	require.NoError(t, err)
	require.NotNil(t, item)
	assert.Equal(t, "n1", item.ID)
	assert.Equal(t, `{"a":1}`, item.Notes)
}

func TestApiBwClient_GetItemByName_Missing(t *testing.T) {
	crypto := &MockCryptoSession{}
	client := newUnlockedAPIClient(t, nil, crypto)
	item, err := client.GetItemByName("f1", "missing")
	require.NoError(t, err)
	assert.Nil(t, item)
}

func TestApiBwClient_GetItemByName_Duplicate(t *testing.T) {
	crypto := &MockCryptoSession{
		Items: []VaultItem{
			{ID: "n1", Name: "proj", FolderID: "f1", Type: "secure_note"},
			{ID: "n2", Name: "proj", FolderID: "f1", Type: "secure_note"},
		},
	}
	client := newUnlockedAPIClient(t, nil, crypto)
	_, err := client.GetItemByName("f1", "proj")
	assert.ErrorIs(t, err, ErrAPIDuplicateNote)
}

func TestApiBwClient_CreateAndUpdateNoteItem(t *testing.T) {
	crypto := &MockCryptoSession{}
	client := newUnlockedAPIClient(t, nil, crypto)
	require.NoError(t, client.CreateNoteItem("f1", "proj", `{"v":1}`))
	item, err := client.GetItemByName("f1", "proj")
	require.NoError(t, err)
	require.NotNil(t, item)
	require.NoError(t, client.UpdateNoteItem(item.ID, `{"v":2}`))
	got, err := client.GetItemByID(item.ID)
	require.NoError(t, err)
	assert.Equal(t, `{"v":2}`, got.Notes)
}

func TestApiBwClient_VaultRequiresUnlock(t *testing.T) {
	withTempHome(t)
	store := NewMemorySecretStore()
	_ = SaveAPICredentials(store, config.DefaultHostID, APICredentials{ClientID: "user.cid", ClientSecret: "sec"})
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
		testConfig("cloud", "", "a@example.com", ""),
		store,
		identity,
		crypto,
	)
	require.NoError(t, client.Authenticate(context.Background()))
	_, err := client.GetDotenvsFolderID()
	assert.ErrorIs(t, err, ErrAPINotUnlocked)
}

func TestApiBwClient_SyncErrorPropagates(t *testing.T) {
	crypto := &MockCryptoSession{SyncErr: assert.AnError}
	client := newUnlockedAPIClient(t, nil, crypto)
	_, err := client.ListItemsInFolder("f1")
	assert.ErrorIs(t, err, assert.AnError)
}
