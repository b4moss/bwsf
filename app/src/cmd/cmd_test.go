package cmd

import (
	"os"
	"testing"

	"bwsf/src/config"

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

// 正常系: backend コマンドが登録されている
func TestBackendCmd_Registered(t *testing.T) {
	assert.NotNil(t, backendCmd)
	assert.Equal(t, "backend", backendCmd.Use)

	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "backend" {
			found = true
			break
		}
	}
	assert.True(t, found, "backend command should be registered")
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

// 正常系: setup に非対話用フラグがある
func TestSetupCmd_NonInteractiveFlags(t *testing.T) {
	for _, name := range []string{"host-type", "url", "email", "password", "yes"} {
		assert.NotNil(t, setupCmd.Flags().Lookup(name), "missing flag --%s", name)
	}
}

// 異常系: 非対話フラグが一部だけのときはバリデーションエラー
func TestValidateSetupNonInteractiveFlags_Partial(t *testing.T) {
	setupHostType = "selfhosted"
	setupURL = ""
	setupEmail = ""
	setupPassword = ""
	t.Cleanup(func() {
		setupHostType = ""
		setupURL = ""
		setupEmail = ""
		setupPassword = ""
		setupYes = false
	})

	err := validateSetupNonInteractiveFlags()
	assert.Error(t, err)
}

// 正常系: selfhosted 非対話の必須フラグが揃えば OK
func TestValidateSetupNonInteractiveFlags_SelfhostedOK(t *testing.T) {
	setupHostType = "selfhosted"
	setupURL = "https://vaultwarden:80"
	setupEmail = "smoke@bwsf.local"
	setupPassword = "SmokePassw0rd!"
	t.Cleanup(func() {
		setupHostType = ""
		setupURL = ""
		setupEmail = ""
		setupPassword = ""
		setupYes = false
	})

	err := validateSetupNonInteractiveFlags()
	assert.NoError(t, err)
}

// 正常系: pull コマンドに --output フラグがある
func TestPullCmd_OutputFlag(t *testing.T) {
	flag := pullCmd.Flags().Lookup("output")
	assert.NotNil(t, flag)
	assert.Equal(t, ".", flag.DefValue)
}

// 正常系: backend コマンドに --set フラグがある
func TestBackendCmd_SetFlag(t *testing.T) {
	flag := backendCmd.Flags().Lookup("set")
	assert.NotNil(t, flag)
	assert.Equal(t, "", flag.DefValue)
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
	// backend=api では認証が bwsf auth に分離されることを案内する
	assert.Contains(t, setupCmd.Long, "bwsf auth")
}

// 正常系: backend コマンドに説明がある
func TestBackendCmd_Description(t *testing.T) {
	assert.NotEmpty(t, backendCmd.Short)
	assert.NotEmpty(t, backendCmd.Long)
}

// 正常系: auth コマンドが登録されている
func TestAuthCmd_Registered(t *testing.T) {
	assert.NotNil(t, authCmd)
	assert.Equal(t, "auth", authCmd.Use)

	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "auth" {
			found = true
			break
		}
	}
	assert.True(t, found, "auth command should be registered")
}

// 正常系: auth コマンドに --clear フラグがある
func TestAuthCmd_ClearFlag(t *testing.T) {
	flag := authCmd.Flags().Lookup("clear")
	assert.NotNil(t, flag)
	assert.Equal(t, "false", flag.DefValue)
}

// 正常系: auth コマンドに説明がある
func TestAuthCmd_Description(t *testing.T) {
	assert.NotEmpty(t, authCmd.Short)
	assert.NotEmpty(t, authCmd.Long)
}


// =============================================================================
// backend コマンドのロジックテスト
// =============================================================================

// 正常系: 設定なしでは currentBackend が api を返す
func TestCurrentBackend_DefaultAPI(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	backend, err := currentBackend()
	assert.NoError(t, err)
	assert.Equal(t, config.BackendAPI, backend)
}

// 正常系: setBackend で api に更新できる
func TestSetBackend_API(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	err := setBackend(config.BackendAPI)
	assert.NoError(t, err)

	backend, err := currentBackend()
	assert.NoError(t, err)
	assert.Equal(t, config.BackendAPI, backend)

	cfg, err := config.LoadConfig()
	assert.NoError(t, err)
	assert.Equal(t, config.BackendAPI, cfg.Backend)
}

// 正常系: setBackend で既存設定を保ったまま backend だけ更新
func TestSetBackend_PreservesOtherFields(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	err := config.SaveConfig(&config.Config{
		HostType: "cloud",
		Email:    "user@example.com",
		Backend:  config.BackendBW,
	})
	assert.NoError(t, err)

	err = setBackend(config.BackendAPI)
	assert.NoError(t, err)

	cfg, err := config.LoadConfig()
	assert.NoError(t, err)
	assert.Equal(t, "cloud", cfg.HostType)
	assert.Equal(t, "user@example.com", cfg.Email)
	assert.Equal(t, config.BackendAPI, cfg.Backend)
}

// 異常系: 不正な backend 値はエラー
func TestSetBackend_Invalid(t *testing.T) {
	err := setBackend("cli")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid backend")
}



