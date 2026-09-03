package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// isColorEnabled のテスト
// =============================================================================

// 正常系: NO_COLOR が設定されている場合は false
func TestIsColorEnabled_NoColorSet(t *testing.T) {
	origNoColor := os.Getenv("NO_COLOR")
	os.Setenv("NO_COLOR", "1")
	defer func() {
		if origNoColor == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", origNoColor)
		}
	}()

	result := isColorEnabled()

	assert.False(t, result)
}

// 正常系: NO_COLOR が設定されていない場合はターミナル判定に依存
func TestIsColorEnabled_NoColorNotSet(t *testing.T) {
	origNoColor := os.Getenv("NO_COLOR")
	os.Unsetenv("NO_COLOR")
	defer func() {
		if origNoColor != "" {
			os.Setenv("NO_COLOR", origNoColor)
		}
	}()

	// テスト環境では通常ターミナルではないので false になる可能性が高い
	// ただし環境依存のため、エラーにならないことだけ確認
	_ = isColorEnabled()
}

// =============================================================================
// isTerminal のテスト
// =============================================================================

// 正常系: 通常のファイルはターミナルではない
func TestIsTerminal_RegularFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	result := isTerminal(tmpFile)

	assert.False(t, result)
}

// =============================================================================
// colorize のテスト
// =============================================================================

// 正常系: カラーが無効な場合はそのまま返す
func TestColorize_ColorDisabled(t *testing.T) {
	origNoColor := os.Getenv("NO_COLOR")
	os.Setenv("NO_COLOR", "1")
	defer func() {
		if origNoColor == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", origNoColor)
		}
	}()

	result := colorize("test message", colorRed)

	assert.Equal(t, "test message", result)
}

// =============================================================================
// Color* 関数のテスト
// =============================================================================

// 正常系: ColorError が文字列を返す
func TestColorError(t *testing.T) {
	origNoColor := os.Getenv("NO_COLOR")
	os.Setenv("NO_COLOR", "1")
	defer func() {
		if origNoColor == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", origNoColor)
		}
	}()

	result := ColorError("error message")

	assert.Equal(t, "error message", result)
}

// 正常系: ColorSuccess が文字列を返す
func TestColorSuccess(t *testing.T) {
	origNoColor := os.Getenv("NO_COLOR")
	os.Setenv("NO_COLOR", "1")
	defer func() {
		if origNoColor == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", origNoColor)
		}
	}()

	result := ColorSuccess("success message")

	assert.Equal(t, "success message", result)
}

// 正常系: ColorWarning が文字列を返す
func TestColorWarning(t *testing.T) {
	origNoColor := os.Getenv("NO_COLOR")
	os.Setenv("NO_COLOR", "1")
	defer func() {
		if origNoColor == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", origNoColor)
		}
	}()

	result := ColorWarning("warning message")

	assert.Equal(t, "warning message", result)
}

// 正常系: ColorInfo が文字列を返す
func TestColorInfo(t *testing.T) {
	origNoColor := os.Getenv("NO_COLOR")
	os.Setenv("NO_COLOR", "1")
	defer func() {
		if origNoColor == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", origNoColor)
		}
	}()

	result := ColorInfo("info message")

	assert.Equal(t, "info message", result)
}

// 正常系: ColorQuestion が文字列を返す
func TestColorQuestion(t *testing.T) {
	origNoColor := os.Getenv("NO_COLOR")
	os.Setenv("NO_COLOR", "1")
	defer func() {
		if origNoColor == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", origNoColor)
		}
	}()

	result := ColorQuestion("question message")

	assert.Equal(t, "question message", result)
}






func TestPrintHelpers_NoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	// Smoke: should not panic when writing to stdout/stderr
	Error("e %s", "x")
	Errorln("eln")
	Success("s %s", "x")
	Successln("sln")
	Warning("w %s", "x")
	Warningln("wln")
	Info("i %s", "x")
	Infoln("iln")
	Question("q %s", "x")
	Questionln("qln")
}
