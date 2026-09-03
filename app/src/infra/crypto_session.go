package infra

import (
	"context"
	"fmt"

	"github.com/bnema/bitwarden-go-sdk/bitwarden"
)

// VaultFolder is a decrypted folder view used by ApiBwClient.
type VaultFolder struct {
	ID   string
	Name string
}

// VaultItem is a decrypted vault item view used by ApiBwClient.
type VaultItem struct {
	ID       string
	Name     string
	Notes    string
	FolderID string
	Type     string // "secure_note", "login", ...
	Deleted  bool
}

// CryptoSession restores vault keys and exposes vault read/write used by ApiBwClient.
// Vault methods are only valid while Unlocked() is true.
type CryptoSession interface {
	UnlockWithPassword(ctx context.Context, email, password, serverURL string, device DeviceInfo) error
	Lock()
	Unlocked() bool

	Sync(ctx context.Context) error
	ListFolders(ctx context.Context) ([]VaultFolder, error)
	CreateFolder(ctx context.Context, name string) (VaultFolder, error)
	ListItems(ctx context.Context) ([]VaultItem, error)
	GetItem(ctx context.Context, id string) (*VaultItem, error)
	CreateSecureNote(ctx context.Context, folderID, name, notes string) (VaultItem, error)
	UpdateSecureNote(ctx context.Context, id, folderID, name, notes string) (VaultItem, error)
}

// SDKCryptoSession wraps bitwarden-go-sdk Client for unlock/lock and vault CRUD.
type SDKCryptoSession struct {
	client *bitwarden.Client
}

// NewSDKCryptoSession returns a locked crypto session.
func NewSDKCryptoSession() *SDKCryptoSession {
	return &SDKCryptoSession{}
}

// UnlockWithPassword performs SDK BeginLogin to restore user keys.
func (s *SDKCryptoSession) UnlockWithPassword(ctx context.Context, email, password, serverURL string, device DeviceInfo) error {
	if email == "" {
		return fmt.Errorf("email is required in config for API vault unlock (run bwsf setup)")
	}
	if password == "" {
		return fmt.Errorf("master password cannot be empty")
	}

	var opts []bitwarden.Option
	if serverURL != "" {
		opts = append(opts, bitwarden.WithServerURL(serverURL))
	}

	client, err := bitwarden.NewClient(opts...)
	if err != nil {
		return fmt.Errorf("failed to create vault SDK client: %w", err)
	}

	result, err := client.BeginLogin(ctx, bitwarden.LoginOptions{
		Email:            email,
		Password:         password,
		DeviceType:       sdkDeviceTypeName(device.Type),
		DeviceIdentifier: device.Identifier,
		DeviceName:       device.Name,
	})
	if err != nil {
		_ = client.Logout(ctx)
		return fmt.Errorf("failed to unlock vault with master password: %w", err)
	}
	if result.Challenge != nil {
		result.Challenge.Close()
		_ = client.Logout(ctx)
		return fmt.Errorf("two-factor authentication is required but not supported yet for API unlock")
	}

	if s.client != nil {
		_ = s.client.Logout(context.Background())
	}
	s.client = client
	return nil
}

// Lock discards in-memory SDK keys/tokens.
func (s *SDKCryptoSession) Lock() {
	if s == nil || s.client == nil {
		return
	}
	s.client.Lock()
	_ = s.client.Logout(context.Background())
	s.client = nil
}

// Unlocked reports whether the SDK client holds decryption keys.
func (s *SDKCryptoSession) Unlocked() bool {
	return s != nil && s.client != nil && !s.client.IsLocked()
}

func (s *SDKCryptoSession) requireClient() (*bitwarden.Client, error) {
	if s == nil || s.client == nil || s.client.IsLocked() {
		return nil, ErrAPINotUnlocked
	}
	return s.client, nil
}

// Sync pulls the latest personal vault state.
func (s *SDKCryptoSession) Sync(ctx context.Context) error {
	client, err := s.requireClient()
	if err != nil {
		return err
	}
	if err := client.Sync(ctx); err != nil {
		return fmt.Errorf("failed to sync vault: %w", err)
	}
	return nil
}

// ListFolders returns personal vault folders.
func (s *SDKCryptoSession) ListFolders(ctx context.Context) ([]VaultFolder, error) {
	client, err := s.requireClient()
	if err != nil {
		return nil, err
	}
	folders, err := client.Folders().List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list folders: %w", err)
	}
	out := make([]VaultFolder, 0, len(folders))
	for _, f := range folders {
		out = append(out, VaultFolder{ID: f.ID, Name: f.Name})
	}
	return out, nil
}

// CreateFolder creates a personal vault folder by name.
func (s *SDKCryptoSession) CreateFolder(ctx context.Context, name string) (VaultFolder, error) {
	client, err := s.requireClient()
	if err != nil {
		return VaultFolder{}, err
	}
	folder, err := client.Folders().Create(ctx, name)
	if err != nil {
		return VaultFolder{}, fmt.Errorf("failed to create folder: %w", err)
	}
	return VaultFolder{ID: folder.ID, Name: folder.Name}, nil
}

// ListItems returns vault items (caller filters types / deleted).
func (s *SDKCryptoSession) ListItems(ctx context.Context) ([]VaultItem, error) {
	client, err := s.requireClient()
	if err != nil {
		return nil, err
	}
	items, err := client.Vault().List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list vault items: %w", err)
	}
	out := make([]VaultItem, 0, len(items))
	for _, item := range items {
		out = append(out, vaultItemFromSDK(item))
	}
	return out, nil
}

// GetItem returns one vault item by id.
func (s *SDKCryptoSession) GetItem(ctx context.Context, id string) (*VaultItem, error) {
	client, err := s.requireClient()
	if err != nil {
		return nil, err
	}
	item, err := client.Vault().Get(ctx, bitwarden.ItemID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get vault item: %w", err)
	}
	v := vaultItemFromSDK(item)
	return &v, nil
}

// CreateSecureNote creates a Secure Note cipher.
func (s *SDKCryptoSession) CreateSecureNote(ctx context.Context, folderID, name, notes string) (VaultItem, error) {
	client, err := s.requireClient()
	if err != nil {
		return VaultItem{}, err
	}
	created, err := client.Vault().Create(ctx, bitwarden.Item{
		Name:     name,
		Notes:    notes,
		Type:     bitwarden.ItemTypeSecureNote,
		FolderID: folderID,
	})
	if err != nil {
		return VaultItem{}, fmt.Errorf("failed to create secure note: %w", err)
	}
	return vaultItemFromSDK(created), nil
}

// UpdateSecureNote updates an existing Secure Note's notes (and keeps name/folder).
func (s *SDKCryptoSession) UpdateSecureNote(ctx context.Context, id, folderID, name, notes string) (VaultItem, error) {
	client, err := s.requireClient()
	if err != nil {
		return VaultItem{}, err
	}
	updated, err := client.Vault().Update(ctx, bitwarden.ItemID(id), bitwarden.Item{
		ID:       bitwarden.ItemID(id),
		Name:     name,
		Notes:    notes,
		Type:     bitwarden.ItemTypeSecureNote,
		FolderID: folderID,
	})
	if err != nil {
		return VaultItem{}, fmt.Errorf("failed to update secure note: %w", err)
	}
	return vaultItemFromSDK(updated), nil
}

func vaultItemFromSDK(item bitwarden.Item) VaultItem {
	return VaultItem{
		ID:       string(item.ID),
		Name:     item.Name,
		Notes:    item.Notes,
		FolderID: item.FolderID,
		Type:     string(item.Type),
		Deleted:  false, // public List/Get expose active items
	}
}

func sdkDeviceTypeName(numericOrName string) string {
	switch numericOrName {
	case "6", "WindowsDesktop":
		return "WindowsDesktop"
	case "7", "MacOsDesktop":
		return "MacOsDesktop"
	case "8", "LinuxDesktop":
		return "LinuxDesktop"
	default:
		if numericOrName != "" {
			return numericOrName
		}
		return "LinuxDesktop"
	}
}
