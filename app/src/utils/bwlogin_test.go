package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// CheckBwCommand のテスト
// =============================================================================

// 注: これらのテストは実際の bw コマンドの存在に依存するため、
// CI 環境では条件付きでスキップするか、モック化が必要

// 正常系: bw コマンドが存在する場合（環境依存）
func TestCheckBwCommand_Exists(t *testing.T) {
	// 実際の環境で bw がインストールされているかどうかで結果が変わる
	installed, path := CheckBwCommand()

	if installed {
		assert.NotEmpty(t, path)
	} else {
		assert.Empty(t, path)
	}
}

// =============================================================================
// BwLogin の構造テスト
// =============================================================================

// 注: BwLogin は実際の bw コマンドを呼び出すため、
// 単体テストではモック化が必要。ここでは関数シグネチャの確認のみ

// 正常系: 関数が正しいシグネチャを持つ
func TestBwLogin_Signature(t *testing.T) {
	// BwLogin の型を確認（コンパイル時チェック）
	var fn func(email, password, serverURL string) (bool, string) = BwLogin
	assert.NotNil(t, fn)
}

func TestLooksLikeSessionKey(t *testing.T) {
	valid := "P4tHpDULkFR5+NLL1lbfxD43q9NqIS2tmKnG0GMAn/Ft8w4JOipXty4uY4EQ5/gkTXDPGpidXuoC155F65X5sQ=="
	assert.True(t, looksLikeSessionKey(valid))
	assert.False(t, looksLikeSessionKey(""))
	assert.False(t, looksLikeSessionKey("short"))
	assert.False(t, looksLikeSessionKey("Warning: Provided passwordenv BW_PASSWORD is not set"))
	assert.False(t, looksLikeSessionKey("has spaces in the middle of a long enough stringxxxxxxxxxxxx"))
}

func TestExtractSessionKey(t *testing.T) {
	valid := "P4tHpDULkFR5+NLL1lbfxD43q9NqIS2tmKnG0GMAn/Ft8w4JOipXty4uY4EQ5/gkTXDPGpidXuoC155F65X5sQ=="
	assert.Equal(t, valid, extractSessionKey(valid))
	assert.Equal(t, valid, extractSessionKey("Warning: something\n"+valid+"\n"))
	assert.Equal(t, "", extractSessionKey("Warning: Provided passwordenv BW_PASSWORD is not set"))
	assert.Equal(t, "", extractSessionKey(""))
}

func TestEnvironWithout(t *testing.T) {
	t.Setenv("BWSF_TEST_KEEP", "1")
	t.Setenv("BWSF_TEST_DROP", "2")
	out := environWithout("BWSF_TEST_DROP")
	joined := strings.Join(out, "\n")
	assert.Contains(t, joined, "BWSF_TEST_KEEP=1")
	assert.NotContains(t, joined, "BWSF_TEST_DROP=2")
}

func TestTruncateForErr(t *testing.T) {
	assert.Equal(t, "empty output", truncateForErr("   "))
	assert.Equal(t, "hello", truncateForErr(" hello "))
	long := strings.Repeat("a", 130)
	got := truncateForErr(long)
	assert.True(t, strings.HasSuffix(got, "..."))
	assert.Equal(t, 123, len(got))
}
