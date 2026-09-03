package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"bwsf/src/config"
)

// ErrCleanAborted は差分あり時にユーザーが Abort を選んだことを表します。
var ErrCleanAborted = errors.New("clean aborted by user")

// BwClient は Bitwarden CLI とのやり取りを抽象化するインターフェースです。
type BwClient interface {
	GetDotenvsFolderID() (string, error)
	DotenvsFolderExists() (bool, error)
	CreateDotenvsFolder() error
	ListItemsInFolder(folderID string) ([]Item, error)
	GetItemByName(folderID, name string) (*FullItem, error)
	GetItemByID(id string) (*FullItem, error)
	CreateNoteItem(folderID, name, notes string) error
	UpdateNoteItem(id, notes string) error
	Login(email, password, serverURL string) error
	Unlock(masterPassword string) error
}

// FileSystem はファイルシステム操作を抽象化するインターフェースです。
type FileSystem interface {
	OpenEnvFile(path string) ([]byte, error)
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm uint32) error
	Remove(path string) error
	Stat(path string) (FileInfo, error)
	MkdirAll(path string, perm uint32) error
	ReadDir(path string) ([]DirEntry, error)
}

// CleanMismatchAction はローカル/リモート差分時のユーザー選択です。
type CleanMismatchAction int

const (
	// CleanActionAbort は削除を中止します（安全側の既定）。
	CleanActionAbort CleanMismatchAction = iota
	// CleanActionOverwriteRemoteThenClean はリモートをローカルで上書きしてからローカルを削除します。
	CleanActionOverwriteRemoteThenClean
	// CleanActionRemoveLocal はリモートを更新せずローカルのみ削除します（危険）。
	CleanActionRemoveLocal
)

// DirEntry はディレクトリエントリを表します。
type DirEntry interface {
	Name() string
	IsDir() bool
}

// FileInfo は Stat の結果に必要な最小限の情報を表します。
type FileInfo interface {
	IsNotExist() bool
}

// Logger はログ出力を抽象化するインターフェースです。
type Logger interface {
	Error(args ...interface{})
	Info(args ...interface{})
}

// SessionStore は BW_SESSION（セッションキー）の永続化を抽象化します。
// マスターパスワードは扱いません。nil は no-op として扱います。
type SessionStore interface {
	Get() (string, error)
	Set(session string) error
	Delete() error
}

type noopSessionStore struct{}

func (noopSessionStore) Get() (string, error) { return "", nil }
func (noopSessionStore) Set(string) error     { return nil }
func (noopSessionStore) Delete() error        { return nil }

func resolveSessionStore(s SessionStore) SessionStore {
	if s == nil {
		return noopSessionStore{}
	}
	return s
}

// Item は dotenvs フォルダ内に保存される Bitwarden アイテムを表します。
type Item struct {
	ID   string
	Name string
}

// FullItem は Bitwarden アイテムの完全な情報を表します。
type FullItem struct {
	ID    string
	Name  string
	Notes string
}

// EnvData は .env ファイルのデータを表します。
type EnvData struct {
	Lines []string `json:"lines"`
}

// MultiEnvData は複数の .env ファイルのデータを表します。
// キーはファイル名（例: ".env", ".env.staging"）
type MultiEnvData map[string]EnvData

// IsLockedError はエラーが認証リカバリ対象（ロック／未ログイン）かどうかを判定します。
// 現行 bw CLI の "Vault is locked." / "You are not logged in." も含みます（#161）。
func IsLockedError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "Bitwarden CLI is locked") ||
		strings.Contains(errMsg, "Master password") ||
		strings.Contains(errMsg, "master password") ||
		strings.Contains(errMsg, "You are not logged in") ||
		strings.Contains(errMsg, "Vault is locked")
}

// IsNotAuthenticatedError は API 未認証（bwsf auth が必要）かを判定します。
func IsNotAuthenticatedError(err error) bool {
	if err == nil {
		return false
	}
	if strings.Contains(err.Error(), "API backend is not authenticated") {
		return true
	}
	return false
}

// IsNotUnlockedError は復号鍵が無い／unlock が必要かを判定します。
func IsNotUnlockedError(err error) bool {
	if err == nil {
		return false
	}
	if IsLockedError(err) {
		return true
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "API vault is locked") ||
		strings.Contains(errMsg, "Enter your master password to unlock")
}

// WithUnlockRetry は Bitwarden がロックされている場合に Unlock/Login を挟んでリトライする共通処理です。
// SessionStore がある場合、BW_SESSION の自動 restore / save を行います（#130）。
// 認証切れはパスワードでは回復できないため、プロンプトせずそのまま返します。
func WithUnlockRetry(
	bw BwClient,
	cfg *config.Config,
	promptPassword func() (string, error),
	logger Logger,
	sessions SessionStore,
	fn func() error,
) error {
	sessions = resolveSessionStore(sessions)
	restoredFromStore := false

	// 環境変数 BW_SESSION が既にあれば優先し、SessionStore には触れない。
	if strings.TrimSpace(os.Getenv("BW_SESSION")) == "" {
		stored, getErr := sessions.Get()
		if getErr != nil {
			logger.Info("failed to restore session from store:", getErr)
		} else if session := strings.TrimSpace(stored); session != "" {
			_ = os.Setenv("BW_SESSION", session)
			restoredFromStore = true
		}
	}

	err := fn()
	if err == nil {
		return nil
	}

	if IsNotAuthenticatedError(err) {
		return err
	}

	if !IsNotUnlockedError(err) {
		return err
	}

	// ストアから復元したセッションが無効なら破棄してフォールバックする。
	if restoredFromStore {
		if delErr := sessions.Delete(); delErr != nil {
			logger.Info("failed to delete invalid session from store:", delErr)
		}
		_ = os.Unsetenv("BW_SESSION")
	}

	logger.Info("Vault is locked. Please enter your master password to unlock.")

	password, promptErr := promptPassword()
	if promptErr != nil {
		return fmt.Errorf("failed to get master password: %w", promptErr)
	}

	unlockErr := bw.Unlock(password)
	if unlockErr == nil {
		logger.Info("Vault unlocked successfully")
		persistSession(sessions, logger)
		return fn()
	}

	// Unlock 失敗後に Login → Unlock（解決済み host の email / URL を使用）
	email, serverURL := hostLoginFields(cfg)
	if email != "" {
		loginErr := bw.Login(email, password, serverURL)
		if loginErr != nil {
			// api Login は API Key 認証なので、ここでの失敗は認証切れとして返す
			if IsNotAuthenticatedError(loginErr) {
				return loginErr
			}
			return fmt.Errorf("failed to login Bitwarden CLI: %w", loginErr)
		}
		logger.Info("Bitwarden CLI logged in successfully")

		// bw login --raw（および already-logged-in 時の unlock）は BW_SESSION を設定する。
		// 続けて Unlock すると既存セッションが無効化され、失敗時にロック状態へ戻る。
		if strings.TrimSpace(os.Getenv("BW_SESSION")) != "" {
			logger.Info("Vault unlocked successfully")
			persistSession(sessions, logger)
			return fn()
		}

		unlockErr = bw.Unlock(password)
		if unlockErr != nil {
			return fmt.Errorf("failed to unlock after login: %w", unlockErr)
		}
		logger.Info("Vault unlocked successfully")
		persistSession(sessions, logger)
		return fn()
	}

	return fmt.Errorf("failed to unlock vault: %w", unlockErr)
}

// hostLoginFields returns email and self-host server URL from the default host.
func hostLoginFields(cfg *config.Config) (email, serverURL string) {
	if cfg == nil {
		return "", ""
	}
	h := cfg.DefaultHost()
	if h == nil {
		return "", ""
	}
	email = strings.TrimSpace(h.Email)
	if h.Type == config.HostTypeSelfhost {
		serverURL = strings.TrimSpace(h.HostURL)
	}
	return email, serverURL
}

// persistSession はプロセス内の BW_SESSION を SessionStore へ保存します（失敗は非致命）。
func persistSession(sessions SessionStore, logger Logger) {
	session := strings.TrimSpace(os.Getenv("BW_SESSION"))
	if session == "" {
		return
	}
	if err := sessions.Set(session); err != nil {
		logger.Info("failed to persist session to store:", err)
	}
}

// PushEnvCore は管理対象ファイル（.env* / *.tfvars / *.tfvars.json）を Bitwarden にプッシュします。
// .example を含む名前のファイルは除外します。
// filter は基盤ルール通過後に適用する（#133 save_files / not_save_files）。
func PushEnvCore(
	fromDir, projectName string,
	filter ManagedFileFilter,
	fs FileSystem,
	bw BwClient,
	cfg *config.Config,
	promptPassword func() (string, error),
	logger Logger,
	sessions SessionStore,
) error {
	envFiles, err := findEnvFilesFromFS(fs, fromDir, filter)
	if err != nil {
		return fmt.Errorf("failed to find managed files: %w", err)
	}

	if len(envFiles) == 0 {
		return fmt.Errorf("no managed files found in %s", fromDir)
	}

	// 各ファイルを読み込んで MultiEnvData に格納
	multiData := make(MultiEnvData)
	for _, envPath := range envFiles {
		content, err := fs.ReadFile(envPath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", envPath, err)
		}
		fileName := filepath.Base(envPath)
		envData := parseEnvContent(content)
		multiData[fileName] = *envData
	}

	// JSON に変換
	jsonData, err := multiEnvDataToJSON(multiData)
	if err != nil {
		return fmt.Errorf("failed to convert to JSON: %w", err)
	}

	// dotenvs フォルダ ID を取得
	var folderID string
	err = WithUnlockRetry(bw, cfg, promptPassword, logger, sessions, func() error {
		var innerErr error
		folderID, innerErr = bw.GetDotenvsFolderID()
		return innerErr
	})
	if err != nil {
		return fmt.Errorf("failed to get dotenvs folder: %w", err)
	}

	// 既存アイテムを検索
	var existingItem *FullItem
	err = WithUnlockRetry(bw, cfg, promptPassword, logger, sessions, func() error {
		var innerErr error
		existingItem, innerErr = bw.GetItemByName(folderID, projectName)
		return innerErr
	})
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	// 既存アイテムがあれば更新、なければ新規作成
	if existingItem != nil {
		err = WithUnlockRetry(bw, cfg, promptPassword, logger, sessions, func() error {
			return bw.UpdateNoteItem(existingItem.ID, jsonData)
		})
		if err != nil {
			return fmt.Errorf("failed to update item: %w", err)
		}
	} else {
		err = WithUnlockRetry(bw, cfg, promptPassword, logger, sessions, func() error {
			return bw.CreateNoteItem(folderID, projectName, jsonData)
		})
		if err != nil {
			return fmt.Errorf("failed to create item: %w", err)
		}
	}

	return nil
}

// GetPushedEnvFiles は push 対象の管理対象ファイル名一覧を返します（表示用）
func GetPushedEnvFiles(
	fromDir string,
	filter ManagedFileFilter,
	fs FileSystem,
) ([]string, error) {
	envFiles, err := findEnvFilesFromFS(fs, fromDir, filter)
	if err != nil {
		return nil, err
	}

	// ファイル名のみを返す
	var names []string
	for _, path := range envFiles {
		names = append(names, filepath.Base(path))
	}
	return names, nil
}

// findEnvFilesFromFS は FileSystem 経由で管理対象ファイルを検出します。
// 対象: .env* / *.tfvars / *.tfvars.json（名前に .example を含むものは除外）
// filter は isManagedFileName 通過後に basename glob で絞り込みます（#133）。
func findEnvFilesFromFS(fs FileSystem, dirPath string, filter ManagedFileFilter) ([]string, error) {
	entries, err := fs.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var envFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !isManagedFileName(name) {
			continue
		}
		if !filter.Allow(name) {
			continue
		}

		envFiles = append(envFiles, filepath.Join(dirPath, name))
	}

	// Sort by filename for consistent ordering
	sortEnvFiles(envFiles)

	return envFiles, nil
}

// isManagedFileName reports whether a basename is a bwsf-managed file.
func isManagedFileName(filename string) bool {
	if isExampleFile(filename) {
		return false
	}
	if strings.HasPrefix(filename, ".env") {
		return true
	}
	if strings.HasSuffix(filename, ".tfvars.json") || strings.HasSuffix(filename, ".tfvars") {
		return true
	}
	return false
}

// isExampleFile checks if a filename contains ".example" anywhere in it
func isExampleFile(filename string) bool {
	return strings.Contains(filename, ".example")
}

// sortEnvFiles sorts env files with .env first, then alphabetically
func sortEnvFiles(files []string) {
	for i := 0; i < len(files); i++ {
		for j := i + 1; j < len(files); j++ {
			nameI := filepath.Base(files[i])
			nameJ := filepath.Base(files[j])

			// .env should always come first
			if nameI == ".env" {
				continue
			}
			if nameJ == ".env" {
				files[i], files[j] = files[j], files[i]
				continue
			}

			// Otherwise, sort alphabetically
			if nameI > nameJ {
				files[i], files[j] = files[j], files[i]
			}
		}
	}
}

// PullEnvCore は Bitwarden から .env ファイルをプルするコアロジックです。
// 複数の .env* ファイルを復元します。
func PullEnvCore(
	outputDir, projectName string,
	fs FileSystem,
	bw BwClient,
	cfg *config.Config,
	promptPassword func() (string, error),
	confirmOverwrite func(path string) (bool, error),
	logger Logger,
	sessions SessionStore,
) error {
	// dotenvs フォルダ ID を取得
	var folderID string
	err := WithUnlockRetry(bw, cfg, promptPassword, logger, sessions, func() error {
		var innerErr error
		folderID, innerErr = bw.GetDotenvsFolderID()
		return innerErr
	})
	if err != nil {
		return fmt.Errorf("failed to get dotenvs folder: %w", err)
	}

	// アイテムを取得
	var item *FullItem
	err = WithUnlockRetry(bw, cfg, promptPassword, logger, sessions, func() error {
		var innerErr error
		item, innerErr = bw.GetItemByName(folderID, projectName)
		return innerErr
	})
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	// アイテムが見つからない場合
	if item == nil {
		return fmt.Errorf("item '%s' not found in dotenvs folder", projectName)
	}

	// JSON から MultiEnvData を復元
	multiData, err := restoreMultiEnvFromJSON(item.Notes)
	if err != nil {
		// 旧形式の場合は単一ファイルとして処理（下位互換性のため）
		envContent, legacyErr := restoreEnvFileFromJSON(item.Notes)
		if legacyErr != nil {
			return fmt.Errorf("failed to restore .env from JSON: %w", err)
		}
		// 旧形式を新形式に変換
		multiData = MultiEnvData{
			".env": EnvData{Lines: strings.Split(envContent, "\n")},
		}
	}

	// ディレクトリを作成（必要に応じて）
	// "." や ".." 以外の場合のみディレクトリ作成を試みる
	if outputDir != "." && outputDir != ".." {
		if err := fs.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// 各ファイルを書き出し
	for fileName, envData := range multiData {
		envPath := filepath.Join(outputDir, fileName)

		// ファイルの存在確認
		info, err := fs.Stat(envPath)
		if err == nil && !info.IsNotExist() {
			// ファイルが存在する場合、上書き確認
			confirmed, confirmErr := confirmOverwrite(envPath)
			if confirmErr != nil {
				return fmt.Errorf("failed to confirm overwrite: %w", confirmErr)
			}
			if !confirmed {
				continue // このファイルはスキップ
			}
		}

		// ファイル内容を復元
		envContent := restoreEnvContentFromData(envData)

		// ファイルを書き出し
		if err := fs.WriteFile(envPath, []byte(envContent), 0644); err != nil {
			return fmt.Errorf("failed to write %s file: %w", fileName, err)
		}
	}

	return nil
}

// CleanEnvCore は管理対象のローカルファイルを、リモートバックアップを確認したうえで削除します。
func CleanEnvCore(
	targetDir, projectName string,
	filter ManagedFileFilter,
	fs FileSystem,
	bw BwClient,
	cfg *config.Config,
	promptPassword func() (string, error),
	selectMismatchAction func(mismatchedFiles []string) (CleanMismatchAction, error),
	logger Logger,
	sessions SessionStore,
) error {
	envFiles, err := findEnvFilesFromFS(fs, targetDir, filter)
	if err != nil {
		return fmt.Errorf("failed to find managed files: %w", err)
	}
	if len(envFiles) == 0 {
		return fmt.Errorf("no managed files found in %s", targetDir)
	}

	localData := make(MultiEnvData)
	for _, envPath := range envFiles {
		content, readErr := fs.ReadFile(envPath)
		if readErr != nil {
			return fmt.Errorf("failed to read %s: %w", envPath, readErr)
		}
		localData[filepath.Base(envPath)] = *parseEnvContent(content)
	}

	var folderID string
	err = WithUnlockRetry(bw, cfg, promptPassword, logger, sessions, func() error {
		var innerErr error
		folderID, innerErr = bw.GetDotenvsFolderID()
		return innerErr
	})
	if err != nil {
		return fmt.Errorf("failed to get dotenvs folder: %w", err)
	}

	var item *FullItem
	err = WithUnlockRetry(bw, cfg, promptPassword, logger, sessions, func() error {
		var innerErr error
		item, innerErr = bw.GetItemByName(folderID, projectName)
		return innerErr
	})
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}
	if item == nil {
		return fmt.Errorf("item '%s' not found in dotenvs folder; aborting clean", projectName)
	}

	remoteData, err := restoreMultiEnvFromJSON(item.Notes)
	if err != nil {
		envContent, legacyErr := restoreEnvFileFromJSON(item.Notes)
		if legacyErr != nil {
			return fmt.Errorf("failed to restore remote env data: %w", err)
		}
		remoteData = MultiEnvData{
			".env": EnvData{Lines: strings.Split(envContent, "\n")},
		}
	}
	if len(remoteData) == 0 {
		return fmt.Errorf("item '%s' has no env files on remote; aborting clean", projectName)
	}

	mismatched := diffMultiEnvData(localData, remoteData)
	if len(mismatched) == 0 {
		return removeLocalEnvFiles(fs, envFiles)
	}

	action, err := selectMismatchAction(mismatched)
	if err != nil {
		return fmt.Errorf("failed to select clean action: %w", err)
	}

	switch action {
	case CleanActionAbort:
		return ErrCleanAborted
	case CleanActionOverwriteRemoteThenClean:
		if err := PushEnvCore(targetDir, projectName, filter, fs, bw, cfg, promptPassword, logger, sessions); err != nil {
			return fmt.Errorf("failed to overwrite remote before clean: %w", err)
		}
		return removeLocalEnvFiles(fs, envFiles)
	case CleanActionRemoveLocal:
		return removeLocalEnvFiles(fs, envFiles)
	default:
		return fmt.Errorf("unknown clean action: %v", action)
	}
}

func removeLocalEnvFiles(fs FileSystem, envFiles []string) error {
	for _, envPath := range envFiles {
		if err := fs.Remove(envPath); err != nil {
			return fmt.Errorf("failed to remove %s: %w", envPath, err)
		}
	}
	return nil
}

func diffMultiEnvData(local, remote MultiEnvData) []string {
	keys := make(map[string]struct{})
	for k := range local {
		keys[k] = struct{}{}
	}
	for k := range remote {
		keys[k] = struct{}{}
	}

	var mismatched []string
	for k := range keys {
		l, lok := local[k]
		r, rok := remote[k]
		if !lok || !rok || !reflect.DeepEqual(l.Lines, r.Lines) {
			mismatched = append(mismatched, k)
		}
	}
	sortFileNames(mismatched)
	return mismatched
}

// GetPulledEnvFiles は pull 対象の .env ファイル名一覧を返します（表示用）
func GetPulledEnvFiles(
	projectName string,
	bw BwClient,
	cfg *config.Config,
	promptPassword func() (string, error),
	logger Logger,
	sessions SessionStore,
) ([]string, error) {
	// dotenvs フォルダ ID を取得
	var folderID string
	err := WithUnlockRetry(bw, cfg, promptPassword, logger, sessions, func() error {
		var innerErr error
		folderID, innerErr = bw.GetDotenvsFolderID()
		return innerErr
	})
	if err != nil {
		return nil, err
	}

	// アイテムを取得
	var item *FullItem
	err = WithUnlockRetry(bw, cfg, promptPassword, logger, sessions, func() error {
		var innerErr error
		item, innerErr = bw.GetItemByName(folderID, projectName)
		return innerErr
	})
	if err != nil {
		return nil, err
	}

	if item == nil {
		return nil, nil
	}

	// JSON から MultiEnvData を復元
	multiData, err := restoreMultiEnvFromJSON(item.Notes)
	if err != nil {
		// 旧形式の場合
		return []string{".env"}, nil
	}

	var names []string
	for fileName := range multiData {
		names = append(names, fileName)
	}

	// ソート
	sortFileNames(names)
	return names, nil
}

// sortFileNames sorts file names with .env first, then alphabetically
func sortFileNames(names []string) {
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			// .env should always come first
			if names[i] == ".env" {
				continue
			}
			if names[j] == ".env" {
				names[i], names[j] = names[j], names[i]
				continue
			}

			// Otherwise, sort alphabetically
			if names[i] > names[j] {
				names[i], names[j] = names[j], names[i]
			}
		}
	}
}

// ListDotenvsCore は dotenvs フォルダ内のアイテム一覧を取得するコアロジックです。
func ListDotenvsCore(
	bw BwClient,
	cfg *config.Config,
	promptPassword func() (string, error),
	logger Logger,
	sessions SessionStore,
) ([]Item, error) {
	// dotenvs フォルダ ID を取得
	var folderID string
	err := WithUnlockRetry(bw, cfg, promptPassword, logger, sessions, func() error {
		var innerErr error
		folderID, innerErr = bw.GetDotenvsFolderID()
		return innerErr
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get dotenvs folder: %w", err)
	}

	// アイテム一覧を取得
	var items []Item
	err = WithUnlockRetry(bw, cfg, promptPassword, logger, sessions, func() error {
		var innerErr error
		items, innerErr = bw.ListItemsInFolder(folderID)
		return innerErr
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}

	return items, nil
}

// collectHostEmailConfig prompts for host type, optional self-hosted URL, and email.
// Host type values are "cloud" or "selfhosted" (CLI/prompt shorthand).
func collectHostEmailConfig(
	selectHostType func() (string, error),
	inputURL func() (string, error),
	inputEmail func() (string, error),
) (hostType, selfhostedURL, email string, err error) {
	hostType, err = selectHostType()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to select host type: %w", err)
	}

	normalized := normalizePromptHostType(hostType)
	if normalized == config.HostTypeSelfhost {
		selfhostedURL, err = inputURL()
		if err != nil {
			return "", "", "", fmt.Errorf("failed to get URL: %w", err)
		}
		if strings.TrimSpace(selfhostedURL) == "" {
			return "", "", "", fmt.Errorf("self-hosted URL cannot be empty")
		}
	}

	email, err = inputEmail()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get email: %w", err)
	}
	if strings.TrimSpace(email) == "" {
		return "", "", "", fmt.Errorf("email cannot be empty")
	}

	return hostType, selfhostedURL, email, nil
}

func normalizePromptHostType(hostType string) string {
	switch strings.TrimSpace(strings.ToLower(hostType)) {
	case "selfhosted", "self-hosted", config.HostTypeSelfhost:
		return config.HostTypeSelfhost
	default:
		return config.HostTypeCloud
	}
}

// MapPromptHostToV2 converts cloud/selfhosted prompt values into a v2 Host.
func MapPromptHostToV2(hostType, url, email, targetSection string) (config.Host, error) {
	hType := normalizePromptHostType(hostType)
	hURL := strings.TrimSpace(url)
	switch hType {
	case config.HostTypeSelfhost:
		if hURL == "" {
			return config.Host{}, fmt.Errorf("self-hosted URL cannot be empty")
		}
	default:
		if hURL == "" {
			hURL = config.DefaultCloudURL
		}
	}
	section := strings.TrimSpace(targetSection)
	if section == "" {
		section = config.DefaultFolderName
	}
	return config.Host{
		ID:            config.DefaultHostID,
		Type:          hType,
		HostURL:       hURL,
		Email:         strings.TrimSpace(email),
		TargetSection: section,
		IsDefault:     true,
	}, nil
}

// preserveSetupFields copies device_identifier onto matching host ids.
func preserveSetupFields(newConfig, existing *config.Config) {
	if existing == nil || newConfig == nil {
		return
	}
	for i := range newConfig.Settings.Hosts {
		nh := &newConfig.Settings.Hosts[i]
		if nh.DeviceIdentifier != "" {
			continue
		}
		if eh := existing.FindHost(nh.ID); eh != nil {
			nh.DeviceIdentifier = eh.DeviceIdentifier
		}
	}
}

// UpsertDefaultHost sets or replaces the default host entry in cfg.
func UpsertDefaultHost(cfg *config.Config, host config.Host) {
	if cfg == nil {
		return
	}
	host.IsDefault = true
	if host.ID == "" {
		host.ID = config.DefaultHostID
	}
	for i := range cfg.Settings.Hosts {
		cfg.Settings.Hosts[i].IsDefault = false
	}
	for i := range cfg.Settings.Hosts {
		if cfg.Settings.Hosts[i].ID == host.ID {
			cfg.Settings.Hosts[i] = host
			return
		}
	}
	cfg.Settings.Hosts = append(cfg.Settings.Hosts, host)
}

// SetupAPIConfigCore configures a default host (type/url/email) without Login.
// Authentication is performed separately via `bwsf auth`.
// Folder creation is handled by EnsureConfiguredFolderCore after auth/unlock is available.
func SetupAPIConfigCore(
	logger Logger,
	selectHostType func() (string, error),
	inputURL func() (string, error),
	inputEmail func() (string, error),
) error {
	return SetupAPIConfigCoreWithFolder(logger, selectHostType, inputURL, inputEmail, "")
}

// SetupAPIConfigCoreWithFolder is SetupAPIConfigCore with an optional target_section.
func SetupAPIConfigCoreWithFolder(
	logger Logger,
	selectHostType func() (string, error),
	inputURL func() (string, error),
	inputEmail func() (string, error),
	targetSection string,
) error {
	existingConfig, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load existing config: %w", err)
	}
	if existingConfig != nil {
		logger.Info("Existing configuration found. Host settings will be updated.")
	}

	hostType, selfhostedURL, email, err := collectHostEmailConfig(selectHostType, inputURL, inputEmail)
	if err != nil {
		return err
	}

	section := strings.TrimSpace(targetSection)
	if section == "" && existingConfig != nil {
		if dh := existingConfig.DefaultHost(); dh != nil {
			section = dh.TargetSection
		}
	}

	host, err := MapPromptHostToV2(hostType, selfhostedURL, email, section)
	if err != nil {
		return err
	}
	if existingConfig != nil {
		if dh := existingConfig.DefaultHost(); dh != nil && host.ID == config.DefaultHostID {
			host.ID = dh.ID
		}
	}

	cfg := existingConfig
	if cfg == nil {
		cfg = config.NewEmptyConfig()
	}
	var priorHosts []config.Host
	if existingConfig != nil {
		priorHosts = append([]config.Host(nil), existingConfig.Settings.Hosts...)
	}
	UpsertDefaultHost(cfg, host)
	preserveSetupFields(cfg, &config.Config{Settings: config.GlobalSettings{Hosts: priorHosts}})

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	logger.Info("Configuration saved. Run `bwsf auth` to authenticate with a Personal API Key.")
	return nil
}

// EnsureConfiguredFolderCore checks the configured Bitwarden folder and optionally creates it.
// Uses the resolved / default host target_section via ResolveFolderName.
func EnsureConfiguredFolderCore(
	bw BwClient,
	cfg *config.Config,
	logger Logger,
	promptPassword func() (string, error),
	confirmCreateFolder func() (bool, error),
) error {
	if bw == nil {
		return fmt.Errorf("bitwarden client is required")
	}
	resolvedFolder := config.ResolveFolderName(cfg)
	if h := cfg.DefaultHost(); h != nil && strings.TrimSpace(h.TargetSection) != "" {
		resolvedFolder = h.TargetSection
	}

	var exists bool
	err := WithUnlockRetry(bw, cfg, promptPassword, logger, nil, func() error {
		var innerErr error
		exists, innerErr = bw.DotenvsFolderExists()
		return innerErr
	})
	if err != nil {
		return fmt.Errorf("failed to check %s folder: %w", resolvedFolder, err)
	}
	if exists {
		return nil
	}

	confirmed, confirmErr := confirmCreateFolder()
	if confirmErr != nil {
		return fmt.Errorf("failed to confirm folder creation: %w", confirmErr)
	}
	if !confirmed {
		logger.Info(resolvedFolder, " folder was not created")
		return nil
	}

	err = WithUnlockRetry(bw, cfg, promptPassword, logger, nil, func() error {
		return bw.CreateDotenvsFolder()
	})
	if err != nil {
		return fmt.Errorf("failed to create %s folder: %w", resolvedFolder, err)
	}
	logger.Info(resolvedFolder, " folder created successfully")
	return nil
}

// SetupBitwardenCore is removed in v0.20 (API-only). Kept as a stub that errors.
func SetupBitwardenCore(
	fs FileSystem,
	bw BwClient,
	logger Logger,
	selectHostType func() (string, error),
	inputURL func() (string, error),
	inputEmail func() (string, error),
	inputPassword func() (string, error),
	confirmCreateFolder func() (bool, error),
) error {
	_, _, _, _, _, _, _ = fs, bw, logger, selectHostType, inputURL, inputEmail, inputPassword
	_ = confirmCreateFolder
	return fmt.Errorf("bw CLI setup path removed; v0.20+ uses API only — run `bwsf setup` then `bwsf auth`")
}

// parseEnvContent は .env ファイルの内容をパースします。
func parseEnvContent(content []byte) *EnvData {
	lines := strings.Split(string(content), "\n")
	// 末尾の空行を削除
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return &EnvData{Lines: lines}
}

// envDataToJSON は EnvData を JSON 文字列に変換します。
func envDataToJSON(data *EnvData) (string, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal env data to JSON: %w", err)
	}
	return string(jsonBytes), nil
}

// restoreEnvFileFromJSON は JSON から .env ファイルの内容を復元します。
func restoreEnvFileFromJSON(jsonStr string) (string, error) {
	var data EnvData
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return strings.Join(data.Lines, "\n"), nil
}

// multiEnvDataToJSON は MultiEnvData を JSON 文字列に変換します。
func multiEnvDataToJSON(data MultiEnvData) (string, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal multi env data to JSON: %w", err)
	}
	return string(jsonBytes), nil
}

// restoreMultiEnvFromJSON は JSON から MultiEnvData を復元します。
func restoreMultiEnvFromJSON(jsonStr string) (MultiEnvData, error) {
	var data MultiEnvData
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return data, nil
}

// restoreEnvContentFromData は EnvData から .env ファイルの内容を復元します。
func restoreEnvContentFromData(data EnvData) string {
	return strings.Join(data.Lines, "\n")
}
