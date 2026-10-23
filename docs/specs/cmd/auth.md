# auth login / logout

現行の `bwsf auth login` / `auth logout` の仕様正本。  
テスト仕様: [`../../tests/cmd/auth_login_logout.md`](../../tests/cmd/auth_login_logout.md)。

関連 Issue: [#174](https://github.com/b4moss/bwsf/issues/174)

## 決定

| 項目 | 決定 |
|------|------|
| コマンド | `auth login` / `auth logout` |
| 引数なし `auth` | ヘルプのみ |
| 旧フラット `auth` / `--clear` | 削除 |
| 対象 | [`../config/host-resolve.md`](../config/host-resolve.md) で解決。`--host` 可。`logout --all` あり |
| `logout` | API Key + `vault_unlock` を削除 |
| `login` 成功時 | Key 保存 → Identity 確認 → unlock まで一気通貫 |
| 同時 login | 可 |

| コマンド | 責務 |
|----------|------|
| `auth login` / `logout` | API Key（logout はセッションも） |
| `unlock` / `lock` | vault セッションのみ |

将来 driver 追加時に見直しうる。

## 関連

- unlock / lock: [`unlock-lock.md`](./unlock-lock.md)
- 秘密ストア: [`../infra/secretstore-hosts.md`](../infra/secretstore-hosts.md)
- 履歴: [`../../_archived/history/v0.20.0-multi-host.md`](../../_archived/history/v0.20.0-multi-host.md) §4

----

以上
