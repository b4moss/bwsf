package infra

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
