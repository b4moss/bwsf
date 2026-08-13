package infra

import "context"

// MockCryptoSession is a test double for CryptoSession.
type MockCryptoSession struct {
	UnlockErr error
	unlocked  bool
	Calls     []string
}

func (m *MockCryptoSession) UnlockWithPassword(ctx context.Context, email, password, serverURL string, device DeviceInfo) error {
	m.Calls = append(m.Calls, "UnlockWithPassword")
	_ = ctx
	_ = email
	_ = password
	_ = serverURL
	_ = device
	if m.UnlockErr != nil {
		return m.UnlockErr
	}
	m.unlocked = true
	return nil
}

func (m *MockCryptoSession) Lock() {
	m.Calls = append(m.Calls, "Lock")
	m.unlocked = false
}

func (m *MockCryptoSession) Unlocked() bool {
	return m.unlocked
}
