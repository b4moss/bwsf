# core / infra: `vault_unlock` 自動 restore（Issue #153 / v0.20.0 §3）

対象パッケージ: `app/src/core`（`WithUnlockRetry` 等）、`app/src/infra`（`ApiBwClient` / `CryptoSession`）  
Issue: [#153](https://github.com/b4moss/bwsf/issues/153)  
製品正本: [`../specs/v0.20.0-multi-host.md`](../specs/v0.20.0-multi-host.md) §3

関連:

- Keychain キー → [`../infra/secretstore_hosts.md`](../infra/secretstore_hosts.md)
- CLI unlock/lock → [`../cmd/unlock_lock.md`](../cmd/unlock_lock.md)
- 旧 Step 3 再試行 → [`unlock_retry_api.md`](./unlock_retry_api.md)（**restore 挿入後も認証切れ vs 未 unlock の分岐は維持**）
- メモリ ClearSession → [`../cmd/session_lifecycle.md`](../cmd/session_lifecycle.md)、[`../infra/apiclient_unlock.md`](../infra/apiclient_unlock.md)

---

### 前提（本仕様で固定）

| # | 方針 |
|---|------|
| R1 | vault 系コマンド（**`push` / `pull` / `list` / `clean`**）は、host 解決後に当該 host の **`vault_unlock` を自動 restore** してから vault 操作する |
| R2 | restore 成功時は **MP プロンプトを出さない** |
| R3 | `vault_unlock` が無い、または restore 失敗／復元後の初回 vault 操作が「無効セッション」を示す場合は、**当該 `vault_unlock` を破棄**し、従来どおり MP プロンプト → Unlock → 再実行へ進む |
| R4 | Unlock（プロンプト経由）成功後は、新しい不透明データを Keychain の `vault_unlock` に保存する |
| R5 | 不透明データの中身は実装詳細（SDK `ExportSession` / `RestoreSession` の `SessionMaterial` を版付きで符号化してよい）。テストは **ラウンドトリップと「無効なら破棄」** を固定し、内部フィールドは問わない |
| R6 | マスターパスワードは Keychain に保存しない |
| R7 | コマンド終了時の `ClearSession` は **プロセスメモリのみ**。Keychain の `vault_unlock` / API Key は残す |
| R8 | 解決 host 以外の `vault_unlock` は読まない・消さない（無効破棄も当該 host のみ） |
| R9 | 認証切れ（API Key / Identity）は restore や MP より先に auth 案内（[`unlock_retry_api.md`](./unlock_retry_api.md)） |

```text
host 解決
  → API 認証（Keychain API Key）
  → vault_unlock があれば Restore
       ├ 成功 → vault 操作
       │         └ 無効セッション → 破棄 → MP → Unlock → 保存 → 再実行
       └ 無し/失敗 → MP → Unlock → 保存 → vault 操作
  → 終了時 ClearSession（メモリのみ）
```

---

## 1. CryptoSession / ApiBwClient：Export・Restore

モックまたはテストダブルで Export/Restore を差し替え可能にする。

#### テスト：正常系

| # | 操作 | 期待 |
|---|------|------|
| E1 | Unlock 成功後に Export | 非空の不透明 blob。MP を含まない（少なくとも平文 MP 文字列がblobに無いことをスポット確認してよい） |
| E2 | ClearSession（メモリ）後に同じ blob で Restore | 再び `IsUnlocked` |
| E3 | Restore 後に vault モック操作が成功 | 鍵が使える |

#### テスト: 異常系

| # | 条件 | 期待 |
|---|------|------|
| E4 | 改ざん・空・未知バージョンの blob で Restore | エラー。unlocked にならない |
| E5 | 未 Unlock で Export | エラー |

---

## 2. Keychain 連携（Unlock 時保存）

#### テスト：正常系

| # | 操作 | 期待 |
|---|------|------|
| S1 | `ApiBwClient.Unlock` 成功 | `hosts/<hostID>/vault_unlock` に blob が保存される |
| S2 | 別 host のクライアントで Unlock | 他 host のキーを上書きしない |

#### テスト: 異常系

| # | 条件 | 期待 |
|---|------|------|
| S3 | Unlock 失敗 | `vault_unlock` を新規作成しない（既存値は維持） |

---

## 3. `WithUnlockRetry` / vault コマンド入口の自動 restore

`push` / `pull` / `list` / `clean` 相当の core 入口、または共有ヘルパをテストする。

#### テスト：正常系

| # | 条件 | 期待 |
|---|------|------|
| A1 | 有効な `vault_unlock` あり | MP プロンプト **0 回**。vault 操作成功 |
| A2 | `vault_unlock` 無し | MP 1 回 → Unlock → 保存 → 成功 |
| A3 | 無効 blob（Restore 失敗） | blob 削除 → MP プロンプト → 成功後に新 blob |
| A4 | Restore は成功するが直後の vault 操作が無効セッション | blob 削除 → MP → 再 Unlock → 再実行成功 |
| A5 | host `work` を解決 | `hosts/work/vault_unlock` のみ使用。`default` の blob では restore しない |

#### テスト: 異常系

| # | 条件 | 期待 |
|---|------|------|
| A6 | API Key 無し | auth 案内。MP も restore もしない（または restore 前に失敗） |
| A7 | 無効 blob 破棄後、MP も失敗 | エラー。無限ループしない（再試行回数は現行 Unlock 再試行方針に従う） |
| A8 | 秘密情報（API Key / MP / blob）がログ・エラーに出ない | 出力検査 |

---

## 4. `ClearSession` との関係

#### テスト：正常系

| # | 操作 | 期待 |
|---|------|------|
| C1 | Unlock（保存済み）→ ClearSession | メモリは locked。Keychain の `vault_unlock` と API Key は残る |
| C2 | その後同じクライアント（または新インスタンス）で TryRestore | 再び unlocked にできる |

---

## 対象外

- `bwsf unlock` / `lock` CLI の引数表（[`unlock_lock.md`](../cmd/unlock_lock.md)）
- `auth logout` による API Key + `vault_unlock` 一括削除（#174 → [`../cmd/auth_login_logout.md`](../cmd/auth_login_logout.md)）
- `bwsf init`（#193）
- 旧 `BW_SESSION` / darwin SessionStore の維持・廃止判断（API 経路の契約外。壊さない範囲で放置可）
