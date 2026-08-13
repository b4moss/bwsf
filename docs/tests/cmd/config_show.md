# cmd: `bwsf config show`（Issue #125 / v0.15.0）

対象パッケージ: `app/src/cmd`（必要なら薄い表示整形を `app/src/config` に置く）  
Issue: [#125](https://github.com/b4moss/bwsf/issues/125) — 現在の config 値を確認するコマンド。

関連: 既存 `LoadConfig` / `Config` / `ResolveFolderName`（`docs/TEST.md` §6、`app/src/config`）。

---

### 前提（本仕様で固定）

| # | 方針 |
|---|------|
| C1 | コマンド形は `bwsf config show`（親 `config` + 子 `show`）。`bwsf config` 単体はヘルプ表示でよい |
| C2 | 表示対象は `Config` の永続フィールドのみ（`host_type` / `selfhosted_url` / `email` / `folder_name`）。パスワード等は持たない・出さない |
| C3 | `folder_name` は **実効値** を出す（未設定・空白は `ResolveFolderName` により `dotenvs`）。併せて「設定ファイル上の値」が空だったことが分かる表示でもよいが、テストで固定するのは実効値の行 |
| C4 | 設定ファイルが無い（`LoadConfig` が `nil, nil`）ときは **エラー終了**（セットアップを促すメッセージ）。空の Config を成功表示しない |
| C5 | 読み取り専用。ファイルを変更しない |
| C6 | 出力は人間向けの固定ラベル行（テストで部分一致可能）。デフォルトは JSON ダンプにしない（将来 `--json` を足す余地は残すが本 Issue 外） |

---

### `bwsf config`（親コマンド）

- cobra に `config` を登録する
- 引数なし実行時はサブコマンド一覧／Usage を出して終了する（実装は cobra 既定で可）

#### テスト：正常系

- root の子コマンドに `config` が登録されている
- `config` の子に `show` が登録されている

#### テスト: 異常系

- （任意）未知のサブコマンドは cobra 既定のエラーになること（退行レベルでよい）

---

### `bwsf config show`

- `LoadConfig` 相当で `~/.config/bwsf/config.json` を読む（テストは HOME 差し替えまたは DI）
- 成功時は少なくとも次のラベル付き値を stdout（または既存 `utils` 情報出力）へ出す:
  - Host type: `{host_type}`
  - Self-hosted URL: `{selfhosted_url}`（cloud で空なら空文字のままでよい）
  - Email: `{email}`
  - Folder name: `{ResolveFolderName(cfg)}`
- `bw` CLI の有無は **問わない**（ローカル設定の表示のみ）
- Bitwarden へのネットワークアクセスはしない

#### テスト：正常系

- cloud 設定（`host_type=cloud`, email あり, URL 空, folder 未設定）で、上記ラベルが出て Folder name が `dotenvs`
- selfhosted 設定（URL・email・folder_name 明示）で、各値がそのまま（folder は設定値）出る
- 既存 `setup` / `push` 等が読むのと同じパスの内容を表示する（HOME 下 `.config/bwsf/config.json`）

#### テスト: 異常系

- 設定ファイル無し → 非ゼロ終了相当のエラー（メッセージに config / setup を促す旨が含まれる）
- 壊れた JSON → `LoadConfig` と同様にパースエラーを伝播し、非ゼロ終了
- （退行）成功パスで `SaveConfig` や設定ファイルへの書き込みが呼ばれない

---

### 表示整形（任意ヘルパ）

実装が `FormatConfigShow(cfg *Config) string` のような純関数に寄せる場合、ユニットはここを厚くしてよい。

#### テスト：正常系

- `folder_name` 空 → 出力に実効フォルダ `dotenvs` が含まれる
- `folder_name` が `"team-envs"` → 出力に `team-envs` が含まれる

#### テスト: 異常系

- `cfg == nil` はヘルパに渡さない（呼び出し側で C4 を担保）。渡した場合のパニック防止は任意
