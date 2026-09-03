package infra

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"bwsf/src/config"
	"bwsf/src/core"

	"github.com/google/uuid"
)

// ErrAPINotImplemented is retained for older call sites; vault methods no longer return it.
var ErrAPINotImplemented = errors.New(
	"API vault operation is unavailable. " +
		"Authenticate with `bwsf auth` and unlock with your master password when prompted",
)

// ErrAPINotAuthenticated means Identity login has not succeeded in this process.
var ErrAPINotAuthenticated = errors.New(
	"API backend is not authenticated. Run `bwsf auth` to store your Personal API Key and obtain a token",
)

// ErrAPINotUnlocked means Identity auth succeeded but vault decryption keys are missing.
var ErrAPINotUnlocked = errors.New(
	"API vault is locked. Enter your master password to unlock decryption keys",
)

// ErrAPIUnlockNotImplemented is retained for older call sites; Unlock is implemented.
var ErrAPIUnlockNotImplemented = errors.New(
	"API vault unlock is unavailable; enter your master password when prompted",
)

// ErrAPIFolderNotFound means the configured folder name does not exist (active).
var ErrAPIFolderNotFound = errors.New("configured Bitwarden folder not found")

// ErrAPIDuplicateFolder means multiple active folders share the configured name.
var ErrAPIDuplicateFolder = errors.New("multiple Bitwarden folders with the same name")

// ErrAPIDuplicateNote means multiple active Secure Notes share the requested name.
var ErrAPIDuplicateNote = errors.New("multiple secure notes with the same name")

// ApiBwClient is the Bitwarden API backend adapter.
// Step 2+: Personal API Key + Identity token.
// Step 3+: master-password unlock (SDK).
// Step 4+: folder / Secure Note CRUD via CryptoSession.
// Step 5: API is the only factory-selected backend; bw adapter remains optional.
type ApiBwClient struct {
	cfg      *config.Config
	host     *config.Host
	store    SecretStore
	identity *IdentityClient
	crypto   CryptoSession

	mu    sync.Mutex
	token *TokenSet
}

// NewApiBwClient creates an ApiBwClient using the config's default host.
func NewApiBwClient(cfg *config.Config) *ApiBwClient {
	var host *config.Host
	if cfg != nil {
		host = cfg.DefaultHost()
	}
	return NewApiBwClientForHost(cfg, host)
}

// NewApiBwClientForHost creates an ApiBwClient bound to a specific host.
func NewApiBwClientForHost(cfg *config.Config, host *config.Host) *ApiBwClient {
	return NewApiBwClientWithDepsForHost(cfg, host, NewKeyringStore(), NewIdentityClient(), NewSDKCryptoSession())
}

// NewApiBwClientWithDeps creates an ApiBwClient with injectable dependencies (tests).
// When host is nil, DefaultHost() from cfg is used.
func NewApiBwClientWithDeps(cfg *config.Config, store SecretStore, identity *IdentityClient, crypto CryptoSession) *ApiBwClient {
	var host *config.Host
	if cfg != nil {
		host = cfg.DefaultHost()
	}
	return NewApiBwClientWithDepsForHost(cfg, host, store, identity, crypto)
}

// NewApiBwClientWithDepsForHost creates an ApiBwClient for a host with injectable deps.
func NewApiBwClientWithDepsForHost(
	cfg *config.Config,
	host *config.Host,
	store SecretStore,
	identity *IdentityClient,
	crypto CryptoSession,
) *ApiBwClient {
	if cfg == nil {
		cfg = &config.Config{}
	}
	if store == nil {
		store = NewKeyringStore()
	}
	if identity == nil {
		identity = NewIdentityClient()
	}
	if crypto == nil {
		crypto = NewSDKCryptoSession()
	}
	return &ApiBwClient{
		cfg:      cfg,
		host:     host,
		store:    store,
		identity: identity,
		crypto:   crypto,
	}
}

func (c *ApiBwClient) requireHost() error {
	if c == nil || c.host == nil {
		return fmt.Errorf("no host selected; configure a default host or pass --host")
	}
	return nil
}

// EnsureDeviceIdentifier returns a stable device id for host, persisting one when missing.
func EnsureDeviceIdentifier(cfg *config.Config, host *config.Host) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("config is nil")
	}
	if host == nil {
		return "", fmt.Errorf("no host selected; configure a default host or pass --host")
	}
	if host.DeviceIdentifier != "" {
		return host.DeviceIdentifier, nil
	}
	id := uuid.NewString()
	if err := config.UpdateHostDeviceIdentifier(cfg, host.ID, id); err != nil {
		return "", fmt.Errorf("failed to persist device_identifier: %w", err)
	}
	return id, nil
}

// Authenticate loads the Personal API Key from the secret store and obtains an Identity token.
func (c *ApiBwClient) Authenticate(ctx context.Context) error {
	if err := c.requireHost(); err != nil {
		return err
	}
	creds, err := LoadAPICredentials(c.store, c.host.ID)
	if err != nil {
		if errors.Is(err, ErrSecretNotFound) {
			return ErrAPINotAuthenticated
		}
		return err
	}
	return c.authenticateWithCredentials(ctx, creds)
}

// AuthenticateWithCredentials stores creds (optional persist) and obtains a token.
func (c *ApiBwClient) AuthenticateWithCredentials(ctx context.Context, creds APICredentials, persist bool) error {
	if persist {
		if err := c.requireHost(); err != nil {
			return err
		}
		if err := SaveAPICredentials(c.store, c.host.ID, creds); err != nil {
			return err
		}
	}
	return c.authenticateWithCredentials(ctx, creds)
}

func (c *ApiBwClient) authenticateWithCredentials(ctx context.Context, creds APICredentials) error {
	if creds.ClientID == "" || creds.ClientSecret == "" {
		return fmt.Errorf("client_id and client_secret are required")
	}
	if err := c.requireHost(); err != nil {
		return err
	}

	identityBase, err := ResolveIdentityBase(c.host.Type, c.host.HostURL)
	if err != nil {
		return err
	}

	deviceID, err := EnsureDeviceIdentifier(c.cfg, c.host)
	if err != nil {
		return err
	}
	device := DefaultDeviceInfo(deviceID)

	token, err := c.identity.FetchClientCredentialsToken(ctx, identityBase, creds.ClientID, creds.ClientSecret, device)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.token = token
	c.mu.Unlock()
	return nil
}

// EnsureAccessToken returns a valid access token, refreshing or re-authing when needed.
func (c *ApiBwClient) EnsureAccessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	token := c.token
	c.mu.Unlock()

	if token.Valid() {
		return token.AccessToken, nil
	}

	if token != nil && token.RefreshToken != "" {
		if err := c.requireHost(); err != nil {
			return "", err
		}
		identityBase, err := ResolveIdentityBase(c.host.Type, c.host.HostURL)
		if err != nil {
			return "", err
		}
		deviceID, err := EnsureDeviceIdentifier(c.cfg, c.host)
		if err != nil {
			return "", err
		}
		creds, _ := LoadAPICredentials(c.store, c.host.ID)
		refreshed, err := c.identity.RefreshAccessToken(
			ctx, identityBase, token.RefreshToken, creds.ClientID, DefaultDeviceInfo(deviceID),
		)
		if err == nil {
			c.mu.Lock()
			c.token = refreshed
			c.mu.Unlock()
			return refreshed.AccessToken, nil
		}
	}

	if err := c.Authenticate(ctx); err != nil {
		return "", err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token == nil || c.token.AccessToken == "" {
		return "", ErrAPINotAuthenticated
	}
	return c.token.AccessToken, nil
}

// IsAuthenticated reports whether this process holds a non-expired access token.
func (c *ApiBwClient) IsAuthenticated() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.token.Valid()
}

// IsUnlocked reports whether vault decryption keys are present in memory.
func (c *ApiBwClient) IsUnlocked() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.crypto != nil && c.crypto.Unlocked()
}

// ClearSession drops the in-memory token and decryption keys (does not remove Keychain secrets).
func (c *ApiBwClient) ClearSession() {
	c.mu.Lock()
	c.token = nil
	crypto := c.crypto
	c.mu.Unlock()
	if crypto != nil {
		crypto.Lock()
	}
}

// HostID returns the bound host id, or "".
func (c *ApiBwClient) HostID() string {
	if c == nil || c.host == nil {
		return ""
	}
	return c.host.ID
}

// TryRestoreVaultUnlock loads hosts/<id>/vault_unlock and restores crypto keys.
// Returns restored=true on success. Missing blob → (false, nil).
// On restore failure the blob is deleted and (false, nil) is returned.
func (c *ApiBwClient) TryRestoreVaultUnlock() (bool, error) {
	if err := c.requireHost(); err != nil {
		return false, err
	}
	blob, err := GetVaultUnlock(c.store, c.host.ID)
	if err != nil {
		if errors.Is(err, ErrSecretNotFound) {
			return false, nil
		}
		return false, err
	}
	if strings.TrimSpace(blob) == "" {
		_ = DeleteVaultUnlock(c.store, c.host.ID)
		return false, nil
	}

	c.mu.Lock()
	crypto := c.crypto
	c.mu.Unlock()
	if crypto == nil {
		crypto = NewSDKCryptoSession()
		c.mu.Lock()
		c.crypto = crypto
		c.mu.Unlock()
	}

	serverURL := ""
	if c.host.Type == config.HostTypeSelfhost || c.host.Type == "selfhosted" {
		serverURL = c.host.HostURL
	}
	if err := crypto.RestoreUnlockBlob(context.Background(), blob, serverURL); err != nil {
		_ = DeleteVaultUnlock(c.store, c.host.ID)
		return false, nil
	}
	return true, nil
}

// DiscardVaultUnlock deletes the host's vault_unlock blob (and locks in-memory crypto).
func (c *ApiBwClient) DiscardVaultUnlock() error {
	if err := c.requireHost(); err != nil {
		return err
	}
	c.mu.Lock()
	crypto := c.crypto
	c.mu.Unlock()
	if crypto != nil {
		crypto.Lock()
	}
	return DeleteVaultUnlock(c.store, c.host.ID)
}

// LockVaultSession deletes vault_unlock for the bound host (API keys remain).
// Missing blob is a no-op success. Does not require the client to be unlocked.
func (c *ApiBwClient) LockVaultSession() error {
	if err := c.requireHost(); err != nil {
		return err
	}
	c.mu.Lock()
	crypto := c.crypto
	c.mu.Unlock()
	if crypto != nil {
		crypto.Lock()
	}
	return DeleteVaultUnlock(c.store, c.host.ID)
}

// LockVaultSessionForHost deletes vault_unlock for an explicit host id (CLI lock / lock --all).
func LockVaultSessionForHost(store SecretStore, hostID string) error {
	return DeleteVaultUnlock(store, hostID)
}

// TokenExpiresAt exposes token expiry for diagnostics/tests.
func (c *ApiBwClient) TokenExpiresAt() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token == nil {
		return time.Time{}
	}
	return c.token.ExpiresAt
}

func (c *ApiBwClient) requireUnlocked() error {
	if !c.IsAuthenticated() {
		if err := c.Authenticate(context.Background()); err != nil {
			return err
		}
	}
	if !c.IsUnlocked() {
		return ErrAPINotUnlocked
	}
	return nil
}

func (c *ApiBwClient) syncVault(ctx context.Context) error {
	if err := c.requireUnlocked(); err != nil {
		return err
	}
	c.mu.Lock()
	crypto := c.crypto
	c.mu.Unlock()
	if crypto == nil {
		return ErrAPINotUnlocked
	}
	return crypto.Sync(ctx)
}

func (c *ApiBwClient) configuredFolderName() string {
	if c.host != nil {
		if name := strings.TrimSpace(c.host.TargetSection); name != "" {
			return name
		}
	}
	return config.DefaultFolderName
}

func (c *ApiBwClient) findConfiguredFolders(ctx context.Context) ([]VaultFolder, error) {
	if err := c.syncVault(ctx); err != nil {
		return nil, err
	}
	c.mu.Lock()
	crypto := c.crypto
	c.mu.Unlock()
	folders, err := crypto.ListFolders(ctx)
	if err != nil {
		return nil, err
	}
	want := c.configuredFolderName()
	var matched []VaultFolder
	for _, f := range folders {
		if strings.TrimSpace(f.Name) == want {
			matched = append(matched, f)
		}
	}
	return matched, nil
}

func (c *ApiBwClient) GetDotenvsFolderID() (string, error) {
	matched, err := c.findConfiguredFolders(context.Background())
	if err != nil {
		return "", err
	}
	switch len(matched) {
	case 0:
		return "", ErrAPIFolderNotFound
	case 1:
		return matched[0].ID, nil
	default:
		return "", ErrAPIDuplicateFolder
	}
}

func (c *ApiBwClient) DotenvsFolderExists() (bool, error) {
	matched, err := c.findConfiguredFolders(context.Background())
	if err != nil {
		return false, err
	}
	switch len(matched) {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, ErrAPIDuplicateFolder
	}
}

func (c *ApiBwClient) CreateDotenvsFolder() error {
	ctx := context.Background()
	matched, err := c.findConfiguredFolders(ctx)
	if err != nil {
		return err
	}
	if len(matched) > 0 {
		return fmt.Errorf("folder %q already exists", c.configuredFolderName())
	}
	c.mu.Lock()
	crypto := c.crypto
	c.mu.Unlock()
	_, err = crypto.CreateFolder(ctx, c.configuredFolderName())
	return err
}

func (c *ApiBwClient) ListItemsInFolder(folderID string) ([]core.Item, error) {
	ctx := context.Background()
	if err := c.syncVault(ctx); err != nil {
		return nil, err
	}
	c.mu.Lock()
	crypto := c.crypto
	c.mu.Unlock()
	items, err := crypto.ListItems(ctx)
	if err != nil {
		return nil, err
	}
	var out []core.Item
	for _, item := range items {
		if item.Deleted || item.Type != "secure_note" {
			continue
		}
		if item.FolderID != folderID {
			continue
		}
		out = append(out, core.Item{ID: item.ID, Name: item.Name})
	}
	return out, nil
}

func (c *ApiBwClient) GetItemByName(folderID, name string) (*core.FullItem, error) {
	ctx := context.Background()
	if err := c.syncVault(ctx); err != nil {
		return nil, err
	}
	c.mu.Lock()
	crypto := c.crypto
	c.mu.Unlock()
	items, err := crypto.ListItems(ctx)
	if err != nil {
		return nil, err
	}
	var matched []VaultItem
	for _, item := range items {
		if item.Deleted || item.Type != "secure_note" {
			continue
		}
		if item.FolderID != folderID || item.Name != name {
			continue
		}
		matched = append(matched, item)
	}
	switch len(matched) {
	case 0:
		return nil, nil
	case 1:
		return &core.FullItem{ID: matched[0].ID, Name: matched[0].Name, Notes: matched[0].Notes}, nil
	default:
		return nil, ErrAPIDuplicateNote
	}
}

func (c *ApiBwClient) GetItemByID(id string) (*core.FullItem, error) {
	ctx := context.Background()
	if err := c.syncVault(ctx); err != nil {
		return nil, err
	}
	c.mu.Lock()
	crypto := c.crypto
	c.mu.Unlock()
	item, err := crypto.GetItem(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil || item.Deleted || item.Type != "secure_note" {
		return nil, fmt.Errorf("secure note %q not found", id)
	}
	return &core.FullItem{ID: item.ID, Name: item.Name, Notes: item.Notes}, nil
}

func (c *ApiBwClient) CreateNoteItem(folderID, name, notes string) error {
	ctx := context.Background()
	if err := c.syncVault(ctx); err != nil {
		return err
	}
	c.mu.Lock()
	crypto := c.crypto
	c.mu.Unlock()
	_, err := crypto.CreateSecureNote(ctx, folderID, name, notes)
	return err
}

func (c *ApiBwClient) UpdateNoteItem(id, notes string) error {
	ctx := context.Background()
	if err := c.syncVault(ctx); err != nil {
		return err
	}
	c.mu.Lock()
	crypto := c.crypto
	c.mu.Unlock()
	existing, err := crypto.GetItem(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil || existing.Deleted || existing.Type != "secure_note" {
		return fmt.Errorf("secure note %q not found", id)
	}
	_, err = crypto.UpdateSecureNote(ctx, id, existing.FolderID, existing.Name, notes)
	return err
}

// Login for the API backend authenticates with a stored Personal API Key.
// email/password are ignored (CLI-compatible signature); use `bwsf auth` to store the key.
func (c *ApiBwClient) Login(email, password, serverURL string) error {
	_ = email
	_ = password
	if err := c.requireHost(); err != nil {
		return err
	}
	if serverURL != "" {
		switch c.host.Type {
		case config.HostTypeSelfhost, "selfhosted", "":
			if strings.TrimSpace(c.host.HostURL) == "" {
				c.host.HostURL = serverURL
			}
			if c.host.Type == "" {
				c.host.Type = config.HostTypeSelfhost
			}
		}
	}
	return c.Authenticate(context.Background())
}

// Unlock restores vault decryption keys with the master password.
// Caveat (pending Scenario C): keys are obtained via SDK password login using
// host email + MP. Identity Personal API Key tokens remain separate.
func (c *ApiBwClient) Unlock(masterPassword string) error {
	if masterPassword == "" {
		return fmt.Errorf("master password cannot be empty")
	}
	if err := c.requireHost(); err != nil {
		return err
	}
	if !c.IsAuthenticated() {
		if err := c.Authenticate(context.Background()); err != nil {
			return err
		}
	}
	if strings.TrimSpace(c.host.Email) == "" {
		return fmt.Errorf("email is required in config for API vault unlock (run bwsf setup)")
	}

	deviceID, err := EnsureDeviceIdentifier(c.cfg, c.host)
	if err != nil {
		return err
	}
	device := DefaultDeviceInfo(deviceID)

	serverURL := ""
	if c.host.Type == config.HostTypeSelfhost || c.host.Type == "selfhosted" {
		serverURL = c.host.HostURL
	}

	c.mu.Lock()
	crypto := c.crypto
	c.mu.Unlock()
	if crypto == nil {
		crypto = NewSDKCryptoSession()
		c.mu.Lock()
		c.crypto = crypto
		c.mu.Unlock()
	}

	if err := crypto.UnlockWithPassword(context.Background(), c.host.Email, masterPassword, serverURL, device); err != nil {
		return err
	}

	blob, err := crypto.ExportUnlockBlob(context.Background())
	if err != nil {
		return fmt.Errorf("vault unlocked but failed to export session: %w", err)
	}
	if err := SetVaultUnlock(c.store, c.host.ID, blob); err != nil {
		return fmt.Errorf("vault unlocked but failed to persist session: %w", err)
	}
	return nil
}
