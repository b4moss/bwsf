package e2e

import (
	"encoding/json"
	"strings"
	"testing"

	"bwsf/src/config"
	"bwsf/src/core"
	"bwsf/src/infra"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// E2Eテスト
// bwsfの主要コマンドフロー（setup, list, push, pull）をモックでテスト
// =============================================================================

// TestE2E_FullWorkflow は完全なワークフローをテストします。
// 1. Setup (ログイン)
// 2. Push (.envファイルをアップロード)
// 3. List (アイテム一覧を取得)
// 4. Pull (.envファイルをダウンロード)
func TestE2E_FullWorkflow(t *testing.T) {
	// モックを作成
	bw := infra.NewMockBwClient()
	fs := infra.NewMockFileSystem()
	logger := infra.NewMockLogger()

	// テストデータをセットアップ
	bw.SetupTestData()

	// テスト用の.envファイルを作成
	testEnvContent := `DATABASE_URL=postgres://localhost/test
API_KEY=secret-key-123
DEBUG=true`
	fs.SetFile("/project/.env", []byte(testEnvContent))

	// テスト用の設定
	cfg := &config.Config{
		HostType:      "selfhosted",
		SelfhostedURL: "https://vault.example.com",
		Email:         "test@example.com",
	}

	// パスワード入力のモック
	promptPassword := func() (string, error) {
		return "testpassword", nil
	}

	// 上書き確認のモック（常にYes）
	confirmOverwrite := func(path string) (bool, error) {
		return true, nil
	}

	projectName := "test-project"

	// =========================================================================
	// Step 1: Push - .envファイルをアップロード
	// =========================================================================
	t.Run("Push", func(t *testing.T) {
		err := core.PushEnvCore(
			"/project",
			projectName,
			fs,
			bw,
			cfg,
			promptPassword,
			logger,
		nil,
	)
		require.NoError(t, err, "PushEnvCore should succeed")

		// アイテムが作成されたことを確認
		assert.Equal(t, 1, bw.GetItemCount(), "Should have 1 item after push")
	})

	// =========================================================================
	// Step 2: List - アイテム一覧を取得
	// =========================================================================
	t.Run("List", func(t *testing.T) {
		items, err := core.ListDotenvsCore(
			bw,
			cfg,
			promptPassword,
			logger,
		nil,
	)
		require.NoError(t, err, "ListDotenvsCore should succeed")

		// アイテムが1つあることを確認
		require.Len(t, items, 1, "Should have 1 item")
		assert.Equal(t, projectName, items[0].Name, "Item name should match project name")
	})

	// =========================================================================
	// Step 3: Pull - .envファイルをダウンロード
	// =========================================================================
	t.Run("Pull", func(t *testing.T) {
		outputDir := "/output"

		err := core.PullEnvCore(
			outputDir,
			projectName,
			fs,
			bw,
			cfg,
			promptPassword,
			confirmOverwrite,
			logger,
		nil,
	)
		require.NoError(t, err, "PullEnvCore should succeed")

		// .envファイルが作成されたことを確認
		pulledContent, ok := fs.GetFile("/output/.env")
		require.True(t, ok, ".env file should be created")

		// 内容を検証
		assert.Contains(t, string(pulledContent), "DATABASE_URL=postgres://localhost/test")
		assert.Contains(t, string(pulledContent), "API_KEY=secret-key-123")
		assert.Contains(t, string(pulledContent), "DEBUG=true")
	})
}

// TestE2E_PushUpdate は既存アイテムの更新をテストします。
func TestE2E_PushUpdate(t *testing.T) {
	bw := infra.NewMockBwClient()
	fs := infra.NewMockFileSystem()
	logger := infra.NewMockLogger()

	bw.SetupTestData()

	cfg := &config.Config{
		HostType: "cloud",
		Email:    "test@example.com",
	}

	promptPassword := func() (string, error) {
		return "testpassword", nil
	}

	projectName := "update-test"

	// 最初のPush
	fs.SetFile("/project/.env", []byte("KEY=value1"))
	err := core.PushEnvCore("/project", projectName, fs, bw, cfg, promptPassword, logger, nil)
	require.NoError(t, err)

	// 2回目のPush（更新）
	fs.SetFile("/project/.env", []byte("KEY=value2\nNEW_KEY=newvalue"))
	err = core.PushEnvCore("/project", projectName, fs, bw, cfg, promptPassword, logger, nil)
	require.NoError(t, err)

	// アイテム数は変わらないはず（更新なので）
	assert.Equal(t, 1, bw.GetItemCount(), "Should still have 1 item after update")

	// Pullして内容を確認
	confirmOverwrite := func(path string) (bool, error) { return true, nil }
	err = core.PullEnvCore("/output", projectName, fs, bw, cfg, promptPassword, confirmOverwrite, logger, nil)
	require.NoError(t, err)

	pulledContent, _ := fs.GetFile("/output/.env")
	assert.Contains(t, string(pulledContent), "KEY=value2")
	assert.Contains(t, string(pulledContent), "NEW_KEY=newvalue")
}

// TestE2E_MultipleProjects は複数プロジェクトの管理をテストします。
func TestE2E_MultipleProjects(t *testing.T) {
	bw := infra.NewMockBwClient()
	fs := infra.NewMockFileSystem()
	logger := infra.NewMockLogger()

	bw.SetupTestData()

	cfg := &config.Config{
		HostType: "cloud",
		Email:    "test@example.com",
	}

	promptPassword := func() (string, error) {
		return "testpassword", nil
	}

	// 複数のプロジェクトをPush
	projects := []struct {
		name    string
		content string
	}{
		{"project-a", "ENV=production\nAPI_KEY=key-a"},
		{"project-b", "ENV=staging\nAPI_KEY=key-b"},
		{"project-c", "ENV=development\nAPI_KEY=key-c"},
	}

	for _, p := range projects {
		fs.SetFile("/project/.env", []byte(p.content))
		err := core.PushEnvCore("/project", p.name, fs, bw, cfg, promptPassword, logger, nil)
		require.NoError(t, err, "Push should succeed for %s", p.name)
	}

	// Listで全てのプロジェクトが見えることを確認
	items, err := core.ListDotenvsCore(bw, cfg, promptPassword, logger, nil)
	require.NoError(t, err)
	assert.Len(t, items, 3, "Should have 3 projects")

	// 各プロジェクトをPullして内容を確認
	confirmOverwrite := func(path string) (bool, error) { return true, nil }
	for _, p := range projects {
		err := core.PullEnvCore("/output", p.name, fs, bw, cfg, promptPassword, confirmOverwrite, logger, nil)
		require.NoError(t, err, "Pull should succeed for %s", p.name)

		pulledContent, _ := fs.GetFile("/output/.env")
		assert.Contains(t, string(pulledContent), strings.Split(p.content, "\n")[0])
	}
}

// TestE2E_PullNotFound は存在しないプロジェクトのPullをテストします。
func TestE2E_PullNotFound(t *testing.T) {
	bw := infra.NewMockBwClient()
	fs := infra.NewMockFileSystem()
	logger := infra.NewMockLogger()

	bw.SetupTestData()

	cfg := &config.Config{
		HostType: "cloud",
		Email:    "test@example.com",
	}

	promptPassword := func() (string, error) {
		return "testpassword", nil
	}

	confirmOverwrite := func(path string) (bool, error) {
		return true, nil
	}

	// 存在しないプロジェクトをPull
	err := core.PullEnvCore("/output", "nonexistent-project", fs, bw, cfg, promptPassword, confirmOverwrite, logger, nil)
	assert.Error(t, err, "Pull should fail for nonexistent project")
	assert.Contains(t, err.Error(), "not found")
}

// TestE2E_EnvDataFormat は.envファイルのJSON変換をテストします。
func TestE2E_EnvDataFormat(t *testing.T) {
	bw := infra.NewMockBwClient()
	fs := infra.NewMockFileSystem()
	logger := infra.NewMockLogger()

	bw.SetupTestData()

	cfg := &config.Config{
		HostType: "cloud",
		Email:    "test@example.com",
	}

	promptPassword := func() (string, error) {
		return "testpassword", nil
	}

	// 複雑な.envファイル
	complexEnv := `# Database settings
DATABASE_URL=postgres://user:pass@localhost:5432/db
DATABASE_POOL_SIZE=10

# API settings
API_KEY=sk-123456789
API_SECRET="secret with spaces"

# Feature flags
FEATURE_NEW_UI=true
FEATURE_BETA=false`

	fs.SetFile("/project/.env", []byte(complexEnv))

	err := core.PushEnvCore("/project", "complex-project", fs, bw, cfg, promptPassword, logger, nil)
	require.NoError(t, err)

	// 内部でJSONに変換されていることを確認（GetItemByNameで取得）
	folderID, _ := bw.GetDotenvsFolderID()
	item, err := bw.GetItemByName(folderID, "complex-project")
	require.NoError(t, err)
	require.NotNil(t, item)

	// NotesがMultiEnvData形式のJSON形式であることを確認
	var multiEnvData core.MultiEnvData
	err = json.Unmarshal([]byte(item.Notes), &multiEnvData)
	require.NoError(t, err, "Notes should be valid JSON")
	assert.Contains(t, multiEnvData, ".env", "Should have .env key")
	assert.Greater(t, len(multiEnvData[".env"].Lines), 0, "Should have lines")

	// Pullして内容が復元されることを確認
	confirmOverwrite := func(path string) (bool, error) { return true, nil }
	err = core.PullEnvCore("/output", "complex-project", fs, bw, cfg, promptPassword, confirmOverwrite, logger, nil)
	require.NoError(t, err)

	pulledContent, _ := fs.GetFile("/output/.env")
	assert.Contains(t, string(pulledContent), "# Database settings")
	assert.Contains(t, string(pulledContent), "DATABASE_URL=postgres://user:pass@localhost:5432/db")
	assert.Contains(t, string(pulledContent), `API_SECRET="secret with spaces"`)
}

// TestE2E_LockedVault はロック状態のVaultへのアクセスをテストします。
func TestE2E_LockedVault(t *testing.T) {
	bw := infra.NewMockBwClient()
	fs := infra.NewMockFileSystem()
	logger := infra.NewMockLogger()

	bw.SetupTestData()
	bw.SetUnlocked(false) // ロック状態にする

	cfg := &config.Config{
		HostType: "cloud",
		Email:    "test@example.com",
	}

	// パスワード入力でアンロック
	promptPassword := func() (string, error) {
		bw.SetUnlocked(true) // アンロックする
		return "testpassword", nil
	}

	fs.SetFile("/project/.env", []byte("KEY=value"))

	// ロック状態でもPushが成功する（自動アンロック）
	err := core.PushEnvCore("/project", "locked-test", fs, bw, cfg, promptPassword, logger, nil)
	require.NoError(t, err, "Push should succeed after unlock")
}

