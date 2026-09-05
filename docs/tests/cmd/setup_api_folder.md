# cmd: setup（api）のフォルダ作成（Issue #53 Step 4）

> **v0.20.0 / #177:** フォルダ名は host の `target_section`。setup フロー本体は [`setup_v2.md`](./setup_v2.md)。本文書の「未認証ならスキップ／`--yes` で作成」の精神は維持。

対象パッケージ: `app/src/cmd` / `app/src/core`（`SetupAPIConfigCore` または後継）  
合意: api でも setup から設定フォルダを作成する（Q27）。push/pull/list はフォルダ欠落時に自動作成しない（Q23）。

前提: Step 3 の「setup は設定のみ・Login しない」は維持。Step 4 で **vault が使えるようになった後**、フォルダ作成確認を api 経路に接続する。

---

### runSetupAPI / SetupAPIConfigCore（フォルダ）

- host/email 等の保存後（または一連の setup 内で）、設定フォルダの有無を確認する。
- 無い場合は作成確認（対話 y/N、または非対話 `--yes`）のうえ `CreateDotenvsFolder` を呼ぶ。
- Personal API Key の Login（email/password）は行わない。必要なら setup 中に auth/unlock 相当が必要になる場合は、既存 `bwsf auth` + unlock 再試行方針と矛盾しない形にする（実装が setup 内で Unlock するなら、その前提をテストで固定）。

#### テスト：正常系

- フォルダが既にある場合、作成を呼ばず setup 成功。
- フォルダが無く、確認で yes（または `--yes`）のとき `CreateDotenvsFolder` が呼ばれ成功する。
- フォルダが無く、確認で no のとき作成せず setup は設定保存までは成功（または既存 bw setup と同様の終了方針をテストで固定）。
- `Login`（email/password）は呼ばれない（Step 3 退行）。

#### テスト: 異常系

- `CreateDotenvsFolder` が失敗した場合、エラーを返し、成功メッセージを出さない。
- vault 未 unlock / 未 auth でフォルダ確認ができない場合、認証案内または unlock 経路に適切に分岐し、秘密情報を出さない。
- 非対話で `--yes` 無し・確認不能な場合に作成を強行しない。

---

### 退行: backend=bw の setup フォルダ作成

#### テスト：正常系

- `backend=bw` では従来どおり folder 確認・作成が動作する（既存テストで担保可）。

#### テスト: 異常系

- api 専用の folder 作成ロジックが bw 経路の Login を壊さない。
