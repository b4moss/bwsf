package core

import (
	"errors"
	"fmt"
	"testing"

	"bwsf/src/config"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// テスト用モック実装
// =============================================================================

// ErrBitwardenLocked はテスト用のロックエラー
var ErrBitwardenLocked = errors.New("Bitwarden CLI is locked")

// --- mockBwClient ---

type mockBwClient struct {
	calls []string

	// GetDotenvsFolderID の挙動制御
	folderID    string
	folderIDErr error

	// DotenvsFolderExists の挙動制御
	folderExists    bool
	folderExistsErr error

	// CreateDotenvsFolder の挙動制御
	createFolderErr error

	// ListItemsInFolder の挙動制御
	items        []Item
	listItemsErr error

	// GetItemByName の挙動制御
	itemByName    *FullItem
	itemByNameErr error

	// GetItemByID の挙動制御
	itemByID    *FullItem
	itemByIDErr error

	// CreateNoteItem の挙動制御
	createErr error

	// UpdateNoteItem の挙動制御
	updateErr error

	// Login の挙動制御
	loginErr error

	// Unlock の挙動制御
	unlockErr error
}

func (m *mockBwClient) GetDotenvsFolderID() (string, error) {
	m.calls = append(m.calls, "GetDotenvsFolderID")
	if m.folderIDErr != nil {
		return "", m.folderIDErr
	}
	if m.folderID == "" {
		return "dummy-folder-id", nil
	}
	return m.folderID, nil
}

func (m *mockBwClient) DotenvsFolderExists() (bool, error) {
	m.calls = append(m.calls, "DotenvsFolderExists")
	if m.folderExistsErr != nil {
		return false, m.folderExistsErr
	}
	return m.folderExists, nil
}

func (m *mockBwClient) CreateDotenvsFolder() error {
	m.calls = append(m.calls, "CreateDotenvsFolder")
	return m.createFolderErr
}

func (m *mockBwClient) ListItemsInFolder(folderID string) ([]Item, error) {
	m.calls = append(m.calls, fmt.Sprintf("ListItemsInFolder(%s)", folderID))
	if m.listItemsErr != nil {
		return nil, m.listItemsErr
	}
	return m.items, nil
}

func (m *mockBwClient) GetItemByName(folderID, name string) (*FullItem, error) {
	m.calls = append(m.calls, fmt.Sprintf("GetItemByName(%s,%s)", folderID, name))
	if m.itemByNameErr != nil {
		return nil, m.itemByNameErr
	}
	return m.itemByName, nil
}

func (m *mockBwClient) GetItemByID(id string) (*FullItem, error) {
	m.calls = append(m.calls, fmt.Sprintf("GetItemByID(%s)", id))
	if m.itemByIDErr != nil {
		return nil, m.itemByIDErr
	}
	return m.itemByID, nil
}

func (m *mockBwClient) CreateNoteItem(folderID, name, notes string) error {
	m.calls = append(m.calls, fmt.Sprintf("CreateNoteItem(%s,%s)", folderID, name))
	return m.createErr
}

func (m *mockBwClient) UpdateNoteItem(id, notes string) error {
	m.calls = append(m.calls, fmt.Sprintf("UpdateNoteItem(%s)", id))
	return m.updateErr
}

func (m *mockBwClient) Login(email, password, serverURL string) error {
	m.calls = append(m.calls, fmt.Sprintf("Login(%s,%s)", email, serverURL))
	return m.loginErr
}

func (m *mockBwClient) Unlock(masterPassword string) error {
	m.calls = append(m.calls, "Unlock")
	return m.unlockErr
}

// --- mockFileSystem ---

type mockFileInfo struct {
	notExist bool
}

func (f *mockFileInfo) IsNotExist() bool {
	return f.notExist
}

type mockFileSystem struct {
	calls []string

	// OpenEnvFile の挙動制御
	envFileContent []byte
	envFileErr     error

	// ReadFile の挙動制御
	readContent    []byte
	readContentMap map[string][]byte // ファイルパスごとの内容
	readErr        error

	// WriteFile の挙動制御
	writtenPath  string
	writtenData  []byte
	writtenFiles map[string][]byte // 複数ファイル用
	writeErr     error

	// Remove の挙動制御
	removedFiles []string
	removeErr    error

	// Stat の挙動制御
	statInfo    FileInfo
	statInfoMap map[string]FileInfo // ファイルパスごとの情報
	statErr     error

	// MkdirAll の挙動制御
	mkdirErr error

	// ReadDir の挙動制御
	dirEntries []DirEntry
	readDirErr error
}

func (m *mockFileSystem) OpenEnvFile(path string) ([]byte, error) {
	m.calls = append(m.calls, fmt.Sprintf("OpenEnvFile(%s)", path))
	if m.envFileErr != nil {
		return nil, m.envFileErr
	}
	return m.envFileContent, nil
}

func (m *mockFileSystem) ReadFile(path string) ([]byte, error) {
	m.calls = append(m.calls, fmt.Sprintf("ReadFile(%s)", path))
	if m.readErr != nil {
		return nil, m.readErr
	}
	if m.readContentMap != nil {
		if content, ok := m.readContentMap[path]; ok {
			return content, nil
		}
	}
	return m.readContent, nil
}

func (m *mockFileSystem) WriteFile(path string, data []byte, perm uint32) error {
	m.calls = append(m.calls, fmt.Sprintf("WriteFile(%s)", path))
	m.writtenPath = path
	m.writtenData = data
	if m.writtenFiles == nil {
		m.writtenFiles = make(map[string][]byte)
	}
	m.writtenFiles[path] = data
	return m.writeErr
}

func (m *mockFileSystem) Remove(path string) error {
	m.calls = append(m.calls, fmt.Sprintf("Remove(%s)", path))
	m.removedFiles = append(m.removedFiles, path)
	if m.readContentMap != nil {
		delete(m.readContentMap, path)
	}
	return m.removeErr
}

func (m *mockFileSystem) Stat(path string) (FileInfo, error) {
	m.calls = append(m.calls, fmt.Sprintf("Stat(%s)", path))
	if m.statErr != nil {
		return nil, m.statErr
	}
	if m.statInfoMap != nil {
		if info, ok := m.statInfoMap[path]; ok {
			return info, nil
		}
	}
	if m.statInfo != nil {
		return m.statInfo, nil
	}
	return &mockFileInfo{notExist: true}, nil
}

func (m *mockFileSystem) MkdirAll(path string, perm uint32) error {
	m.calls = append(m.calls, fmt.Sprintf("MkdirAll(%s)", path))
	return m.mkdirErr
}

func (m *mockFileSystem) ReadDir(path string) ([]DirEntry, error) {
	m.calls = append(m.calls, fmt.Sprintf("ReadDir(%s)", path))
	if m.readDirErr != nil {
		return nil, m.readDirErr
	}
	return m.dirEntries, nil
}

// mockDirEntry は DirEntry インターフェースの実装です。
type mockDirEntry struct {
	name  string
	isDir bool
}

func (e *mockDirEntry) Name() string {
	return e.name
}

func (e *mockDirEntry) IsDir() bool {
	return e.isDir
}

// --- mockLogger ---

type mockLogger struct {
	errors   []string
	infos    []string
	warnings []string
	success  []string
}

func (l *mockLogger) Error(args ...interface{}) {
	l.errors = append(l.errors, fmt.Sprint(args...))
}

func (l *mockLogger) Info(args ...interface{}) {
	l.infos = append(l.infos, fmt.Sprint(args...))
}

func (l *mockLogger) Warning(args ...interface{}) {
	l.warnings = append(l.warnings, fmt.Sprint(args...))
}

func (l *mockLogger) Success(args ...interface{}) {
	l.success = append(l.success, fmt.Sprint(args...))
}

// =============================================================================
// WithUnlockRetry のテスト
// =============================================================================

// 正常系: fn が一度で成功する場合は Unlock/Login を行わず、そのまま成功すること
func TestWithUnlockRetry_SuccessWithoutRetry(t *testing.T) {
	bw := &mockBwClient{}
	logger := &mockLogger{}
	cfg := &config.Config{}

	called := 0
	fn := func() error {
		called++
		return nil
	}

	err := WithUnlockRetry(
		bw,
		cfg,
		func() (string, error) {
			return "", errors.New("prompt should not be called")
		},
		logger,
		fn,
	)

	assert.NoError(t, err)
	assert.Equal(t, 1, called, "fn should be called exactly once")
	assert.NotContains(t, bw.calls, "Unlock")
	assert.NotContains(t, bw.calls, "Login")
}

// 正常系: fn が 1 回目で ErrBitwardenLocked を返し、Unlock 成功後の 2 回目で成功するケース
func TestWithUnlockRetry_LockThenUnlockSuccess(t *testing.T) {
	bw := &mockBwClient{}
	logger := &mockLogger{}
	cfg := &config.Config{Email: "test@example.com"}

	callCount := 0
	fn := func() error {
		callCount++
		if callCount == 1 {
			return ErrBitwardenLocked
		}
		return nil
	}

	promptCalled := false
	err := WithUnlockRetry(
		bw,
		cfg,
		func() (string, error) {
			promptCalled = true
			return "password123", nil
		},
		logger,
		fn,
	)

	assert.NoError(t, err)
	assert.Equal(t, 2, callCount, "fn should be called twice")
	assert.True(t, promptCalled, "promptPassword should be called")
	assert.Contains(t, bw.calls, "Unlock")
}

// 異常系: fn がロック関連以外のエラーを返した場合は、Unlock/Login を試みずにエラーをそのまま返すこと
func TestWithUnlockRetry_NonLockErrorPropagates(t *testing.T) {
	bw := &mockBwClient{}
	logger := &mockLogger{}
	cfg := &config.Config{}

	expectedErr := errors.New("some error")
	fn := func() error {
		return expectedErr
	}

	err := WithUnlockRetry(
		bw,
		cfg,
		func() (string, error) {
			return "pwd", nil
		},
		logger,
		fn,
	)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.NotContains(t, bw.calls, "Unlock")
	assert.NotContains(t, bw.calls, "Login")
}

// 異常系: promptPassword がエラーを返した場合、Unlock/Login が呼ばれずにそのエラーが返る
func TestWithUnlockRetry_PromptPasswordError(t *testing.T) {
	bw := &mockBwClient{}
	logger := &mockLogger{}
	cfg := &config.Config{}

	callCount := 0
	fn := func() error {
		callCount++
		if callCount == 1 {
			return ErrBitwardenLocked
		}
		return nil
	}

	promptErr := errors.New("prompt cancelled")
	err := WithUnlockRetry(
		bw,
		cfg,
		func() (string, error) {
			return "", promptErr
		},
		logger,
		fn,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prompt cancelled")
	assert.Equal(t, 1, callCount, "fn should be called only once before prompt error")
	assert.NotContains(t, bw.calls, "Unlock")
}

// 異常系: Unlock と Login の両方が失敗する場合
func TestWithUnlockRetry_UnlockAndLoginBothFail(t *testing.T) {
	bw := &mockBwClient{
		unlockErr: errors.New("unlock failed"),
		loginErr:  errors.New("login failed"),
	}
	logger := &mockLogger{}
	cfg := &config.Config{Email: "test@example.com", SelfhostedURL: "https://bw.example.com"}

	callCount := 0
	fn := func() error {
		callCount++
		return ErrBitwardenLocked
	}

	err := WithUnlockRetry(
		bw,
		cfg,
		func() (string, error) {
			return "password", nil
		},
		logger,
		fn,
	)

	assert.Error(t, err)
	assert.Equal(t, 1, callCount, "fn should not be retried after unlock/login failure")
	assert.Contains(t, bw.calls, "Unlock")
	assert.Contains(t, bw.calls, "Login(test@example.com,https://bw.example.com)")
}

// =============================================================================
// PushEnvCore のテスト
// =============================================================================

// 正常系: .env が存在し、既存アイテムがないケースで CreateNoteItem が呼ばれる
func TestPushEnvCore_CreateNewItem(t *testing.T) {
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: nil, // アイテムなし
	}
	fs := &mockFileSystem{
		dirEntries: []DirEntry{
			&mockDirEntry{name: ".env", isDir: false},
		},
		readContentMap: map[string][]byte{
			".env": []byte("KEY=value\n"),
		},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := PushEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		logger,
	)

	assert.NoError(t, err)
	assert.Contains(t, bw.calls, "GetDotenvsFolderID")
	assert.Contains(t, bw.calls, "GetItemByName(folder-123,my-project)")
	assert.Contains(t, bw.calls, "CreateNoteItem(folder-123,my-project)")
}

// 正常系: .env が存在し、既存アイテムがあるケースで UpdateNoteItem が呼ばれる
func TestPushEnvCore_UpdateExistingItem(t *testing.T) {
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: &FullItem{ID: "item-456", Name: "my-project", Notes: "{}"},
	}
	fs := &mockFileSystem{
		dirEntries: []DirEntry{
			&mockDirEntry{name: ".env", isDir: false},
		},
		readContentMap: map[string][]byte{
			".env": []byte("KEY=newvalue\n"),
		},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := PushEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		logger,
	)

	assert.NoError(t, err)
	assert.Contains(t, bw.calls, "UpdateNoteItem(item-456)")
}

// 正常系: .env がなく .env.local のみの場合でも push 可能
func TestPushEnvCore_OnlyEnvLocal(t *testing.T) {
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: nil,
	}
	fs := &mockFileSystem{
		dirEntries: []DirEntry{
			&mockDirEntry{name: ".env.local", isDir: false},
		},
		readContentMap: map[string][]byte{
			".env.local": []byte("KEY=local\n"),
		},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := PushEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		logger,
	)

	assert.NoError(t, err)
	assert.Contains(t, bw.calls, "CreateNoteItem(folder-123,my-project)")
}

// 異常系: .env ファイルが見つからない
func TestPushEnvCore_EnvFileNotFound(t *testing.T) {
	bw := &mockBwClient{folderID: "folder-123"}
	fs := &mockFileSystem{
		dirEntries: []DirEntry{}, // 空のディレクトリ
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := PushEnvCore(
		"/some/path", // "." でも ".." でもないのでフォールバックしない
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		logger,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no .env files found")
}

// 異常系: GetDotenvsFolderID がエラーを返す
func TestPushEnvCore_GetFolderIDError(t *testing.T) {
	bw := &mockBwClient{
		folderIDErr: errors.New("dotenvs folder not found"),
	}
	fs := &mockFileSystem{
		dirEntries: []DirEntry{
			&mockDirEntry{name: ".env", isDir: false},
		},
		readContentMap: map[string][]byte{
			".env": []byte("KEY=value\n"),
		},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := PushEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		logger,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get dotenvs folder")
}

// 異常系: CreateNoteItem がエラーを返す
func TestPushEnvCore_CreateItemError(t *testing.T) {
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: nil,
		createErr:  errors.New("failed to create item"),
	}
	fs := &mockFileSystem{
		dirEntries: []DirEntry{
			&mockDirEntry{name: ".env", isDir: false},
		},
		readContentMap: map[string][]byte{
			".env": []byte("KEY=value\n"),
		},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := PushEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		logger,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create item")
}

// =============================================================================
// PullEnvCore のテスト
// =============================================================================

// 正常系: アイテムが存在し、出力先に .env がない場合に新規作成
func TestPullEnvCore_CreateNewEnvFile(t *testing.T) {
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: &FullItem{ID: "item-456", Name: "my-project", Notes: `{"lines":["KEY=value"]}`},
	}
	fs := &mockFileSystem{
		statInfo: &mockFileInfo{notExist: true},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := PullEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		func(path string) (bool, error) { return true, nil },
		logger,
	)

	assert.NoError(t, err)
	assert.Contains(t, fs.calls, "WriteFile(.env)")
	assert.Equal(t, "KEY=value", string(fs.writtenData))
}

// 正常系: 出力ディレクトリが存在しない場合に MkdirAll が呼ばれる
func TestPullEnvCore_CreateOutputDirectory(t *testing.T) {
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: &FullItem{ID: "item-456", Name: "my-project", Notes: `{"lines":["KEY=value"]}`},
	}
	fs := &mockFileSystem{
		statInfo: &mockFileInfo{notExist: true},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := PullEnvCore(
		"/custom/output",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		func(path string) (bool, error) { return true, nil },
		logger,
	)

	assert.NoError(t, err)
	assert.Contains(t, fs.calls, "MkdirAll(/custom/output)")
	assert.Contains(t, fs.calls, "WriteFile(/custom/output/.env)")
}

// 正常系: outputDir が "." の場合は "." のまま使用される
func TestPullEnvCore_CurrentDirOutputDir(t *testing.T) {
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: &FullItem{ID: "item-456", Name: "my-project", Notes: `{"lines":["KEY=value"]}`},
	}
	fs := &mockFileSystem{
		statInfo: &mockFileInfo{notExist: true},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := PullEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		func(path string) (bool, error) { return true, nil },
		logger,
	)

	assert.NoError(t, err)
	// .env に書き込まれることを確認（filepath.Joinは"."を省略する）
	assert.Contains(t, fs.calls, "WriteFile(.env)")
}

// 異常系: アイテムが見つからない
func TestPullEnvCore_ItemNotFound(t *testing.T) {
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: nil, // アイテムなし
	}
	fs := &mockFileSystem{}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := PullEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		func(path string) (bool, error) { return true, nil },
		logger,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "item 'my-project' not found")
}

// 異常系: confirmOverwrite が false を返す場合
func TestPullEnvCore_OverwriteCancelled(t *testing.T) {
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: &FullItem{ID: "item-456", Name: "my-project", Notes: `{".env":{"lines":["KEY=value"]}}`},
	}
	fs := &mockFileSystem{
		statInfo: &mockFileInfo{notExist: false}, // ファイルが存在する
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := PullEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		func(path string) (bool, error) { return false, nil }, // キャンセル
		logger,
	)

	// キャンセルの場合はエラーなしで終了
	assert.NoError(t, err)
	// WriteFile は呼ばれない
	for _, call := range fs.calls {
		assert.NotContains(t, call, "WriteFile")
	}
}

// 異常系: RestoreEnvFileFromJSON が壊れた JSON でエラー
func TestPullEnvCore_InvalidJSON(t *testing.T) {
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: &FullItem{ID: "item-456", Name: "my-project", Notes: "not valid json"},
	}
	fs := &mockFileSystem{
		statInfo: &mockFileInfo{notExist: true},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := PullEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		func(path string) (bool, error) { return true, nil },
		logger,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to restore .env from JSON")
}

// =============================================================================
// ListDotenvsCore のテスト
// =============================================================================

// 正常系: 3 件のアイテムが返る
func TestListDotenvsCore_ReturnsItems(t *testing.T) {
	bw := &mockBwClient{
		folderID: "folder-123",
		items: []Item{
			{ID: "1", Name: "project-a"},
			{ID: "2", Name: "project-b"},
			{ID: "3", Name: "project-c"},
		},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	items, err := ListDotenvsCore(
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		logger,
	)

	assert.NoError(t, err)
	assert.Len(t, items, 3)
	assert.Equal(t, "project-a", items[0].Name)
	assert.Equal(t, "project-b", items[1].Name)
	assert.Equal(t, "project-c", items[2].Name)
}

// 正常系: 空スライスが返る
func TestListDotenvsCore_ReturnsEmptySlice(t *testing.T) {
	bw := &mockBwClient{
		folderID: "folder-123",
		items:    []Item{},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	items, err := ListDotenvsCore(
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		logger,
	)

	assert.NoError(t, err)
	assert.Empty(t, items)
}

// 異常系: GetDotenvsFolderID がロック関連エラーを返し、リトライも失敗
func TestListDotenvsCore_FolderIDLockError(t *testing.T) {
	bw := &mockBwClient{
		folderIDErr: ErrBitwardenLocked,
		unlockErr:   errors.New("unlock failed"),
		loginErr:    errors.New("login failed"),
	}
	logger := &mockLogger{}
	cfg := &config.Config{Email: "test@example.com"}

	items, err := ListDotenvsCore(
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		logger,
	)

	assert.Error(t, err)
	assert.Nil(t, items)
}

// 異常系: ListItemsInFolder がエラーを返す
func TestListDotenvsCore_ListItemsError(t *testing.T) {
	bw := &mockBwClient{
		folderID:     "folder-123",
		listItemsErr: errors.New("failed to list items"),
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	items, err := ListDotenvsCore(
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		logger,
	)

	assert.Error(t, err)
	assert.Nil(t, items)
	assert.Contains(t, err.Error(), "failed to list items")
}

// =============================================================================
// SetupBitwardenCore のテスト
// =============================================================================

// 正常系: cloud が選択され、Login 成功、SaveConfig が呼ばれる
func TestSetupBitwardenCore_CloudSuccess(t *testing.T) {
	bw := &mockBwClient{folderExists: true}
	fs := &mockFileSystem{}
	logger := &mockLogger{}

	err := SetupBitwardenCore(
		fs,
		bw,
		logger,
		func() (string, error) { return "cloud", nil },
		func() (string, error) { return "", errors.New("should not be called") },
		func() (string, error) { return "test@example.com", nil },
		func() (string, error) { return "password123", nil },
		func() (bool, error) { return false, nil },
	)

	assert.NoError(t, err)
	assert.Contains(t, bw.calls, "Login(test@example.com,)")
}

// 正常系: selfhosted が選択され、URL 入力、Login 成功
func TestSetupBitwardenCore_SelfhostedSuccess(t *testing.T) {
	bw := &mockBwClient{folderExists: true}
	fs := &mockFileSystem{}
	logger := &mockLogger{}

	err := SetupBitwardenCore(
		fs,
		bw,
		logger,
		func() (string, error) { return "selfhosted", nil },
		func() (string, error) { return "https://bw.example.com", nil },
		func() (string, error) { return "test@example.com", nil },
		func() (string, error) { return "password123", nil },
		func() (bool, error) { return false, nil },
	)

	assert.NoError(t, err)
	assert.Contains(t, bw.calls, "Login(test@example.com,https://bw.example.com)")
}

// 異常系: selectHostType がエラーを返す
func TestSetupBitwardenCore_SelectHostTypeError(t *testing.T) {
	bw := &mockBwClient{}
	fs := &mockFileSystem{}
	logger := &mockLogger{}

	err := SetupBitwardenCore(
		fs,
		bw,
		logger,
		func() (string, error) { return "", errors.New("user cancelled") },
		func() (string, error) { return "", nil },
		func() (string, error) { return "", nil },
		func() (string, error) { return "", nil },
		func() (bool, error) { return false, nil },
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to select host type")
}

// 異常系: Login がエラーを返す場合、設定保存が行われない
func TestSetupBitwardenCore_LoginError(t *testing.T) {
	bw := &mockBwClient{
		loginErr: errors.New("invalid credentials"),
	}
	fs := &mockFileSystem{}
	logger := &mockLogger{}

	err := SetupBitwardenCore(
		fs,
		bw,
		logger,
		func() (string, error) { return "cloud", nil },
		func() (string, error) { return "", nil },
		func() (string, error) { return "test@example.com", nil },
		func() (string, error) { return "wrongpassword", nil },
		func() (bool, error) { return false, nil },
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to login")
}

// =============================================================================
// BwClient インターフェースのテスト（モック検証）
// =============================================================================

// GetDotenvsFolderID: 正常系
func TestBwClient_GetDotenvsFolderID_Success(t *testing.T) {
	bw := &mockBwClient{folderID: "test-folder-id"}

	id, err := bw.GetDotenvsFolderID()

	assert.NoError(t, err)
	assert.Equal(t, "test-folder-id", id)
	assert.Contains(t, bw.calls, "GetDotenvsFolderID")
}

// GetDotenvsFolderID: フォルダが見つからない
func TestBwClient_GetDotenvsFolderID_NotFound(t *testing.T) {
	bw := &mockBwClient{
		folderIDErr: errors.New("dotenvs folder not found"),
	}

	id, err := bw.GetDotenvsFolderID()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dotenvs folder not found")
	assert.Empty(t, id)
}

// GetDotenvsFolderID: ロック状態
func TestBwClient_GetDotenvsFolderID_Locked(t *testing.T) {
	bw := &mockBwClient{
		folderIDErr: ErrBitwardenLocked,
	}

	id, err := bw.GetDotenvsFolderID()

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrBitwardenLocked)
	assert.Empty(t, id)
}

// ListItemsInFolder: 正常系 - 3 件
func TestBwClient_ListItemsInFolder_Success(t *testing.T) {
	bw := &mockBwClient{
		items: []Item{
			{ID: "1", Name: "a"},
			{ID: "2", Name: "b"},
			{ID: "3", Name: "c"},
		},
	}

	items, err := bw.ListItemsInFolder("folder-id")

	assert.NoError(t, err)
	assert.Len(t, items, 3)
}

// ListItemsInFolder: 正常系 - 空
func TestBwClient_ListItemsInFolder_Empty(t *testing.T) {
	bw := &mockBwClient{
		items: []Item{},
	}

	items, err := bw.ListItemsInFolder("folder-id")

	assert.NoError(t, err)
	assert.Empty(t, items)
}

// ListItemsInFolder: ロック状態
func TestBwClient_ListItemsInFolder_Locked(t *testing.T) {
	bw := &mockBwClient{
		listItemsErr: ErrBitwardenLocked,
	}

	items, err := bw.ListItemsInFolder("folder-id")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrBitwardenLocked)
	assert.Nil(t, items)
}

// GetItemByName: 正常系 - アイテム発見
func TestBwClient_GetItemByName_Found(t *testing.T) {
	bw := &mockBwClient{
		itemByName: &FullItem{ID: "item-1", Name: "my-project", Notes: "{}"},
	}

	item, err := bw.GetItemByName("folder-id", "my-project")

	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, "my-project", item.Name)
}

// GetItemByName: 正常系 - アイテムなし
func TestBwClient_GetItemByName_NotFound(t *testing.T) {
	bw := &mockBwClient{
		itemByName: nil,
	}

	item, err := bw.GetItemByName("folder-id", "nonexistent")

	assert.NoError(t, err)
	assert.Nil(t, item)
}

// GetItemByName: ロック状態
func TestBwClient_GetItemByName_Locked(t *testing.T) {
	bw := &mockBwClient{
		itemByNameErr: ErrBitwardenLocked,
	}

	item, err := bw.GetItemByName("folder-id", "my-project")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrBitwardenLocked)
	assert.Nil(t, item)
}

// GetItemByID: 正常系
func TestBwClient_GetItemByID_Success(t *testing.T) {
	bw := &mockBwClient{
		itemByID: &FullItem{ID: "item-123", Name: "test", Notes: "content"},
	}

	item, err := bw.GetItemByID("item-123")

	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, "item-123", item.ID)
}

// GetItemByID: 異常系 - 空出力
func TestBwClient_GetItemByID_Empty(t *testing.T) {
	bw := &mockBwClient{
		itemByIDErr: errors.New("no output from bw get item command"),
	}

	item, err := bw.GetItemByID("item-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no output")
	assert.Nil(t, item)
}

// GetItemByID: ロック状態
func TestBwClient_GetItemByID_Locked(t *testing.T) {
	bw := &mockBwClient{
		itemByIDErr: ErrBitwardenLocked,
	}

	item, err := bw.GetItemByID("item-123")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrBitwardenLocked)
	assert.Nil(t, item)
}

// CreateNoteItem: 正常系
func TestBwClient_CreateNoteItem_Success(t *testing.T) {
	bw := &mockBwClient{}

	err := bw.CreateNoteItem("folder-id", "new-item", "{}")

	assert.NoError(t, err)
	assert.Contains(t, bw.calls, "CreateNoteItem(folder-id,new-item)")
}

// CreateNoteItem: 異常系
func TestBwClient_CreateNoteItem_Error(t *testing.T) {
	bw := &mockBwClient{
		createErr: errors.New("failed to create item"),
	}

	err := bw.CreateNoteItem("folder-id", "new-item", "{}")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create item")
}

// UpdateNoteItem: 正常系
func TestBwClient_UpdateNoteItem_Success(t *testing.T) {
	bw := &mockBwClient{}

	err := bw.UpdateNoteItem("item-123", "new notes")

	assert.NoError(t, err)
	assert.Contains(t, bw.calls, "UpdateNoteItem(item-123)")
}

// UpdateNoteItem: 異常系
func TestBwClient_UpdateNoteItem_Error(t *testing.T) {
	bw := &mockBwClient{
		updateErr: errors.New("failed to update item"),
	}

	err := bw.UpdateNoteItem("item-123", "new notes")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update item")
}

// Login: 正常系
func TestBwClient_Login_Success(t *testing.T) {
	bw := &mockBwClient{}

	err := bw.Login("test@example.com", "password", "")

	assert.NoError(t, err)
	assert.Contains(t, bw.calls, "Login(test@example.com,)")
}

// Login: 異常系
func TestBwClient_Login_Error(t *testing.T) {
	bw := &mockBwClient{
		loginErr: errors.New("invalid credentials"),
	}

	err := bw.Login("test@example.com", "wrong", "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid credentials")
}

// Unlock: 正常系
func TestBwClient_Unlock_Success(t *testing.T) {
	bw := &mockBwClient{}

	err := bw.Unlock("masterpassword")

	assert.NoError(t, err)
	assert.Contains(t, bw.calls, "Unlock")
}

// Unlock: 異常系
func TestBwClient_Unlock_Error(t *testing.T) {
	bw := &mockBwClient{
		unlockErr: errors.New("wrong master password"),
	}

	err := bw.Unlock("wrongpassword")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "wrong master password")
}

// =============================================================================
// FileSystem インターフェースのテスト（モック検証）
// =============================================================================

// WriteFile 後に同じ内容が取得できる
func TestFileSystem_WriteAndRead(t *testing.T) {
	fs := &mockFileSystem{}

	err := fs.WriteFile("/test/.env", []byte("KEY=value"), 0644)
	assert.NoError(t, err)
	assert.Equal(t, "/test/.env", fs.writtenPath)
	assert.Equal(t, []byte("KEY=value"), fs.writtenData)
}

// MkdirAll: 正常系
func TestFileSystem_MkdirAll_Success(t *testing.T) {
	fs := &mockFileSystem{}

	err := fs.MkdirAll("/test/dir", 0755)

	assert.NoError(t, err)
	assert.Contains(t, fs.calls, "MkdirAll(/test/dir)")
}

// MkdirAll: 異常系
func TestFileSystem_MkdirAll_Error(t *testing.T) {
	fs := &mockFileSystem{
		mkdirErr: errors.New("permission denied"),
	}

	err := fs.MkdirAll("/readonly/dir", 0755)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

// Stat: ファイルが存在しない
func TestFileSystem_Stat_NotExist(t *testing.T) {
	fs := &mockFileSystem{
		statInfo: &mockFileInfo{notExist: true},
	}

	info, err := fs.Stat("/nonexistent")

	assert.NoError(t, err)
	assert.True(t, info.IsNotExist())
}

// Stat: ファイルが存在する
func TestFileSystem_Stat_Exists(t *testing.T) {
	fs := &mockFileSystem{
		statInfo: &mockFileInfo{notExist: false},
	}

	info, err := fs.Stat("/existing")

	assert.NoError(t, err)
	assert.False(t, info.IsNotExist())
}

// =============================================================================
// 追加カバレッジテスト
// =============================================================================

// IsLockedError: nil エラーの場合は false
func TestIsLockedError_NilError(t *testing.T) {
	result := IsLockedError(nil)

	assert.False(t, result)
}

// IsLockedError: "Master password" を含むエラー
func TestIsLockedError_MasterPasswordCapital(t *testing.T) {
	err := errors.New("Master password required")

	result := IsLockedError(err)

	assert.True(t, result)
}

// IsLockedError: "master password" を含むエラー（小文字）
func TestIsLockedError_MasterPasswordLower(t *testing.T) {
	err := errors.New("Please enter your master password")

	result := IsLockedError(err)

	assert.True(t, result)
}

// WithUnlockRetry: Unlock 成功後に fn がエラーを返す場合
func TestWithUnlockRetry_UnlockSuccessThenFnFails(t *testing.T) {
	bw := &mockBwClient{}
	logger := &mockLogger{}
	cfg := &config.Config{Email: "test@example.com"}

	callCount := 0
	fn := func() error {
		callCount++
		if callCount == 1 {
			return ErrBitwardenLocked
		}
		return errors.New("fn failed after unlock")
	}

	err := WithUnlockRetry(
		bw,
		cfg,
		func() (string, error) { return "password", nil },
		logger,
		fn,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fn failed after unlock")
	assert.Equal(t, 2, callCount)
}

// WithUnlockRetry: cfg が nil の場合でも Unlock だけ試行
func TestWithUnlockRetry_NilConfig(t *testing.T) {
	bw := &mockBwClient{
		unlockErr: errors.New("unlock failed"),
	}
	logger := &mockLogger{}

	callCount := 0
	fn := func() error {
		callCount++
		return ErrBitwardenLocked
	}

	err := WithUnlockRetry(
		bw,
		nil, // cfg が nil
		func() (string, error) { return "password", nil },
		logger,
		fn,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unlock")
	assert.Equal(t, 1, callCount)
	assert.NotContains(t, bw.calls, "Login") // Login は呼ばれない
}

// WithUnlockRetry: cfg.Email が空の場合は Login を試行しない
func TestWithUnlockRetry_EmptyEmail(t *testing.T) {
	bw := &mockBwClient{
		unlockErr: errors.New("unlock failed"),
	}
	logger := &mockLogger{}
	cfg := &config.Config{Email: ""} // Email が空

	fn := func() error {
		return ErrBitwardenLocked
	}

	err := WithUnlockRetry(
		bw,
		cfg,
		func() (string, error) { return "password", nil },
		logger,
		fn,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unlock")
	assert.NotContains(t, bw.calls, "Login")
}

// WithUnlockRetry: Login 成功後に fn が成功する場合
func TestWithUnlockRetry_LoginThenFnSuccess(t *testing.T) {
	bw := &mockBwClientWithUnlockCount{
		mockBwClient: mockBwClient{},
	}
	logger := &mockLogger{}
	cfg := &config.Config{Email: "test@example.com", SelfhostedURL: "https://bw.example.com"}

	callCount := 0
	fn := func() error {
		callCount++
		if callCount == 1 {
			return ErrBitwardenLocked
		}
		return nil
	}

	err := WithUnlockRetry(
		bw,
		cfg,
		func() (string, error) { return "password", nil },
		logger,
		fn,
	)

	assert.NoError(t, err)
	assert.Equal(t, 2, callCount)
}

// mockBwClientWithUnlockCount は Unlock の呼び出し回数をカウントするモック
type mockBwClientWithUnlockCount struct {
	mockBwClient
	unlockCallCount int
}

func (m *mockBwClientWithUnlockCount) Unlock(masterPassword string) error {
	m.calls = append(m.calls, "Unlock")
	m.unlockCallCount++
	if m.unlockCallCount == 1 {
		return errors.New("unlock failed first time")
	}
	return nil
}

// PullEnvCore: confirmOverwrite がエラーを返す場合
func TestPullEnvCore_ConfirmOverwriteError(t *testing.T) {
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: &FullItem{ID: "item-456", Name: "my-project", Notes: `{".env":{"lines":["KEY=value"]}}`},
	}
	fs := &mockFileSystem{
		statInfo: &mockFileInfo{notExist: false}, // ファイルが存在する
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := PullEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		func(path string) (bool, error) { return false, errors.New("confirm error") },
		logger,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to confirm overwrite")
}

// PullEnvCore: MkdirAll がエラーを返す場合
func TestPullEnvCore_MkdirAllError(t *testing.T) {
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: &FullItem{ID: "item-456", Name: "my-project", Notes: `{"lines":["KEY=value"]}`},
	}
	fs := &mockFileSystem{
		statInfo: &mockFileInfo{notExist: true},
		mkdirErr: errors.New("mkdir failed"),
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := PullEnvCore(
		"/custom/output", // "." でも "/project-root" でもないので MkdirAll が呼ばれる
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		func(path string) (bool, error) { return true, nil },
		logger,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create output directory")
}

// PullEnvCore: WriteFile がエラーを返す場合
func TestPullEnvCore_WriteFileError(t *testing.T) {
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: &FullItem{ID: "item-456", Name: "my-project", Notes: `{".env":{"lines":["KEY=value"]}}`},
	}
	fs := &mockFileSystem{
		statInfo: &mockFileInfo{notExist: true},
		writeErr: errors.New("write failed"),
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := PullEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		func(path string) (bool, error) { return true, nil },
		logger,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write .env file")
}

// PullEnvCore: GetDotenvsFolderID がエラーを返す場合
func TestPullEnvCore_GetFolderIDError(t *testing.T) {
	bw := &mockBwClient{
		folderIDErr: errors.New("folder not found"),
	}
	fs := &mockFileSystem{}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := PullEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		func(path string) (bool, error) { return true, nil },
		logger,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get dotenvs folder")
}

// PullEnvCore: GetItemByName がエラーを返す場合
func TestPullEnvCore_GetItemByNameError(t *testing.T) {
	bw := &mockBwClient{
		folderID:      "folder-123",
		itemByNameErr: errors.New("get item failed"),
	}
	fs := &mockFileSystem{}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := PullEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		func(path string) (bool, error) { return true, nil },
		logger,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get item")
}

// PushEnvCore: GetItemByName がエラーを返す場合
func TestPushEnvCore_GetItemByNameError(t *testing.T) {
	bw := &mockBwClient{
		folderID:      "folder-123",
		itemByNameErr: errors.New("get item failed"),
	}
	fs := &mockFileSystem{
		dirEntries: []DirEntry{
			&mockDirEntry{name: ".env", isDir: false},
		},
		readContentMap: map[string][]byte{
			".env": []byte("KEY=value\n"),
		},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := PushEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		logger,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get item")
}

// PushEnvCore: UpdateNoteItem がエラーを返す場合
func TestPushEnvCore_UpdateItemError(t *testing.T) {
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: &FullItem{ID: "item-456", Name: "my-project", Notes: "{}"},
		updateErr:  errors.New("update failed"),
	}
	fs := &mockFileSystem{
		dirEntries: []DirEntry{
			&mockDirEntry{name: ".env", isDir: false},
		},
		readContentMap: map[string][]byte{
			".env": []byte("KEY=value\n"),
		},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := PushEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		logger,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update item")
}

// SetupBitwardenCore: inputURL がエラーを返す場合
func TestSetupBitwardenCore_InputURLError(t *testing.T) {
	bw := &mockBwClient{}
	fs := &mockFileSystem{}
	logger := &mockLogger{}

	err := SetupBitwardenCore(
		fs,
		bw,
		logger,
		func() (string, error) { return "selfhosted", nil },
		func() (string, error) { return "", errors.New("url input cancelled") },
		func() (string, error) { return "test@example.com", nil },
		func() (string, error) { return "password123", nil },
		func() (bool, error) { return false, nil },
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get URL")
}

// SetupBitwardenCore: inputEmail がエラーを返す場合
func TestSetupBitwardenCore_InputEmailError(t *testing.T) {
	bw := &mockBwClient{}
	fs := &mockFileSystem{}
	logger := &mockLogger{}

	err := SetupBitwardenCore(
		fs,
		bw,
		logger,
		func() (string, error) { return "cloud", nil },
		func() (string, error) { return "", nil },
		func() (string, error) { return "", errors.New("email input cancelled") },
		func() (string, error) { return "password123", nil },
		func() (bool, error) { return false, nil },
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get email")
}

// SetupBitwardenCore: inputPassword がエラーを返す場合
func TestSetupBitwardenCore_InputPasswordError(t *testing.T) {
	bw := &mockBwClient{}
	fs := &mockFileSystem{}
	logger := &mockLogger{}

	err := SetupBitwardenCore(
		fs,
		bw,
		logger,
		func() (string, error) { return "cloud", nil },
		func() (string, error) { return "", nil },
		func() (string, error) { return "test@example.com", nil },
		func() (string, error) { return "", errors.New("password input cancelled") },
		func() (bool, error) { return false, nil },
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get password")
}

// parseEnvContent: 空のコンテンツ
func TestParseEnvContent_Empty(t *testing.T) {
	content := []byte("")

	result := parseEnvContent(content)

	assert.NotNil(t, result)
	assert.Empty(t, result.Lines)
}

// parseEnvContent: 末尾に改行がない場合
func TestParseEnvContent_NoTrailingNewline(t *testing.T) {
	content := []byte("KEY1=value1\nKEY2=value2")

	result := parseEnvContent(content)

	assert.NotNil(t, result)
	assert.Len(t, result.Lines, 2)
	assert.Equal(t, "KEY1=value1", result.Lines[0])
	assert.Equal(t, "KEY2=value2", result.Lines[1])
}

// parseEnvContent: 末尾に改行がある場合
func TestParseEnvContent_WithTrailingNewline(t *testing.T) {
	content := []byte("KEY1=value1\nKEY2=value2\n")

	result := parseEnvContent(content)

	assert.NotNil(t, result)
	assert.Len(t, result.Lines, 2)
	assert.Equal(t, "KEY1=value1", result.Lines[0])
	assert.Equal(t, "KEY2=value2", result.Lines[1])
}

// PullEnvCore: ".." の場合も ".." のまま使用される
func TestPullEnvCore_DoubleDotOutputDir(t *testing.T) {
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: &FullItem{ID: "item-456", Name: "my-project", Notes: `{".env":{"lines":["KEY=value"]}}`},
	}
	fs := &mockFileSystem{
		statInfo: &mockFileInfo{notExist: true},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := PullEnvCore(
		"..",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		func(path string) (bool, error) { return true, nil },
		logger,
	)

	assert.NoError(t, err)
	assert.Contains(t, fs.calls, "WriteFile(../.env)")
}

// PushEnvCore: ".." の場合も ".." のまま使用される
func TestPushEnvCore_DoubleDotFromDir(t *testing.T) {
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: nil,
	}
	fs := &mockFileSystem{
		dirEntries: []DirEntry{
			&mockDirEntry{name: ".env", isDir: false},
		},
		readContentMap: map[string][]byte{
			"../.env": []byte("KEY=value\n"),
		},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := PushEnvCore(
		"..",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		logger,
	)

	assert.NoError(t, err)
	assert.Contains(t, bw.calls, "CreateNoteItem(folder-123,my-project)")
}

// =============================================================================
// MultiEnvData のテスト
// =============================================================================

// 正常系: MultiEnvData を JSON に変換して復元
func TestMultiEnvData_RoundTrip(t *testing.T) {
	original := MultiEnvData{
		".env":         EnvData{Lines: []string{"KEY1=value1", "KEY2=value2"}},
		".env.staging": EnvData{Lines: []string{"KEY1=staging1", "KEY2=staging2"}},
	}

	// JSON に変換
	jsonStr, err := multiEnvDataToJSON(original)
	assert.NoError(t, err)
	assert.Contains(t, jsonStr, `".env"`)
	assert.Contains(t, jsonStr, `".env.staging"`)

	// 復元
	restored, err := restoreMultiEnvFromJSON(jsonStr)
	assert.NoError(t, err)
	assert.Equal(t, original, restored)
}

// 正常系: restoreEnvContentFromData
func TestRestoreEnvContentFromData(t *testing.T) {
	data := EnvData{Lines: []string{"KEY1=value1", "KEY2=value2", "# comment"}}

	content := restoreEnvContentFromData(data)

	assert.Equal(t, "KEY1=value1\nKEY2=value2\n# comment", content)
}

// =============================================================================
// PushEnvCore 複数ファイル対応のテスト
// =============================================================================

// 正常系: 複数の .env ファイルを push
func TestPushEnvCore_MultipleEnvFiles(t *testing.T) {
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: nil, // 新規作成
	}
	fs := &mockFileSystem{
		dirEntries: []DirEntry{
			&mockDirEntry{name: ".env", isDir: false},
			&mockDirEntry{name: ".env.staging", isDir: false},
			&mockDirEntry{name: ".env.production", isDir: false},
		},
		readContentMap: map[string][]byte{
			".env":            []byte("KEY=base\n"),
			".env.staging":    []byte("KEY=staging\n"),
			".env.production": []byte("KEY=production\n"),
		},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := PushEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		logger,
	)

	assert.NoError(t, err)
	assert.Contains(t, bw.calls, "CreateNoteItem(folder-123,my-project)")
}

// 正常系: .example ファイルは除外される
func TestPushEnvCore_ExcludesExampleFiles(t *testing.T) {
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: nil,
	}
	fs := &mockFileSystem{
		dirEntries: []DirEntry{
			&mockDirEntry{name: ".env", isDir: false},
			&mockDirEntry{name: ".env.example", isDir: false},
			&mockDirEntry{name: ".env.staging", isDir: false},
			&mockDirEntry{name: ".env.staging.example", isDir: false},
		},
		readContentMap: map[string][]byte{
			".env":         []byte("KEY=base\n"),
			".env.staging": []byte("KEY=staging\n"),
		},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := PushEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		logger,
	)

	assert.NoError(t, err)

	// ReadFile が .example ファイルに対して呼ばれていないことを確認
	for _, call := range fs.calls {
		assert.NotContains(t, call, ".example")
	}
}

// 異常系: .env ファイルが一つもない
func TestPushEnvCore_NoEnvFiles(t *testing.T) {
	bw := &mockBwClient{folderID: "folder-123"}
	fs := &mockFileSystem{
		dirEntries: []DirEntry{
			&mockDirEntry{name: "README.md", isDir: false},
			&mockDirEntry{name: ".env.example", isDir: false}, // example は除外
		},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := PushEnvCore(
		"/some/path",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		logger,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no .env files found")
}

// =============================================================================
// PullEnvCore 複数ファイル対応のテスト
// =============================================================================

// 正常系: 複数の .env ファイルを pull
func TestPullEnvCore_MultipleEnvFiles(t *testing.T) {
	multiEnvJSON := `{".env":{"lines":["KEY=base"]},".env.staging":{"lines":["KEY=staging"]}}`
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: &FullItem{ID: "item-456", Name: "my-project", Notes: multiEnvJSON},
	}
	fs := &mockFileSystem{
		statInfo: &mockFileInfo{notExist: true},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := PullEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		func(path string) (bool, error) { return true, nil },
		logger,
	)

	assert.NoError(t, err)

	// 両方のファイルが書き込まれたことを確認（filepath.Joinは"."を省略する）
	assert.Contains(t, fs.writtenFiles, ".env")
	assert.Contains(t, fs.writtenFiles, ".env.staging")
	assert.Equal(t, "KEY=base", string(fs.writtenFiles[".env"]))
	assert.Equal(t, "KEY=staging", string(fs.writtenFiles[".env.staging"]))
}

// 正常系: 旧形式（単一ファイル）との下位互換性
func TestPullEnvCore_LegacyFormat(t *testing.T) {
	// 旧形式: {"lines":["KEY=value"]}
	legacyJSON := `{"lines":["KEY=value"]}`
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: &FullItem{ID: "item-456", Name: "my-project", Notes: legacyJSON},
	}
	fs := &mockFileSystem{
		statInfo: &mockFileInfo{notExist: true},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := PullEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		func(path string) (bool, error) { return true, nil },
		logger,
	)

	assert.NoError(t, err)
	assert.Contains(t, fs.calls, "WriteFile(.env)")
}

// 正常系: 一部のファイルの上書きをキャンセル
func TestPullEnvCore_PartialOverwriteCancel(t *testing.T) {
	multiEnvJSON := `{".env":{"lines":["KEY=base"]},".env.staging":{"lines":["KEY=staging"]}}`
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: &FullItem{ID: "item-456", Name: "my-project", Notes: multiEnvJSON},
	}
	fs := &mockFileSystem{
		statInfoMap: map[string]FileInfo{
			".env":         &mockFileInfo{notExist: false}, // 存在する
			".env.staging": &mockFileInfo{notExist: true},  // 存在しない
		},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	confirmCalls := []string{}
	err := PullEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		func(path string) (bool, error) {
			confirmCalls = append(confirmCalls, path)
			// .env は上書きをキャンセル、.env.staging は許可
			if path == ".env" {
				return false, nil
			}
			return true, nil
		},
		logger,
	)

	assert.NoError(t, err)

	// .env の上書き確認が呼ばれたことを確認
	assert.Contains(t, confirmCalls, ".env")

	// .env は書き込まれない、.env.staging は書き込まれる
	assert.NotContains(t, fs.writtenFiles, ".env")
	assert.Contains(t, fs.writtenFiles, ".env.staging")
}

// =============================================================================
// GetPushedEnvFiles / GetPulledEnvFiles のテスト
// =============================================================================

// 正常系: GetPushedEnvFiles でファイル名一覧を取得
func TestGetPushedEnvFiles_Success(t *testing.T) {
	fs := &mockFileSystem{
		dirEntries: []DirEntry{
			&mockDirEntry{name: ".env", isDir: false},
			&mockDirEntry{name: ".env.staging", isDir: false},
			&mockDirEntry{name: ".env.example", isDir: false}, // 除外される
		},
	}

	files, err := GetPushedEnvFiles(".", fs)

	assert.NoError(t, err)
	assert.Len(t, files, 2)
	assert.Contains(t, files, ".env")
	assert.Contains(t, files, ".env.staging")
	assert.NotContains(t, files, ".env.example")
}

// 正常系: GetPulledEnvFiles でファイル名一覧を取得
func TestGetPulledEnvFiles_Success(t *testing.T) {
	multiEnvJSON := `{".env":{"lines":["KEY=base"]},".env.staging":{"lines":["KEY=staging"]},".env.production":{"lines":["KEY=prod"]}}`
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: &FullItem{ID: "item-456", Name: "my-project", Notes: multiEnvJSON},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	files, err := GetPulledEnvFiles(
		"my-project",
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		logger,
	)

	assert.NoError(t, err)
	assert.Len(t, files, 3)
	// .env が最初に来ることを確認
	assert.Equal(t, ".env", files[0])
}

// 正常系: GetPulledEnvFiles で旧形式の場合
func TestGetPulledEnvFiles_LegacyFormat(t *testing.T) {
	legacyJSON := `{"lines":["KEY=value"]}`
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: &FullItem{ID: "item-456", Name: "my-project", Notes: legacyJSON},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	files, err := GetPulledEnvFiles(
		"my-project",
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		logger,
	)

	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, ".env", files[0])
}

// =============================================================================
// CleanEnvCore のテスト
// =============================================================================

func multiEnvNotes(files map[string][]string) string {
	data := make(MultiEnvData)
	for name, lines := range files {
		data[name] = EnvData{Lines: lines}
	}
	jsonStr, err := multiEnvDataToJSON(data)
	if err != nil {
		panic(err)
	}
	return jsonStr
}

// 正常系: 内容一致なら確認なしでローカル削除
func TestCleanEnvCore_MatchRemovesLocal(t *testing.T) {
	notes := multiEnvNotes(map[string][]string{".env": {"KEY=value"}})
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: &FullItem{ID: "item-456", Name: "my-project", Notes: notes},
	}
	fs := &mockFileSystem{
		dirEntries: []DirEntry{&mockDirEntry{name: ".env"}},
		readContentMap: map[string][]byte{
			".env": []byte("KEY=value"),
		},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}
	selectCalled := false

	err := CleanEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		func(mismatched []string) (CleanMismatchAction, error) {
			selectCalled = true
			return CleanActionAbort, nil
		},
		logger,
	)

	assert.NoError(t, err)
	assert.False(t, selectCalled)
	assert.Contains(t, fs.calls, "Remove(.env)")
	assert.NotContains(t, bw.calls, "UpdateNoteItem(item-456)")
}

// 正常系: 差分ありで overwrite then clean
func TestCleanEnvCore_OverwriteRemoteThenClean(t *testing.T) {
	notes := multiEnvNotes(map[string][]string{".env": {"KEY=remote"}})
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: &FullItem{ID: "item-456", Name: "my-project", Notes: notes},
	}
	fs := &mockFileSystem{
		dirEntries: []DirEntry{&mockDirEntry{name: ".env"}},
		readContentMap: map[string][]byte{
			".env": []byte("KEY=local"),
		},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := CleanEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		func(mismatched []string) (CleanMismatchAction, error) {
			assert.Contains(t, mismatched, ".env")
			return CleanActionOverwriteRemoteThenClean, nil
		},
		logger,
	)

	assert.NoError(t, err)
	assert.Contains(t, bw.calls, "UpdateNoteItem(item-456)")
	assert.Contains(t, fs.calls, "Remove(.env)")
}

// 正常系: 差分ありで remove local のみ
func TestCleanEnvCore_RemoveLocalOnly(t *testing.T) {
	notes := multiEnvNotes(map[string][]string{".env": {"KEY=remote"}})
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: &FullItem{ID: "item-456", Name: "my-project", Notes: notes},
	}
	fs := &mockFileSystem{
		dirEntries: []DirEntry{&mockDirEntry{name: ".env"}},
		readContentMap: map[string][]byte{
			".env": []byte("KEY=local"),
		},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := CleanEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		func(mismatched []string) (CleanMismatchAction, error) {
			return CleanActionRemoveLocal, nil
		},
		logger,
	)

	assert.NoError(t, err)
	assert.Contains(t, fs.calls, "Remove(.env)")
	for _, call := range bw.calls {
		assert.NotContains(t, call, "UpdateNoteItem")
		assert.NotContains(t, call, "CreateNoteItem")
	}
}

// 異常系: リモートアイテム無し
func TestCleanEnvCore_RemoteItemMissing(t *testing.T) {
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: nil,
	}
	fs := &mockFileSystem{
		dirEntries: []DirEntry{&mockDirEntry{name: ".env"}},
		readContentMap: map[string][]byte{
			".env": []byte("KEY=value"),
		},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := CleanEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		func(mismatched []string) (CleanMismatchAction, error) {
			return CleanActionRemoveLocal, nil
		},
		logger,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Empty(t, fs.removedFiles)
}

// 異常系: リモートに管理対象ファイルが空
func TestCleanEnvCore_RemoteEmptyFiles(t *testing.T) {
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: &FullItem{ID: "item-456", Name: "my-project", Notes: `{}`},
	}
	fs := &mockFileSystem{
		dirEntries: []DirEntry{&mockDirEntry{name: ".env"}},
		readContentMap: map[string][]byte{
			".env": []byte("KEY=value"),
		},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := CleanEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		func(mismatched []string) (CleanMismatchAction, error) {
			return CleanActionRemoveLocal, nil
		},
		logger,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no env files")
	assert.Empty(t, fs.removedFiles)
}

// 異常系: 差分ありで Abort
func TestCleanEnvCore_AbortOnMismatch(t *testing.T) {
	notes := multiEnvNotes(map[string][]string{".env": {"KEY=remote"}})
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: &FullItem{ID: "item-456", Name: "my-project", Notes: notes},
	}
	fs := &mockFileSystem{
		dirEntries: []DirEntry{&mockDirEntry{name: ".env"}},
		readContentMap: map[string][]byte{
			".env": []byte("KEY=local"),
		},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}

	err := CleanEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		func(mismatched []string) (CleanMismatchAction, error) {
			return CleanActionAbort, nil
		},
		logger,
	)

	assert.ErrorIs(t, err, ErrCleanAborted)
	assert.Empty(t, fs.removedFiles)
	for _, call := range bw.calls {
		assert.NotContains(t, call, "UpdateNoteItem")
	}
}

// 正常系: 複数ファイルのうち1つでも差分があれば差分扱い
func TestCleanEnvCore_OneFileDiffTreatsAsMismatch(t *testing.T) {
	notes := multiEnvNotes(map[string][]string{
		".env":         {"KEY=value"},
		".env.staging": {"ENV=staging"},
	})
	bw := &mockBwClient{
		folderID:   "folder-123",
		itemByName: &FullItem{ID: "item-456", Name: "my-project", Notes: notes},
	}
	fs := &mockFileSystem{
		dirEntries: []DirEntry{
			&mockDirEntry{name: ".env"},
			&mockDirEntry{name: ".env.staging"},
		},
		readContentMap: map[string][]byte{
			".env":         []byte("KEY=value"),
			".env.staging": []byte("ENV=changed"),
		},
	}
	logger := &mockLogger{}
	cfg := &config.Config{}
	var gotMismatched []string

	err := CleanEnvCore(
		".",
		"my-project",
		fs,
		bw,
		cfg,
		func() (string, error) { return "pwd", nil },
		func(mismatched []string) (CleanMismatchAction, error) {
			gotMismatched = mismatched
			return CleanActionAbort, nil
		},
		logger,
	)

	assert.ErrorIs(t, err, ErrCleanAborted)
	assert.Contains(t, gotMismatched, ".env.staging")
	assert.Empty(t, fs.removedFiles)
}

// =============================================================================
// isExampleFile / sortEnvFiles のテスト
// =============================================================================

// isExampleFile のテスト
func TestIsExampleFile(t *testing.T) {
	assert.True(t, isExampleFile(".env.example"))
	assert.True(t, isExampleFile(".env.staging.example"))
	assert.True(t, isExampleFile(".env.example.staging"))
	assert.False(t, isExampleFile(".env"))
	assert.False(t, isExampleFile(".env.staging"))
	assert.False(t, isExampleFile(".env.local"))
}

// sortEnvFiles のテスト
func TestSortEnvFiles(t *testing.T) {
	files := []string{
		".env.staging",
		".env.production",
		".env",
		".env.local",
	}

	sortEnvFiles(files)

	assert.Equal(t, ".env", files[0])
	assert.Equal(t, ".env.local", files[1])
	assert.Equal(t, ".env.production", files[2])
	assert.Equal(t, ".env.staging", files[3])
}

// sortFileNames のテスト
func TestSortFileNames(t *testing.T) {
	names := []string{".env.staging", ".env.production", ".env", ".env.local"}

	sortFileNames(names)

	assert.Equal(t, ".env", names[0])
	assert.Equal(t, ".env.local", names[1])
	assert.Equal(t, ".env.production", names[2])
	assert.Equal(t, ".env.staging", names[3])
}
