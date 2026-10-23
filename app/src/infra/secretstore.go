package infra

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zalando/go-keyring"
)

const (
	keyringService = "bwsf"

	// Legacy flat keys (pre–v0.20). Read-through / migrate only for host "default".
	keyAPIClientID     = "api_client_id"
	keyAPIClientSecret = "api_client_secret"

	keySuffixAPIClientID     = "api_client_id"
	keySuffixAPIClientSecret = "api_client_secret"
	keySuffixVaultUnlock     = "vault_unlock"

	// legacyDefaultHostID is the config migration initial host id (K5).
	legacyDefaultHostID = "default"
)

// ErrSecretNotFound is returned when a secret is missing from the store.
var ErrSecretNotFound = errors.New("secret not found")

// SecretStore persists small secrets (Personal API Key) outside config.json.
type SecretStore interface {
	Set(key, value string) error
	Get(key string) (string, error)
	Delete(key string) error
}

// KeyringStore stores secrets in the OS keyring (macOS Keychain / Linux secret service).
type KeyringStore struct {
	service string
}

// NewKeyringStore creates a SecretStore backed by zalando/go-keyring.
func NewKeyringStore() *KeyringStore {
	return &KeyringStore{service: keyringService}
}

// DefaultSecretStore returns a file-backed store when BWSF_SECRET_DIR is set
// (headless / CI smoke), otherwise the OS keyring.
func DefaultSecretStore() SecretStore {
	if dir := strings.TrimSpace(os.Getenv("BWSF_SECRET_DIR")); dir != "" {
		return NewFileSecretStore(dir)
	}
	return NewKeyringStore()
}

// FileSecretStore persists secrets as files under a directory (smoke / CI).
// Not for production desktop use — prefer the OS keyring.
type FileSecretStore struct {
	dir string
}

// NewFileSecretStore creates a SecretStore under dir (created on first Set).
func NewFileSecretStore(dir string) *FileSecretStore {
	return &FileSecretStore{dir: dir}
}

func (s *FileSecretStore) pathFor(key string) (string, error) {
	if s == nil || strings.TrimSpace(s.dir) == "" {
		return "", fmt.Errorf("file secret store directory is empty")
	}
	if key == "" || strings.Contains(key, "..") {
		return "", fmt.Errorf("invalid secret key %q", key)
	}
	// Encode path separators so host-scoped keys stay a single file name.
	safe := strings.ReplaceAll(key, "/", "__")
	return filepath.Join(s.dir, safe), nil
}

func (s *FileSecretStore) Set(key, value string) error {
	path, err := s.pathFor(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return fmt.Errorf("file secret store mkdir: %w", err)
	}
	if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
		return fmt.Errorf("file secret store set %q: %w", key, err)
	}
	return nil
}

func (s *FileSecretStore) Get(key string) (string, error) {
	path, err := s.pathFor(key)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrSecretNotFound
		}
		return "", fmt.Errorf("file secret store get %q: %w", key, err)
	}
	return string(data), nil
}

func (s *FileSecretStore) Delete(key string) error {
	path, err := s.pathFor(key)
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrSecretNotFound
		}
		return fmt.Errorf("file secret store delete %q: %w", key, err)
	}
	return nil
}

func (s *KeyringStore) Set(key, value string) error {
	if err := keyring.Set(s.service, key, value); err != nil {
		return fmt.Errorf("keyring set %q: %w", key, err)
	}
	return nil
}

func (s *KeyringStore) Get(key string) (string, error) {
	value, err := keyring.Get(s.service, key)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", ErrSecretNotFound
		}
		return "", fmt.Errorf("keyring get %q: %w", key, err)
	}
	return value, nil
}

func (s *KeyringStore) Delete(key string) error {
	err := keyring.Delete(s.service, key)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return ErrSecretNotFound
		}
		return fmt.Errorf("keyring delete %q: %w", key, err)
	}
	return nil
}

// MemorySecretStore is an in-memory SecretStore for unit tests.
type MemorySecretStore struct {
	data map[string]string
}

// NewMemorySecretStore creates an empty in-memory store.
func NewMemorySecretStore() *MemorySecretStore {
	return &MemorySecretStore{data: make(map[string]string)}
}

func (s *MemorySecretStore) Set(key, value string) error {
	s.data[key] = value
	return nil
}

func (s *MemorySecretStore) Get(key string) (string, error) {
	value, ok := s.data[key]
	if !ok {
		return "", ErrSecretNotFound
	}
	return value, nil
}

func (s *MemorySecretStore) Delete(key string) error {
	if _, ok := s.data[key]; !ok {
		return ErrSecretNotFound
	}
	delete(s.data, key)
	return nil
}

// Has reports whether key exists (tests).
func (s *MemorySecretStore) Has(key string) bool {
	_, ok := s.data[key]
	return ok
}

// APICredentials is a Personal API Key pair.
type APICredentials struct {
	ClientID     string
	ClientSecret string
}

func hostSecretKey(hostID, suffix string) string {
	return "hosts/" + hostID + "/" + suffix
}

func requireHostID(hostID string) error {
	if strings.TrimSpace(hostID) == "" {
		return fmt.Errorf("host id is required")
	}
	return nil
}

func deleteIgnoreMissing(store SecretStore, key string) error {
	if err := store.Delete(key); err != nil && !errors.Is(err, ErrSecretNotFound) {
		return err
	}
	return nil
}

// SaveAPICredentials stores a Personal API Key for hostID.
// On success for host "default", legacy flat keys are removed after writing host-scoped keys.
func SaveAPICredentials(store SecretStore, hostID string, creds APICredentials) error {
	if err := requireHostID(hostID); err != nil {
		return err
	}
	if creds.ClientID == "" || creds.ClientSecret == "" {
		return fmt.Errorf("client_id and client_secret are required")
	}
	idKey := hostSecretKey(hostID, keySuffixAPIClientID)
	secretKey := hostSecretKey(hostID, keySuffixAPIClientSecret)
	if err := store.Set(idKey, creds.ClientID); err != nil {
		return err
	}
	if err := store.Set(secretKey, creds.ClientSecret); err != nil {
		_ = deleteIgnoreMissing(store, idKey)
		return err
	}
	if hostID == legacyDefaultHostID {
		_ = deleteIgnoreMissing(store, keyAPIClientID)
		_ = deleteIgnoreMissing(store, keyAPIClientSecret)
	}
	return nil
}

// LoadAPICredentials reads a Personal API Key for hostID.
// For host "default" only, missing host-scoped keys fall back to legacy flat keys;
// a complete flat pair is migrated to hosts/default/... and flat keys are deleted.
func LoadAPICredentials(store SecretStore, hostID string) (APICredentials, error) {
	if err := requireHostID(hostID); err != nil {
		return APICredentials{}, err
	}

	idKey := hostSecretKey(hostID, keySuffixAPIClientID)
	secretKey := hostSecretKey(hostID, keySuffixAPIClientSecret)

	clientID, idErr := store.Get(idKey)
	clientSecret, secretErr := store.Get(secretKey)
	if idErr == nil && secretErr == nil {
		if hostID == legacyDefaultHostID {
			_ = deleteIgnoreMissing(store, keyAPIClientID)
			_ = deleteIgnoreMissing(store, keyAPIClientSecret)
		}
		return APICredentials{ClientID: clientID, ClientSecret: clientSecret}, nil
	}
	// Incomplete host-scoped pair: do not invent values.
	if (idErr == nil) != (secretErr == nil) {
		if idErr != nil && !errors.Is(idErr, ErrSecretNotFound) {
			return APICredentials{}, idErr
		}
		if secretErr != nil && !errors.Is(secretErr, ErrSecretNotFound) {
			return APICredentials{}, secretErr
		}
		return APICredentials{}, ErrSecretNotFound
	}
	if idErr != nil && !errors.Is(idErr, ErrSecretNotFound) {
		return APICredentials{}, idErr
	}

	// Flat read-through only for default.
	if hostID != legacyDefaultHostID {
		return APICredentials{}, ErrSecretNotFound
	}

	flatID, flatIDErr := store.Get(keyAPIClientID)
	flatSecret, flatSecretErr := store.Get(keyAPIClientSecret)
	if flatIDErr != nil || flatSecretErr != nil {
		if errors.Is(flatIDErr, ErrSecretNotFound) || errors.Is(flatSecretErr, ErrSecretNotFound) {
			// Incomplete flat pair: do not migrate.
			if (flatIDErr == nil) != (flatSecretErr == nil) {
				return APICredentials{}, ErrSecretNotFound
			}
			if flatIDErr != nil && !errors.Is(flatIDErr, ErrSecretNotFound) {
				return APICredentials{}, flatIDErr
			}
			if flatSecretErr != nil && !errors.Is(flatSecretErr, ErrSecretNotFound) {
				return APICredentials{}, flatSecretErr
			}
			return APICredentials{}, ErrSecretNotFound
		}
		if flatIDErr != nil {
			return APICredentials{}, flatIDErr
		}
		return APICredentials{}, flatSecretErr
	}

	creds := APICredentials{ClientID: flatID, ClientSecret: flatSecret}
	// Migrate on successful load.
	if err := store.Set(idKey, creds.ClientID); err != nil {
		return creds, nil // read-through succeeded; migration best-effort
	}
	if err := store.Set(secretKey, creds.ClientSecret); err != nil {
		_ = deleteIgnoreMissing(store, idKey)
		return creds, nil
	}
	_ = deleteIgnoreMissing(store, keyAPIClientID)
	_ = deleteIgnoreMissing(store, keyAPIClientSecret)
	return creds, nil
}

// ClearAPICredentials removes the Personal API Key for hostID (not vault_unlock).
// For host "default", legacy flat keys are also removed.
func ClearAPICredentials(store SecretStore, hostID string) error {
	if err := requireHostID(hostID); err != nil {
		return err
	}
	var firstErr error
	for _, key := range []string{
		hostSecretKey(hostID, keySuffixAPIClientID),
		hostSecretKey(hostID, keySuffixAPIClientSecret),
	} {
		if err := store.Delete(key); err != nil && !errors.Is(err, ErrSecretNotFound) {
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	if hostID == legacyDefaultHostID {
		for _, key := range []string{keyAPIClientID, keyAPIClientSecret} {
			if err := store.Delete(key); err != nil && !errors.Is(err, ErrSecretNotFound) {
				if firstErr == nil {
					firstErr = err
				}
			}
		}
	}
	return firstErr
}

// SaveVaultUnlock stores an opaque vault unlock blob for hostID. Empty blob is rejected.
func SaveVaultUnlock(store SecretStore, hostID, blob string) error {
	if err := requireHostID(hostID); err != nil {
		return err
	}
	if blob == "" {
		return fmt.Errorf("vault_unlock blob cannot be empty")
	}
	return store.Set(hostSecretKey(hostID, keySuffixVaultUnlock), blob)
}

// LoadVaultUnlock reads the opaque vault unlock blob for hostID.
func LoadVaultUnlock(store SecretStore, hostID string) (string, error) {
	if err := requireHostID(hostID); err != nil {
		return "", err
	}
	return store.Get(hostSecretKey(hostID, keySuffixVaultUnlock))
}

// ClearVaultUnlock removes the vault unlock blob for hostID. Missing key is a no-op success.
func ClearVaultUnlock(store SecretStore, hostID string) error {
	if err := requireHostID(hostID); err != nil {
		return err
	}
	return deleteIgnoreMissing(store, hostSecretKey(hostID, keySuffixVaultUnlock))
}

// SetVaultUnlock is an alias of SaveVaultUnlock.
func SetVaultUnlock(store SecretStore, hostID, blob string) error {
	return SaveVaultUnlock(store, hostID, blob)
}

// GetVaultUnlock is an alias of LoadVaultUnlock.
func GetVaultUnlock(store SecretStore, hostID string) (string, error) {
	return LoadVaultUnlock(store, hostID)
}

// DeleteVaultUnlock is an alias of ClearVaultUnlock.
func DeleteVaultUnlock(store SecretStore, hostID string) error {
	return ClearVaultUnlock(store, hostID)
}
