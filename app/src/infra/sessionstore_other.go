//go:build !darwin

package infra

import "bwsf/src/core"

// noopSessionStore is used on non-macOS platforms (Linux no-op for #130).
type noopSessionStore struct{}

// NewSessionStore returns a no-op SessionStore on non-macOS platforms.
func NewSessionStore() core.SessionStore {
	return noopSessionStore{}
}

func (noopSessionStore) Get() (string, error) { return "", nil }
func (noopSessionStore) Set(string) error     { return nil }
func (noopSessionStore) Delete() error        { return nil }
