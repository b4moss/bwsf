//go:build darwin

package infra

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"bwsf/src/core"
)

const (
	sessionService = "bwsf"
	sessionAccount = "bw-session"
)

// keychainSessionStore stores BW_SESSION in the macOS Keychain via the
// `security` CLI (single slot). Avoids CGO so darwin builds stay simple.
type keychainSessionStore struct{}

// NewSessionStore returns a Keychain-backed SessionStore on macOS.
func NewSessionStore() core.SessionStore {
	return &keychainSessionStore{}
}

func (s *keychainSessionStore) Get() (string, error) {
	cmd := exec.Command("security", "find-generic-password",
		"-s", sessionService,
		"-a", sessionAccount,
		"-w",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		// Item not found is not an error for callers.
		msg := stderr.String() + err.Error()
		if strings.Contains(msg, "could not be found") ||
			strings.Contains(msg, "The specified item could not be found") {
			return "", nil
		}
		return "", fmt.Errorf("keychain get: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

func (s *keychainSessionStore) Set(session string) error {
	// -U updates if present; still delete first for a clean single-slot overwrite.
	_ = exec.Command("security", "delete-generic-password",
		"-s", sessionService,
		"-a", sessionAccount,
	).Run()

	cmd := exec.Command("security", "add-generic-password",
		"-s", sessionService,
		"-a", sessionAccount,
		"-l", "bwsf BW_SESSION",
		"-w", session,
		"-T", "", // allow this binary / default access; session lifetime is best-effort
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("keychain set: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func (s *keychainSessionStore) Delete() error {
	cmd := exec.Command("security", "delete-generic-password",
		"-s", sessionService,
		"-a", sessionAccount,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		msg := stderr.String() + err.Error()
		if strings.Contains(msg, "could not be found") ||
			strings.Contains(msg, "The specified item could not be found") {
			return nil
		}
		return fmt.Errorf("keychain delete: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}
