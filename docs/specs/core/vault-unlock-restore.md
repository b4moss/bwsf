# vault_unlock 自動 restore

現行の vault 系コマンドにおける `vault_unlock` 自動 restore の仕様正本。  
テスト仕様: [`../../tests/core/vault_unlock_restore.md`](../../tests/core/vault_unlock_restore.md)。

関連 Issue: [#153](https://github.com/b4moss/bwsf/issues/153)

## 決定

| 項目 | 決定 |
|------|------|
| 自動 restore | vault 系コマンド（push / pull / list / clean）で、解決 host の `vault_unlock` を利用 |
| 無効時 | 破棄し再プロンプトまたは `unlock` へ |

## 関連

- 秘密ストア: [`../infra/secretstore-hosts.md`](../infra/secretstore-hosts.md)
- unlock / lock: [`../cmd/unlock-lock.md`](../cmd/unlock-lock.md)
- 履歴: [`../../_archived/history/v0.20.0-multi-host.md`](../../_archived/history/v0.20.0-multi-host.md) §3

----

以上
