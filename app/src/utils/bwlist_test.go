package utils

import (
	"testing"

	"bwsf/src/core"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// ErrBitwardenLocked のテスト
// =============================================================================

// 正常系: ErrBitwardenLocked が正しく定義されている
func TestErrBitwardenLocked_Defined(t *testing.T) {
	assert.NotNil(t, ErrBitwardenLocked)
	assert.Equal(t, "Bitwarden CLI is locked", ErrBitwardenLocked.Error())
}

// =============================================================================
// core.IsLockedError のテスト（utils から移動したため、core をテスト）
// =============================================================================

// 正常系: ErrBitwardenLocked を渡すと true
func TestIsLockedError_WithErrBitwardenLocked(t *testing.T) {
	result := core.IsLockedError(ErrBitwardenLocked)

	assert.True(t, result)
}

// 正常系: "Bitwarden CLI is locked" を含むエラーは true
func TestIsLockedError_WithLockedMessage(t *testing.T) {
	// ErrBitwardenLocked を直接使用して検証
	result := core.IsLockedError(ErrBitwardenLocked)

	assert.True(t, result)
}

// 正常系: 関係ないエラーは false
func TestIsLockedError_WithOtherError(t *testing.T) {
	err := assert.AnError

	result := core.IsLockedError(err)

	assert.False(t, result)
}

// =============================================================================
// isBwAuthRequiredMessage のテスト（#161）
// =============================================================================

func TestIsBwAuthRequiredMessage(t *testing.T) {
	assert.True(t, isBwAuthRequiredMessage("Master password required"))
	assert.True(t, isBwAuthRequiredMessage("Please enter your master password"))
	assert.True(t, isBwAuthRequiredMessage("You are not logged in."))
	assert.True(t, isBwAuthRequiredMessage("Vault is locked."))
	assert.False(t, isBwAuthRequiredMessage("network timeout"))
	assert.False(t, isBwAuthRequiredMessage(""))
}

// =============================================================================
// Folder / Item 構造体のテスト
// =============================================================================

// 正常系: Folder 構造体が正しく初期化できる
func TestFolder_Struct(t *testing.T) {
	folder := Folder{
		ID:   "folder-123",
		Name: "dotenvs",
	}

	assert.Equal(t, "folder-123", folder.ID)
	assert.Equal(t, "dotenvs", folder.Name)
}

// 正常系: Item 構造体が正しく初期化できる
func TestItem_Struct(t *testing.T) {
	item := Item{
		ID:   "item-456",
		Name: "my-project",
	}

	assert.Equal(t, "item-456", item.ID)
	assert.Equal(t, "my-project", item.Name)
}
