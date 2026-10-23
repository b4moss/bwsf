# unlock / lock

現行の `bwsf unlock` / `lock` の仕様正本。  
テスト仕様: [`../../tests/cmd/unlock_lock.md`](../../tests/cmd/unlock_lock.md) / [`../../tests/cmd/session_lifecycle.md`](../../tests/cmd/session_lifecycle.md)。

関連 Issue: [#153](https://github.com/b4moss/bwsf/issues/153)

## 決定

| コマンド | 決定 |
|----------|------|
| `unlock` | host は [`../config/host-resolve.md`](../config/host-resolve.md) で解決。`--host` 可 |
| `lock` | 解決した host の `vault_unlock` のみ削除 |
| `lock --all` | 登録済み全 host の `vault_unlock` 削除（`hosts` が空なら no-op 成功） |

| 項目 | 決定 |
|------|------|
| 有効期間 | `lock` / `auth logout` まで（再起動後も残る） |
| 責務境界 | `unlock` / `lock` は vault セッションのみ。API Key は触らない |

終了時はメモリ上のセッションのみ破棄。Keychain の `vault_unlock` は残す。

## 関連

- 秘密ストア: [`../infra/secretstore-hosts.md`](../infra/secretstore-hosts.md)
- 自動 restore: [`../core/vault-unlock-restore.md`](../core/vault-unlock-restore.md)
- auth: [`auth.md`](./auth.md)
- 履歴: [`../../_archived/history/v0.20.0-multi-host.md`](../../_archived/history/v0.20.0-multi-host.md) §3

----

以上
