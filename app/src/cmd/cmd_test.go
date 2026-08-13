package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Execute のテスト
// =============================================================================

// 注: Execute は os.Exit を呼び出すため、直接テストするのは難しい
// ここでは rootCmd が正しく初期化されていることを確認

// 正常系: rootCmd が正しく初期化されている
func TestRootCmd_Initialized(t *testing.T) {
	assert.NotNil(t, rootCmd)
	assert.Equal(t, "bwsf", rootCmd.Use)
}

// 正常系: rootCmd に Short/Long が設定されている
func TestRootCmd_Description(t *testing.T) {
	assert.NotEmpty(t, rootCmd.Short)
	assert.NotEmpty(t, rootCmd.Long)
}

// =============================================================================
// サブコマンドの登録テスト
// =============================================================================

// 正常系: push コマンドが登録されている
func TestPushCmd_Registered(t *testing.T) {
	assert.NotNil(t, pushCmd)
	assert.Equal(t, "push", pushCmd.Use)

	// rootCmd のサブコマンドとして登録されているか確認
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "push" {
			found = true
			break
		}
	}
	assert.True(t, found, "push command should be registered")
}

// 正常系: pull コマンドが登録されている
func TestPullCmd_Registered(t *testing.T) {
	assert.NotNil(t, pullCmd)
	assert.Equal(t, "pull", pullCmd.Use)

	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "pull" {
			found = true
			break
		}
	}
	assert.True(t, found, "pull command should be registered")
}

// 正常系: list コマンドが登録されている
func TestListCmd_Registered(t *testing.T) {
	assert.NotNil(t, listCmd)
	assert.Equal(t, "list", listCmd.Use)

	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "list" {
			found = true
			break
		}
	}
	assert.True(t, found, "list command should be registered")
}

// 正常系: setup コマンドが登録されている
func TestSetupCmd_Registered(t *testing.T) {
	assert.NotNil(t, setupCmd)
	assert.Equal(t, "setup", setupCmd.Use)

	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "setup" {
			found = true
			break
		}
	}
	assert.True(t, found, "setup command should be registered")
}

// 正常系: clean コマンドが登録されている
func TestCleanCmd_Registered(t *testing.T) {
	assert.NotNil(t, cleanCmd)
	assert.Equal(t, "clean", cleanCmd.Use)

	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "clean" {
			found = true
			break
		}
	}
	assert.True(t, found, "clean command should be registered")
}

// =============================================================================
// フラグのテスト
// =============================================================================

// 正常系: push コマンドに --from フラグがある
func TestPushCmd_FromFlag(t *testing.T) {
	flag := pushCmd.Flags().Lookup("from")
	assert.NotNil(t, flag)
	assert.Equal(t, ".", flag.DefValue)
}

// 正常系: pull コマンドに --output フラグがある
func TestPullCmd_OutputFlag(t *testing.T) {
	flag := pullCmd.Flags().Lookup("output")
	assert.NotNil(t, flag)
	assert.Equal(t, ".", flag.DefValue)
}

// 正常系: clean コマンドに --from フラグがある
func TestCleanCmd_FromFlag(t *testing.T) {
	flag := cleanCmd.Flags().Lookup("from")
	assert.NotNil(t, flag)
	assert.Equal(t, ".", flag.DefValue)
}

// =============================================================================
// コマンドの Short/Long 説明のテスト
// =============================================================================

// 正常系: push コマンドに説明がある
func TestPushCmd_Description(t *testing.T) {
	assert.NotEmpty(t, pushCmd.Short)
	assert.NotEmpty(t, pushCmd.Long)
}

// 正常系: pull コマンドに説明がある
func TestPullCmd_Description(t *testing.T) {
	assert.NotEmpty(t, pullCmd.Short)
	assert.NotEmpty(t, pullCmd.Long)
}

// 正常系: list コマンドに説明がある
func TestListCmd_Description(t *testing.T) {
	assert.NotEmpty(t, listCmd.Short)
	assert.NotEmpty(t, listCmd.Long)
}

// 正常系: setup コマンドに説明がある
func TestSetupCmd_Description(t *testing.T) {
	assert.NotEmpty(t, setupCmd.Short)
	assert.NotEmpty(t, setupCmd.Long)
}

// 正常系: clean コマンドに説明がある
func TestCleanCmd_Description(t *testing.T) {
	assert.NotEmpty(t, cleanCmd.Short)
	assert.NotEmpty(t, cleanCmd.Long)
}
