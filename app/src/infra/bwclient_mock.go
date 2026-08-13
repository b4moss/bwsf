package infra

import (
	"bwsf/src/core"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// =============================================================================
// MockFileSystem - テスト用のファイルシステムモック
// =============================================================================

// MockFileSystem はテスト用のモック FileSystem 実装です。
type MockFileSystem struct {
	mu    sync.RWMutex
	files map[string][]byte
}

// NewMockFileSystem は MockFileSystem の新しいインスタンスを作成します。
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		files: make(map[string][]byte),
	}
}

// OpenEnvFile は .env ファイルを読み込みます。
func (fs *MockFileSystem) OpenEnvFile(path string) ([]byte, error) {
	return fs.ReadFile(path)
}

// ReadFile はファイルを読み込みます。
func (fs *MockFileSystem) ReadFile(path string) ([]byte, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	data, ok := fs.files[path]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	return data, nil
}

// WriteFile はファイルを書き込みます。
func (fs *MockFileSystem) WriteFile(path string, data []byte, perm uint32) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.files[path] = data
	return nil
}

// Remove はファイルを削除します。
func (fs *MockFileSystem) Remove(path string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if _, ok := fs.files[path]; !ok {
		return fmt.Errorf("file not found: %s", path)
	}
	delete(fs.files, path)
	return nil
}

// Stat はファイル情報を取得します。
func (fs *MockFileSystem) Stat(path string) (core.FileInfo, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	_, ok := fs.files[path]
	return &mockFileInfo{notExist: !ok}, nil
}

// MkdirAll はディレクトリを再帰的に作成します（モックでは何もしない）。
func (fs *MockFileSystem) MkdirAll(path string, perm uint32) error {
	return nil
}

// ReadDir はディレクトリ内のエントリを読み込みます。
// ファイルマップからディレクトリ内のファイルを抽出します。
func (fs *MockFileSystem) ReadDir(path string) ([]core.DirEntry, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	// パスを正規化
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	// ディレクトリ内のファイルを抽出（重複を排除）
	seen := make(map[string]bool)
	var entries []core.DirEntry

	for filePath := range fs.files {
		// ファイルがこのディレクトリにあるかチェック
		if strings.HasPrefix(filePath, path) {
			relativePath := strings.TrimPrefix(filePath, path)
			// サブディレクトリのファイルは除外（直接の子のみ）
			if strings.Contains(relativePath, "/") {
				continue
			}
			fileName := filepath.Base(filePath)
			if !seen[fileName] {
				seen[fileName] = true
				entries = append(entries, &mockDirEntry{name: fileName, isDir: false})
			}
		}
	}

	// ファイル名でソート
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	return entries, nil
}

// mockDirEntry は core.DirEntry インターフェースの実装です。
type mockDirEntry struct {
	name  string
	isDir bool
}

// Name はエントリ名を返します。
func (e *mockDirEntry) Name() string {
	return e.name
}

// IsDir はディレクトリかどうかを返します。
func (e *mockDirEntry) IsDir() bool {
	return e.isDir
}

// SetFile はテスト用にファイルを設定します。
func (fs *MockFileSystem) SetFile(path string, data []byte) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.files[path] = data
}

// GetFile はテスト用にファイルを取得します。
func (fs *MockFileSystem) GetFile(path string) ([]byte, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	data, ok := fs.files[path]
	return data, ok
}

// mockFileInfo は core.FileInfo インターフェースの実装です。
type mockFileInfo struct {
	notExist bool
}

// IsNotExist はファイルが存在しないかどうかを返します。
func (fi *mockFileInfo) IsNotExist() bool {
	return fi.notExist
}

// =============================================================================
// MockLogger - テスト用のロガーモック
// =============================================================================

// MockLogger はテスト用のモック Logger 実装です。
type MockLogger struct {
	mu        sync.Mutex
	InfoLogs  []string
	ErrorLogs []string
}

// NewMockLogger は MockLogger の新しいインスタンスを作成します。
func NewMockLogger() *MockLogger {
	return &MockLogger{
		InfoLogs:  []string{},
		ErrorLogs: []string{},
	}
}

// Info はInfoログを記録します。
func (l *MockLogger) Info(args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.InfoLogs = append(l.InfoLogs, fmt.Sprint(args...))
}

// Error はErrorログを記録します。
func (l *MockLogger) Error(args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.ErrorLogs = append(l.ErrorLogs, fmt.Sprint(args...))
}

// =============================================================================
// MockBwClient - テスト用のBwClientモック
// =============================================================================

// MockBwClient はテスト用のモック BwClient 実装です。
// インメモリでBitwardenの動作をシミュレートします。
type MockBwClient struct {
	mu sync.RWMutex

	// ストレージ
	folders       map[string]string         // folderName -> folderID
	items         map[string]*core.FullItem // itemID -> FullItem
	itemsByFolder map[string][]string       // folderID -> []itemID

	// 認証状態
	isLoggedIn bool
	isUnlocked bool
	email      string
	password   string
	serverURL  string

	// テスト用のフック
	LoginFunc  func(email, password, serverURL string) error
	UnlockFunc func(masterPassword string) error
}

// NewMockBwClient は MockBwClient の新しいインスタンスを作成します。
func NewMockBwClient() *MockBwClient {
	m := &MockBwClient{
		folders:       make(map[string]string),
		items:         make(map[string]*core.FullItem),
		itemsByFolder: make(map[string][]string),
	}
	return m
}

// SetupTestData はテスト用の初期データをセットアップします。
func (m *MockBwClient) SetupTestData() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// dotenvs フォルダを作成
	m.folders["dotenvs"] = "folder-dotenvs-id"
	m.itemsByFolder["folder-dotenvs-id"] = []string{}

	// ログイン状態を設定
	m.isLoggedIn = true
	m.isUnlocked = true
	m.email = "test@example.com"
	m.password = "testpassword"
}

// GetDotenvsFolderID は dotenvs フォルダの ID を取得します。
func (m *MockBwClient) GetDotenvsFolderID() (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.isUnlocked {
		return "", fmt.Errorf("Bitwarden CLI is locked")
	}

	folderID, ok := m.folders["dotenvs"]
	if !ok {
		return "", fmt.Errorf("dotenvs folder not found")
	}
	return folderID, nil
}

// DotenvsFolderExists は dotenvs フォルダが存在するかどうかを確認します。
func (m *MockBwClient) DotenvsFolderExists() (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.isUnlocked {
		return false, fmt.Errorf("Bitwarden CLI is locked")
	}

	_, ok := m.folders["dotenvs"]
	return ok, nil
}

// CreateDotenvsFolder は dotenvs フォルダを作成します。
func (m *MockBwClient) CreateDotenvsFolder() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isUnlocked {
		return fmt.Errorf("Bitwarden CLI is locked")
	}

	// すでに存在する場合はエラー
	if _, ok := m.folders["dotenvs"]; ok {
		return fmt.Errorf("dotenvs folder already exists")
	}

	// フォルダを作成
	folderID := "folder-dotenvs-id"
	m.folders["dotenvs"] = folderID
	m.itemsByFolder[folderID] = []string{}

	return nil
}

// ListItemsInFolder は指定フォルダ内のアイテム一覧を取得します。
func (m *MockBwClient) ListItemsInFolder(folderID string) ([]core.Item, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.isUnlocked {
		return nil, fmt.Errorf("Bitwarden CLI is locked")
	}

	itemIDs, ok := m.itemsByFolder[folderID]
	if !ok {
		return []core.Item{}, nil
	}

	result := make([]core.Item, 0, len(itemIDs))
	for _, id := range itemIDs {
		if item, ok := m.items[id]; ok {
			result = append(result, core.Item{
				ID:   item.ID,
				Name: item.Name,
			})
		}
	}
	return result, nil
}

// GetItemByName は指定フォルダ内のアイテムを名前で検索します。
func (m *MockBwClient) GetItemByName(folderID, name string) (*core.FullItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.isUnlocked {
		return nil, fmt.Errorf("Bitwarden CLI is locked")
	}

	itemIDs, ok := m.itemsByFolder[folderID]
	if !ok {
		return nil, nil
	}

	for _, id := range itemIDs {
		if item, ok := m.items[id]; ok && item.Name == name {
			return &core.FullItem{
				ID:    item.ID,
				Name:  item.Name,
				Notes: item.Notes,
			}, nil
		}
	}
	return nil, nil
}

// GetItemByID は指定 ID のアイテムを取得します。
func (m *MockBwClient) GetItemByID(id string) (*core.FullItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.isUnlocked {
		return nil, fmt.Errorf("Bitwarden CLI is locked")
	}

	item, ok := m.items[id]
	if !ok {
		return nil, nil
	}
	return &core.FullItem{
		ID:    item.ID,
		Name:  item.Name,
		Notes: item.Notes,
	}, nil
}

// CreateNoteItem は新しいノートアイテムを作成します。
func (m *MockBwClient) CreateNoteItem(folderID, name, notes string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isUnlocked {
		return fmt.Errorf("Bitwarden CLI is locked")
	}

	// 新しいIDを生成
	itemID := fmt.Sprintf("item-%s-%d", name, len(m.items)+1)

	// アイテムを作成
	item := &core.FullItem{
		ID:    itemID,
		Name:  name,
		Notes: notes,
	}
	m.items[itemID] = item

	// フォルダに追加
	m.itemsByFolder[folderID] = append(m.itemsByFolder[folderID], itemID)

	return nil
}

// UpdateNoteItem は既存のノートアイテムを更新します。
func (m *MockBwClient) UpdateNoteItem(id, notes string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isUnlocked {
		return fmt.Errorf("Bitwarden CLI is locked")
	}

	item, ok := m.items[id]
	if !ok {
		return fmt.Errorf("item not found: %s", id)
	}

	item.Notes = notes
	return nil
}

// Login は Bitwarden にログインします。
func (m *MockBwClient) Login(email, password, serverURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// カスタムフックがあれば使用
	if m.LoginFunc != nil {
		return m.LoginFunc(email, password, serverURL)
	}

	m.isLoggedIn = true
	m.email = email
	m.password = password
	m.serverURL = serverURL
	return nil
}

// Unlock は Bitwarden をアンロックします。
func (m *MockBwClient) Unlock(masterPassword string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// カスタムフックがあれば使用
	if m.UnlockFunc != nil {
		return m.UnlockFunc(masterPassword)
	}

	if !m.isLoggedIn {
		return fmt.Errorf("not logged in")
	}

	m.isUnlocked = true
	return nil
}

// SetLoggedIn はログイン状態を設定します（テスト用）。
func (m *MockBwClient) SetLoggedIn(loggedIn bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isLoggedIn = loggedIn
}

// SetUnlocked はアンロック状態を設定します（テスト用）。
func (m *MockBwClient) SetUnlocked(unlocked bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isUnlocked = unlocked
}

// GetItemCount はアイテム数を返します（テスト用）。
func (m *MockBwClient) GetItemCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.items)
}

// Reset はモックの状態をリセットします（テスト用）。
func (m *MockBwClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.folders = make(map[string]string)
	m.items = make(map[string]*core.FullItem)
	m.itemsByFolder = make(map[string][]string)
	m.isLoggedIn = false
	m.isUnlocked = false
	m.email = ""
	m.password = ""
	m.serverURL = ""
}
