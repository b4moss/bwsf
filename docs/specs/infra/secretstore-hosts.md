# 秘密ストア（host 単位キー）

現行の Keychain / 秘密ストアキー配置の仕様正本。  
テスト仕様: [`../../tests/infra/secretstore_hosts.md`](../../tests/infra/secretstore_hosts.md)。

関連 Issue: [#153](https://github.com/b4moss/bwsf/issues/153)

## 決定

| 項目 | 決定 |
|------|------|
| キー | `hosts/<id>/api_client_id`, `hosts/<id>/api_client_secret`, `hosts/<id>/vault_unlock` |
| API Key | host 単位。複数同時保持可 |
| `vault_unlock` | unlock 状態復元用の不透明データ。MP は保存しない |

旧 Keychain グローバルキーは `hosts/default/...` へ移行または読み替え（設定マイグレーションと連動。[`../config/global-v2.md`](../config/global-v2.md)）。

## 関連

- unlock / lock: [`../cmd/unlock-lock.md`](../cmd/unlock-lock.md)
- 自動 restore: [`../core/vault-unlock-restore.md`](../core/vault-unlock-restore.md)
- auth: [`../cmd/auth.md`](../cmd/auth.md)
- 履歴: [`../../_archived/history/v0.20.0-multi-host.md`](../../_archived/history/v0.20.0-multi-host.md) §3

----

以上
