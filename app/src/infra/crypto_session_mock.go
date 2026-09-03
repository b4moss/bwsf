package infra

import (
	"context"
	"fmt"
	"strings"
)

// MockCryptoSession is a test double for CryptoSession with an in-memory vault.
type MockCryptoSession struct {
	UnlockErr  error
	SyncErr    error
	ExportErr  error
	RestoreErr error
	ExportBlob string // optional fixed blob returned by ExportUnlockBlob
	unlocked   bool
	Calls      []string
	LastBlob   string
	LastMP     string // last UnlockWithPassword password (tests: must not appear in blob)

	Folders []VaultFolder
	Items   []VaultItem

	nextFolderSeq int
	nextItemSeq   int
}

func (m *MockCryptoSession) UnlockWithPassword(ctx context.Context, email, password, serverURL string, device DeviceInfo) error {
	m.Calls = append(m.Calls, "UnlockWithPassword")
	_ = ctx
	_ = email
	_ = serverURL
	_ = device
	m.LastMP = password
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

func (m *MockCryptoSession) ExportUnlockBlob(ctx context.Context) (string, error) {
	m.Calls = append(m.Calls, "ExportUnlockBlob")
	_ = ctx
	if !m.unlocked {
		return "", fmt.Errorf("not unlocked")
	}
	if m.ExportErr != nil {
		return "", m.ExportErr
	}
	if m.ExportBlob != "" {
		return m.ExportBlob, nil
	}
	// Opaque mock blob: never embeds the master password.
	return "v1:mock-unlock-blob", nil
}

func (m *MockCryptoSession) RestoreUnlockBlob(ctx context.Context, blob, serverURL string) error {
	m.Calls = append(m.Calls, "RestoreUnlockBlob")
	_ = ctx
	_ = serverURL
	m.LastBlob = blob
	if m.RestoreErr != nil {
		return m.RestoreErr
	}
	if strings.TrimSpace(blob) == "" {
		return fmt.Errorf("empty unlock blob")
	}
	if blob == "invalid" || strings.HasPrefix(blob, "bad:") {
		return fmt.Errorf("invalid unlock blob")
	}
	if strings.HasPrefix(blob, "v") && !strings.HasPrefix(blob, "v1:") {
		return fmt.Errorf("unknown unlock blob version")
	}
	m.unlocked = true
	return nil
}

func (m *MockCryptoSession) requireUnlocked() error {
	if !m.unlocked {
		return ErrAPINotUnlocked
	}
	return nil
}

func (m *MockCryptoSession) Sync(ctx context.Context) error {
	m.Calls = append(m.Calls, "Sync")
	_ = ctx
	if err := m.requireUnlocked(); err != nil {
		return err
	}
	return m.SyncErr
}

func (m *MockCryptoSession) ListFolders(ctx context.Context) ([]VaultFolder, error) {
	m.Calls = append(m.Calls, "ListFolders")
	_ = ctx
	if err := m.requireUnlocked(); err != nil {
		return nil, err
	}
	out := make([]VaultFolder, len(m.Folders))
	copy(out, m.Folders)
	return out, nil
}

func (m *MockCryptoSession) CreateFolder(ctx context.Context, name string) (VaultFolder, error) {
	m.Calls = append(m.Calls, "CreateFolder:"+name)
	_ = ctx
	if err := m.requireUnlocked(); err != nil {
		return VaultFolder{}, err
	}
	m.nextFolderSeq++
	folder := VaultFolder{ID: fmt.Sprintf("folder-%d", m.nextFolderSeq), Name: name}
	m.Folders = append(m.Folders, folder)
	return folder, nil
}

func (m *MockCryptoSession) ListItems(ctx context.Context) ([]VaultItem, error) {
	m.Calls = append(m.Calls, "ListItems")
	_ = ctx
	if err := m.requireUnlocked(); err != nil {
		return nil, err
	}
	out := make([]VaultItem, len(m.Items))
	copy(out, m.Items)
	return out, nil
}

func (m *MockCryptoSession) GetItem(ctx context.Context, id string) (*VaultItem, error) {
	m.Calls = append(m.Calls, "GetItem:"+id)
	_ = ctx
	if err := m.requireUnlocked(); err != nil {
		return nil, err
	}
	for i := range m.Items {
		if m.Items[i].ID == id {
			item := m.Items[i]
			return &item, nil
		}
	}
	return nil, fmt.Errorf("item %q not found", id)
}

func (m *MockCryptoSession) CreateSecureNote(ctx context.Context, folderID, name, notes string) (VaultItem, error) {
	m.Calls = append(m.Calls, "CreateSecureNote:"+name)
	_ = ctx
	if err := m.requireUnlocked(); err != nil {
		return VaultItem{}, err
	}
	m.nextItemSeq++
	item := VaultItem{
		ID:       fmt.Sprintf("item-%d", m.nextItemSeq),
		Name:     name,
		Notes:    notes,
		FolderID: folderID,
		Type:     string(bitwardenSecureNoteType()),
	}
	m.Items = append(m.Items, item)
	return item, nil
}

func (m *MockCryptoSession) UpdateSecureNote(ctx context.Context, id, folderID, name, notes string) (VaultItem, error) {
	m.Calls = append(m.Calls, "UpdateSecureNote:"+id)
	_ = ctx
	if err := m.requireUnlocked(); err != nil {
		return VaultItem{}, err
	}
	for i := range m.Items {
		if m.Items[i].ID == id {
			m.Items[i].FolderID = folderID
			m.Items[i].Name = name
			m.Items[i].Notes = notes
			m.Items[i].Type = string(bitwardenSecureNoteType())
			return m.Items[i], nil
		}
	}
	return VaultItem{}, fmt.Errorf("item %q not found", id)
}

func bitwardenSecureNoteType() string {
	return "secure_note"
}

// SetUnlocked marks the mock unlocked without going through UnlockWithPassword.
func (m *MockCryptoSession) SetUnlocked(v bool) {
	m.unlocked = v
}

// FindFolderIDsByName is a helper for tests.
func (m *MockCryptoSession) FindFolderIDsByName(name string) []string {
	var ids []string
	for _, f := range m.Folders {
		if strings.TrimSpace(f.Name) == name {
			ids = append(ids, f.ID)
		}
	}
	return ids
}
