# config / cmd: host 解決と `--host`（Issue #177 / v0.20.0 §1.1）

対象パッケージ:

- `app/src/config`（`ResolveHost` 相当の純関数）
- `app/src/cmd`（`push` / `pull` / `list` / `clean` の `--host` 配線）

Issue: [#177](https://github.com/b4moss/bwsf/issues/177)  
製品正本: [`../specs/v0.20.0-multi-host.md`](../specs/v0.20.0-multi-host.md) §1.1

関連: [`global_v2.md`](./global_v2.md)、[`save_files_bang.md`](./save_files_bang.md)（プロジェクト `host` キー）

---

### 前提（本仕様で固定）

| # | 方針 |
|---|------|
| H1 | 優先度（高→低）: **CLI `--host <id>`** → **プロジェクト設定の `host`** → **グローバル `hosts[]` の `is_default: true`** |
| H2 | 解決できた id の host オブジェクトを返す（接続情報はグローバル `hosts[]` 側） |
| H3 | 指定 id が `hosts[]` に無い → エラー |
| H4 | CLI / プロジェクトとも無く、`hosts` が空 → エラー（setup で host 追加、または CLI／プロジェクトで指定を促す） |
| H5 | `hosts` 非空なのに default 無し／複数は **ロード時スキーマエラー**（本解決関数の前段。[`global_v2.md`](./global_v2.md)） |
| H6 | プロジェクト `host` は **id 参照のみ**。グローバルに同 id が必要 |
| H7 | #177 で `--host` を付けるコマンド: **`push` / `pull` / `list` / `clean`** |
| H8 | `unlock` / `lock` の `--host` は [#153](https://github.com/b4moss/bwsf/issues/153) / [`../cmd/unlock_lock.md`](../cmd/unlock_lock.md)。`auth login` / `auth logout` は [#174](https://github.com/b4moss/bwsf/issues/174) / [`../cmd/auth_login_logout.md`](../cmd/auth_login_logout.md)。解決関数は共有。**本仕様（#177）のコマンド登録テストは H7 のみ** |

---

## 1. `ResolveHost`（単体）

引数例: `(cfg *GlobalConfig, cliHost string, projectHost string) (*Host, error)`

フィクスチャ: hosts に `default`（`is_default: true`）と `work`（false）。

#### テスト：正常系

| # | cliHost | projectHost | 期待 |
|---|---------|-------------|------|
| R1 | `"work"` | `"default"` | `work`（CLI 優先） |
| R2 | `""` | `"work"` | `work` |
| R3 | `""` | `""` | `default`（is_default） |
| R4 | `"default"` | `""` | `default` |

#### テスト: 異常系

| # | 条件 | 期待 |
|---|------|------|
| R5 | cliHost=`"missing"` | エラー（hosts に無い） |
| R6 | projectHost=`"missing"`（cli 空） | エラー |
| R7 | `hosts: []` かつ cli/project 空 | エラー |
| R8 | `hosts: []` だが cliHost=`"default"` | エラー（存在しない） |

---

## 2. コマンド配線（`--host`）

#### テスト：正常系

| # | 期待 |
|---|------|
| C1 | `push` / `pull` / `list` / `clean` に `--host` フラグが登録されている |
| C2 | `--host work` 指定時、クライアント生成／フォルダ名解決が `work` の `target_section` / `host_url` / `type` を使う（スタブまたは HOME 下設定で検証） |
| C3 | `--host` 無し・プロジェクト `host` 無し → default host を使う |

#### テスト: 異常系

| # | 条件 | 期待 |
|---|------|------|
| C4 | `--host` に未知 id | 非ゼロ終了。Bitwarden 呼び出し前に失敗してよい |
| C5 | hosts 空で vault 系コマンド（host 解決が必要） | エラー（setup / 指定を促す） |

---

## 3. 対象外

- Keychain の host 単位キー → [`../infra/secretstore_hosts.md`](../infra/secretstore_hosts.md)（#153）
- unlock / lock の CLI 登録 → [`../cmd/unlock_lock.md`](../cmd/unlock_lock.md)（#153）
- auth login / logout → [`../cmd/auth_login_logout.md`](../cmd/auth_login_logout.md)（#174）
- `bwsf init` が書くプロジェクト `host`（#193）。本仕様は「読んだときの解決」のみ

---

## 実装メモ（仕様外・参考）

```text
id = firstNonEmpty(cliHost, project.host, "")
if id != "" → lookup hosts by id
else → hosts から is_default
```
