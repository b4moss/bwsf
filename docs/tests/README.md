# テスト仕様書（索引）

本ディレクトリは charter（TDD）に基づく **テスト仕様書** の置き場です。  
実装前にここを FIX し、その後 Red → Green → Refactor で進めます。

既存の横断メモ: [`../TEST.md`](../TEST.md)（リファクタ当時の一括仕様。新規は本ディレクトリを優先）。

## 構成

| パス | 対象 |
|------|------|
| [`infra/`](./infra/) | インフラ層（`ApiBwClient`、Identity、SecretStore 等） |
| [`core/`](./core/) | コア層（再試行・エラー判定等） |
| [`cmd/`](./cmd/) | CLI 層（setup / auth / セッション寿命） |

## Issue #53 Step 3

| 文書 | 内容 |
|------|------|
| [`infra/apiclient_unlock.md`](./infra/apiclient_unlock.md) | Unlock / ClearSession / 鍵状態 |
| [`core/unlock_retry_api.md`](./core/unlock_retry_api.md) | 認証切れと未 unlock の分岐再試行 |
| [`cmd/setup_api.md`](./cmd/setup_api.md) | `backend=api` 時の setup 分離 |
| [`cmd/session_lifecycle.md`](./cmd/session_lifecycle.md) | コマンド終了時のセッション破棄 |

Step 3 の実装計画正本: [Issue #53 Step 3 実装計画](https://github.com/b4moss/bwsf/issues/53#issuecomment-5276317436)
