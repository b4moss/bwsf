package infra

import (
	"context"
	"fmt"

	"github.com/bnema/bitwarden-go-sdk/bitwarden"
)

// CryptoSession restores and holds vault decryption keys in process memory.
// Implemented with the Community SDK (password login) until Scenario C confirms
// Personal API Key token injection + MP-only unlock.
type CryptoSession interface {
	UnlockWithPassword(ctx context.Context, email, password, serverURL string, device DeviceInfo) error
	Lock()
	Unlocked() bool
}

// SDKCryptoSession wraps bitwarden-go-sdk Client for unlock/lock.
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
