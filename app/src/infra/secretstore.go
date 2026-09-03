package infra

import (
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	keyringService     = "bwsf"
	keyAPIClientID     = "api_client_id"
	keyAPIClientSecret = "api_client_secret"
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

// APICredentials is a Personal API Key pair.
type APICredentials struct {
	ClientID     string
	ClientSecret string
}

// SaveAPICredentials stores a Personal API Key in the secret store.
func SaveAPICredentials(store SecretStore, creds APICredentials) error {
	if creds.ClientID == "" || creds.ClientSecret == "" {
		return fmt.Errorf("client_id and client_secret are required")
	}
	if err := store.Set(keyAPIClientID, creds.ClientID); err != nil {
		return err
	}
	if err := store.Set(keyAPIClientSecret, creds.ClientSecret); err != nil {
		return err
	}
	return nil
}

// LoadAPICredentials reads a Personal API Key from the secret store.
func LoadAPICredentials(store SecretStore) (APICredentials, error) {
	clientID, err := store.Get(keyAPIClientID)
	if err != nil {
		return APICredentials{}, err
	}
	clientSecret, err := store.Get(keyAPIClientSecret)
	if err != nil {
		return APICredentials{}, err
	}
	return APICredentials{ClientID: clientID, ClientSecret: clientSecret}, nil
}

// ClearAPICredentials removes the Personal API Key from the secret store.
func ClearAPICredentials(store SecretStore) error {
	var firstErr error
	if err := store.Delete(keyAPIClientID); err != nil && !errors.Is(err, ErrSecretNotFound) {
		firstErr = err
	}
	if err := store.Delete(keyAPIClientSecret); err != nil && !errors.Is(err, ErrSecretNotFound) {
		if firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
