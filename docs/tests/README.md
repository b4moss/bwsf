# テスト仕様書（索引）

本ディレクトリは charter（TDD）に基づく **テスト仕様書** の置き場です。  
実装前にここを FIX し、その後 Red → Green → Refactor で進めます。

既存の横断メモ: [`../TEST.md`](../TEST.md)（リファクタ当時の一括仕様。新規は本ディレクトリを優先）。

## 構成

| パス | 対象 |
|------|------|
| [`infra/`](./infra/) | インフラ層（`ApiBwClient`、Identity、SecretStore、保管庫 CRUD 等） |
| [`core/`](./core/) | コア層（再試行・push/pull/list 接続等） |
| [`cmd/`](./cmd/) | CLI 層（setup / auth / セッション寿命） |

## Issue #53 Step 3

| 文書 | 内容 |
|------|------|
| [`infra/apiclient_unlock.md`](./infra/apiclient_unlock.md) | Unlock / ClearSession / 鍵状態 |
| [`core/unlock_retry_api.md`](./core/unlock_retry_api.md) | 認証切れと未 unlock の分岐再試行 |
| [`cmd/setup_api.md`](./cmd/setup_api.md) | `backend=api` 時の setup 分離 |
| [`cmd/session_lifecycle.md`](./cmd/session_lifecycle.md) | コマンド終了時のセッション破棄 |

Step 3 の実装計画正本: [Issue #53 Step 3 実装計画](https://github.com/b4moss/bwsf/issues/53#issuecomment-5276317436)

## Issue #53 Step 4

| 文書 | 内容 |
|------|------|
| [`infra/apiclient_vault.md`](./infra/apiclient_vault.md) | sync / folder / Secure Note CRUD（`ErrAPINotImplemented` 撤去） |
| [`cmd/setup_api_folder.md`](./cmd/setup_api_folder.md) | api setup からの設定フォルダ作成 |
| [`core/vault_ops_api.md`](./core/vault_ops_api.md) | push / pull / list を api adapter に接続したときの伝播・退行 |

Step 4 の実装計画正本: [Issue #53 Step 4 実装計画](https://github.com/b4moss/bwsf/issues/53#issuecomment-5277440346)

### Step 4 で扱わない（仕様書にも本実装テストを置かない）

- `clean` の api 対応
- 組織ボルト / SSO
- `BACKEND=api` の `make smoke` 本実装

## Issue #53 Step 5

仕上げ（default=`api`、docs、spike 削除、`bw` は残置）。**新規テスト仕様書は起こさない。**

実装計画正本: [Issue #53 Step 5 実装計画](https://github.com/b4moss/bwsf/issues/53#issuecomment-5278247151)
