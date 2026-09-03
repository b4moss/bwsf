# cmd: setup の api 分離（Issue #53 Step 3 前提 PR-B）

対象パッケージ: `app/src/cmd` / 必要なら `app/src/core` の Setup  
合意: `backend=api` のとき setup は設定のみ。認証は `bwsf auth`。

---

### runSetup（backend=api）

- `backend=api`（またはセット後に api）では、Bitwarden への `Login`（email/password）を行わない。
- host_type / selfhosted_url / email 等、設定ファイルに必要な項目の入力・保存のみ行う。
- 完了後、api 利用には `bwsf auth` が必要である旨を案内する。
- フォルダ作成確認は Step 3 ではスキップしてよい。**本実装とテストは Step 4**（[`setup_api_folder.md`](./setup_api_folder.md)）。

#### テスト：正常系

- `backend=api` で setup を実行すると、config が保存され、`Login` が呼ばれない。
- setup 後のメッセージ／ログに `bwsf auth` への案内が含まれる。
- `backend=bw` では従来どおり Login を含むフローが維持される（退行）。

#### テスト: 異常系

- 設定保存に失敗した場合、Login を呼ばずエラー終了する（api）。
- 必須入力（例: selfhosted 時の URL）が空の場合、保存せずエラーになる。
- 既に `backend=api` かつ auth 未実施でも setup 自体は設定更新として成功できる（auth 必須にしない）。

---

### runSetup（backend=bw）退行

- 既存 setup の振る舞いを壊さないことの確認用。

#### テスト：正常系

- `backend=bw`（default）で従来の host 選択 → email → password → Login → 設定保存が動作する（既存テストで担保可）。

#### テスト: 異常系

- Login 失敗時に設定を中途半端に確定しない、または既存仕様どおりエラーになる（既存テストに合わせる）。
