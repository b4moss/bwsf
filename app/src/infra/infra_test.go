package infra

import (
	"path/filepath"
	"testing"

	"bwsf/src/config"
	"bwsf/src/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// RealBwClient のテスト
// =============================================================================

// 正常系: NewBwClient が nil でないインスタンスを返す
func TestNewBwClient_ReturnsInstance(t *testing.T) {
	client := NewBwClient()

	assert.NotNil(t, client)
}

// 正常系: RealBwClient が BwClient インターフェースを実装している
func TestRealBwClient_ImplementsInterface(t *testing.T) {
	client := NewBwClient()

	// インターフェースへの代入が可能であることを確認
	var _ interface {
		GetDotenvsFolderID() (string, error)
		Login(email, password, serverURL string) error
		Unlock(masterPassword string) error
	} = client

	assert.NotNil(t, client)
}

// =============================================================================
// RealFileSystem のテスト
// =============================================================================

// 正常系: NewFileSystem が nil でないインスタンスを返す
func TestNewFileSystem_ReturnsInstance(t *testing.T) {
	fs := NewFileSystem()

	assert.NotNil(t, fs)
}

// 正常系: RealFileSystem が FileSystem インターフェースを実装している
func TestRealFileSystem_ImplementsInterface(t *testing.T) {
	fs := NewFileSystem()

	// メソッドが存在することを確認
	assert.NotNil(t, fs.OpenEnvFile)
	assert.NotNil(t, fs.ReadFile)
	assert.NotNil(t, fs.WriteFile)
	assert.NotNil(t, fs.Remove)
	assert.NotNil(t, fs.Stat)
	assert.NotNil(t, fs.MkdirAll)
}

// =============================================================================
// RealLogger のテスト
// =============================================================================

// 正常系: NewLogger が nil でないインスタンスを返す
func TestNewLogger_ReturnsInstance(t *testing.T) {
	logger := NewLogger()

	assert.NotNil(t, logger)
}

// 正常系: RealLogger が Logger インターフェースを実装している
func TestRealLogger_ImplementsInterface(t *testing.T) {
	logger := NewLogger()

	// メソッドが存在することを確認
	assert.NotNil(t, logger.Error)
	assert.NotNil(t, logger.Info)
	assert.NotNil(t, logger.Success)
	assert.NotNil(t, logger.Warning)
}

// =============================================================================
// LoginError / UnlockError のテスト
// =============================================================================

// 正常系: LoginError が error インターフェースを実装している
func TestLoginError_ImplementsError(t *testing.T) {
	err := &LoginError{Message: "login failed"}

	assert.Equal(t, "login failed", err.Error())
}

// 正常系: UnlockError が error インターフェースを実装している
func TestUnlockError_ImplementsError(t *testing.T) {
	err := &UnlockError{Message: "unlock failed"}

	assert.Equal(t, "unlock failed", err.Error())
}

// =============================================================================
// NewBwClientFromConfig / ApiBwClient のテスト
// =============================================================================

// testConfig builds a minimal v2 global config with one default host.
func testConfig(hostType, url, email, folder string) *config.Config {
	t := config.HostTypeCloud
	if hostType == "selfhosted" || hostType == config.HostTypeSelfhost {
		t = config.HostTypeSelfhost
	}
	if url == "" && t == config.HostTypeCloud {
		url = config.DefaultCloudURL
	}
	if folder == "" {
		folder = config.DefaultFolderName
	}
	return &config.Config{
		SchemaVersion: config.SchemaVersion1,
		Settings: config.GlobalSettings{
			Hosts: []config.Host{{
				ID:            config.DefaultHostID,
				Type:          t,
				HostURL:       url,
				Email:         email,
				TargetSection: folder,
				IsDefault:     true,
			}},
		},
	}
}

// 正常系: factory は常に API アダプタを返す
func TestNewBwClientFromConfig_AlwaysAPI(t *testing.T) {
	client, err := NewBwClientFromConfig(nil)
	assert.NoError(t, err)
	assert.IsType(t, &ApiBwClient{}, client)

	client, err = NewBwClientFromConfig(&config.Config{})
	assert.NoError(t, err)
	assert.IsType(t, &ApiBwClient{}, client)

	client, err = NewBwClientFromConfig(testConfig("cloud", "", "a@example.com", ""))
	assert.NoError(t, err)
	assert.IsType(t, &ApiBwClient{}, client)
}

// 異常系: 未認証では保管庫メソッドが auth エラーになる
func TestApiBwClient_VaultRequiresAuth(t *testing.T) {
	client := NewApiBwClientWithDeps(
		testConfig("cloud", "", "", ""),
		NewMemorySecretStore(),
		NewIdentityClient(),
		nil,
	)

	_, err := client.GetDotenvsFolderID()
	assert.ErrorIs(t, err, ErrAPINotAuthenticated)

	_, err = client.ListItemsInFolder("id")
	assert.ErrorIs(t, err, ErrAPINotAuthenticated)

	assert.ErrorIs(t, client.Login("e", "p", ""), ErrAPINotAuthenticated)
	assert.Contains(t, ErrAPINotImplemented.Error(), "bwsf auth login")
}

// 正常系: ApiBwClient が BwClient インターフェースを実装している
func TestApiBwClient_ImplementsInterface(t *testing.T) {
	var _ core.BwClient = NewApiBwClient(nil)
}

// =============================================================================
// realFileInfo のテスト
// =============================================================================

// 正常系: realFileInfo.IsNotExist() が true を返す
func TestRealFileInfo_IsNotExist_True(t *testing.T) {
	info := &realFileInfo{notExist: true}

	assert.True(t, info.IsNotExist())
}

// 正常系: realFileInfo.IsNotExist() が false を返す
func TestRealFileInfo_IsNotExist_False(t *testing.T) {
	info := &realFileInfo{notExist: false}

	assert.False(t, info.IsNotExist())
}

func TestRealFileSystem_CRUD(t *testing.T) {
	dir := t.TempDir()
	fs := NewFileSystem()
	path := filepath.Join(dir, "a.env")

	require.NoError(t, fs.WriteFile(path, []byte("A=1\n"), 0o644))
	data, err := fs.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "A=1\n", string(data))

	data, err = fs.OpenEnvFile(path)
	require.NoError(t, err)
	assert.Equal(t, "A=1\n", string(data))

	info, err := fs.Stat(path)
	require.NoError(t, err)
	assert.False(t, info.IsNotExist())

	missing, err := fs.Stat(filepath.Join(dir, "missing"))
	require.NoError(t, err)
	assert.True(t, missing.IsNotExist())

	sub := filepath.Join(dir, "sub")
	require.NoError(t, fs.MkdirAll(sub, 0o755))
	require.NoError(t, fs.WriteFile(filepath.Join(sub, "b.env"), []byte("B=2\n"), 0o644))
	entries, err := fs.ReadDir(dir)
	require.NoError(t, err)
	names := map[string]bool{}
	for _, e := range entries {
		names[e.Name()] = true
	}
	assert.True(t, names["a.env"])
	assert.True(t, names["sub"])

	require.NoError(t, fs.Remove(path))
	missing2, err := fs.Stat(path)
	require.NoError(t, err)
	assert.True(t, missing2.IsNotExist())
}

func TestRealLogger_Methods(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	l := NewLogger()
	l.Error("err")
	l.Info("info")
	l.Success("ok")
	l.Warning("warn")
}
