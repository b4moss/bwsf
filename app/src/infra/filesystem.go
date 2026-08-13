package infra

import (
	"os"

	"bwsf/src/core"
)

// RealFileSystem は core.FileSystem インターフェースの実装です。
type RealFileSystem struct{}

// NewFileSystem は RealFileSystem のインスタンスを作成します。
func NewFileSystem() *RealFileSystem {
	return &RealFileSystem{}
}

// OpenEnvFile は .env ファイルを読み込みます。
func (fs *RealFileSystem) OpenEnvFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// ReadFile はファイルを読み込みます。
func (fs *RealFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFile はファイルを書き込みます。
func (fs *RealFileSystem) WriteFile(path string, data []byte, perm uint32) error {
	return os.WriteFile(path, data, os.FileMode(perm))
}

// Remove はファイルを削除します。
func (fs *RealFileSystem) Remove(path string) error {
	return os.Remove(path)
}

// Stat はファイル情報を取得します。
func (fs *RealFileSystem) Stat(path string) (core.FileInfo, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &realFileInfo{notExist: true}, nil
		}
		return nil, err
	}
	return &realFileInfo{notExist: false}, nil
}

// MkdirAll はディレクトリを再帰的に作成します。
func (fs *RealFileSystem) MkdirAll(path string, perm uint32) error {
	return os.MkdirAll(path, os.FileMode(perm))
}

// ReadDir はディレクトリ内のエントリを読み込みます。
func (fs *RealFileSystem) ReadDir(path string) ([]core.DirEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	result := make([]core.DirEntry, len(entries))
	for i, entry := range entries {
		result[i] = &realDirEntry{entry: entry}
	}
	return result, nil
}

// realFileInfo は core.FileInfo インターフェースの実装です。
type realFileInfo struct {
	notExist bool
}

// IsNotExist はファイルが存在しないかどうかを返します。
func (fi *realFileInfo) IsNotExist() bool {
	return fi.notExist
}

// realDirEntry は core.DirEntry インターフェースの実装です。
type realDirEntry struct {
	entry os.DirEntry
}

// Name はエントリ名を返します。
func (de *realDirEntry) Name() string {
	return de.entry.Name()
}

// IsDir はディレクトリかどうかを返します。
func (de *realDirEntry) IsDir() bool {
	return de.entry.IsDir()
}
