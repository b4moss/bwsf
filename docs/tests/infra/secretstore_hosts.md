# infra: host 単位 SecretStore / Keychain（Issue #153 / v0.20.0 §3）

対象パッケージ: `app/src/infra`（`SecretStore` および API 資格情報ヘルパ）  
Issue: [#153](https://github.com/b4moss/bwsf/issues/153)  
製品正本: [`../specs/v0.20.0-multi-host.md`](../specs/v0.20.0-multi-host.md) §3（Keychain 移行は §2.6 最終行）

関連:

- Unlock / ClearSession（メモリ）→ [`apiclient_unlock.md`](./apiclient_unlock.md)（**本仕様が Keychain 側を優先**）
- `vault_unlock` の restore フロー → [`../core/vault_unlock_restore.md`](../core/vault_unlock_restore.md)
- CLI unlock / lock → [`../cmd/unlock_lock.md`](../cmd/unlock_lock.md)
- 設定ファイル flat→v2（Keychain は対象外と明記）→ [`../config/migrate_v2.md`](../config/migrate_v2.md) M8

---

### 前提（本仕様で固定）

| # | 方針 |
|---|------|
| K1 | Keychain（OS secret store）の account/key は **`hosts/<id>/api_client_id`**, **`hosts/<id>/api_client_secret`**, **`hosts/<id>/vault_unlock`** |
| K2 | Personal API Key は **host 単位**。複数 host を同時に保持できる |
| K3 | `vault_unlock` は unlock 状態復元用の **不透明データ**。マスターパスワードは保存しない |
| K4 | `vault_unlock` の有効期間は **`lock` / `auth logout` まで**（プロセス終了・再起動後も Keychain に残る）。メモリ上の鍵はコマンド終了時に破棄してよい（[`../cmd/session_lifecycle.md`](../cmd/session_lifecycle.md)） |
| K5 | 旧 flat キー（`api_client_id` / `api_client_secret`）は **読み替え**し、書き込み成功時に **`hosts/default/...` へ移行**して旧キーを削除する（`default` は設定マイグレーションの初期 id と一致） |
| K6 | 旧 flat に `vault_unlock` 相当は無い（新規キーのみ） |
| K7 | `Save` / `Load` / `Clear` の API 資格情報ヘルパは **host id 必須** |
| K8 | host id に `/` は設定スキーマで禁止済み。ヘルパは空 id を拒否する |
| K9 | service 名は従来どおり `bwsf`（変更しない） |
| K10 | ログ・エラーメッセージに Client Secret / `vault_unlock` 生値 / MP を出さない |

---

## 1. キー命名と CRUD

フィクスチャ: `MemorySecretStore`（または同等の差し替え可能な `SecretStore`）。

#### テスト：正常系

| # | 操作 | 期待 |
|---|------|------|
| C1 | host `work` に API Key を Save | `hosts/work/api_client_id` と `hosts/work/api_client_secret` のみがセットされる。flat キーは無い |
| C2 | host `work` を Load | 保存した ClientID / ClientSecret が返る |
| C3 | host `a` と `b` に別キーを Save | 互いに独立。`a` の Clear で `b` が消えない |
| C4 | host `work` の API Key を Clear | 当該 host の id/secret が欠落。`vault_unlock` は消さない（Clear API 資格情報の責務外） |
| C5 | `vault_unlock` を Set / Get / Delete | `hosts/<id>/vault_unlock` のみが対象。API Key キーは不変 |

#### テスト: 異常系

| # | 条件 | 期待 |
|---|------|------|
| C6 | host id 空で Save / Load / Clear | エラー。ストアを変更しない |
| C7 | 未保存 host を Load | `ErrSecretNotFound`（または同等） |
| C8 | ClientID または ClientSecret が空の Save | エラー。部分書き込みしない |

---

## 2. 旧 flat キーの読み替え／移行

初期状態: flat `api_client_id` / `api_client_secret` のみ（host プレフィックス無し）。

#### テスト：正常系

| # | 操作 | 期待 |
|---|------|------|
| M1 | host `default` で Load（新キー無し・flat あり） | flat の値が返る（読み替え成功） |
| M2 | host `default` で Load 後、または Save 成功後 | 新キー `hosts/default/api_client_*` にコピーされ、flat キーは削除される（実装が Load 時移行でも Save 時移行でも可。テストで固定した経路を文書化する） |
| M3 | host `work`（`default` 以外）で Load・新キー無し・flat あり | **読み替えしない**（flat は default 専用）。`ErrSecretNotFound` |
| M4 | 新キーが既にあるとき flat が残っていても | **新キー優先**。flat は触らないか、クリーンアップしてよい（優先は新キー） |

#### テスト: 異常系

| # | 条件 | 期待 |
|---|------|------|
| M5 | flat の一方だけ欠落 | 読み替え失敗（不完全ペアを新キーへ移さない） |

---

## 3. `vault_unlock` 不透明性

#### テスト：正常系

| # | 操作 | 期待 |
|---|------|------|
| V1 | 任意の非空文字列を保存して再取得 | バイト列／文字列が一致（形式は実装の版付き blob でよい） |
| V2 | 空文字の Set | 拒否するか、Lock 相当として Delete 扱いするか実装で固定。MP を意味する値は受け取らない |

#### テスト: 異常系

| # | 条件 | 期待 |
|---|------|------|
| V3 | 欠落時 Get | `ErrSecretNotFound` |

---

## 4. 既存 `auth`（flat コマンド）との接続

#174 の `auth login` / `logout` CLI 契約は [`../cmd/auth_login_logout.md`](../cmd/auth_login_logout.md)。本仕様は Keychain キー形状のみ。login / logout が書く場合も **解決済み host id** で §1 のキーを使う。

#### テスト：正常系

| # | 操作 | 期待 |
|---|------|------|
| A1 | 解決 host `default` で auth 保存 | `hosts/default/api_client_*` に保存 |
| A2 | `--host work` で auth 保存 | `hosts/work/api_client_*` に保存。`default` を上書きしない |

#### テスト: 異常系

| # | 条件 | 期待 |
|---|------|------|
| A3 | host 解決失敗 | Keychain を変更せずエラー |

---

## 対象外

- `bwsf auth login` / `auth logout` / 引数なし `auth` ヘルプ化（#174 → [`../cmd/auth_login_logout.md`](../cmd/auth_login_logout.md)）
- `bwsf init`（#193）
- 設定ファイルの flat→v2 写像（#177 / [`migrate_v2.md`](../config/migrate_v2.md)）
- SDK `ExportSession` / `RestoreSession` の詳細（[`vault_unlock_restore.md`](../core/vault_unlock_restore.md) / CryptoSession 実装）
