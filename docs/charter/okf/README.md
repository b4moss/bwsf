---
type: Charter
title: OKF v0.1
description: プロダクト知識バンドル（OKF）v0.1 の定義。
tags: [charter, okf]
timestamp: 2026-09-27T00:00:00Z
---
# OKF v0.1

**OKF（Org Knowledge Framework）v0.1** は、本組織のプロダクトが `docs/` 配下に持つ**プロダクト知識バンドル**の版である。  
版の識別は各プロダクトの [`docs/index.md`](../../index.md) フロントマター `okf_version: "0.1"` で行う。

| ファイル | 役割 |
|----------|------|
| `docs/index.md` | **OKF の索引のみ**（版フロントマター＋リンク）。意味的なプロダクト正本ではない |
| `docs/README.md` | **意味的な pillar 正本**（目的・スコープ・技術方針のハブ）。旧称 `main.md` |

配置・ライフサイクルの詳細正本は [docs 整理ルール（doc-rule）](../doc-rule.md)。本ページは OKF としての要約と、コピー用の[執筆サンプル](./samples/)への入口である。

---

## 最小ツリー（必須 / 推奨）

| パス | 必須 | 役割 |
|------|:----:|------|
| `docs/index.md` | ✅ | OKF 索引。`okf_version: "0.1"` を付ける（プロダクト本文は書かない） |
| `docs/charter/` | ✅ | 憲章（本リポジトリから取り込む正本） |
| `docs/override-charter.md` | ✅ | 憲章より優先するプロジェクト固有ルール（空でも可） |
| `docs/plans/` | ✅ | **これからやる**内容（版フォルダで切る） |
| `docs/specs/` | ✅ | **現行に存在する**機能の仕様正本（ドメインで切る） |
| `docs/README.md` | 推奨 | プロダクト目的・スコープ・技術方針の pillar（正） |
| `docs/roadmap.md` | 推奨 | SemVer・マイルストーン一覧のハブ |
| `docs/wishlist.md` | 推奨 | PO の未整理メモ |
| `docs/tests/` | 推奨 | テスト仕様（TDD 入力）。`specs/` と同じドメイン切り |
| `docs/_archived/` | 任意 | 削除・置換された仕様や歴史資料 |

charter リポジトリ自体はプロダクト要件を持たないため、`README.md`（pillar）/ `roadmap.md` / `tests/` 等はプレースホルダやサンプルに留めてよい。各プロダクトでは上表に従って揃える。

---

## OKF v0.1 の硬いルール（要約）

詳細は [doc-rule](../doc-rule.md)。ここではプロダクトが迷わない要点だけ示す。

1. **実装済みは `specs/` に置く。** 実装が終わった内容を `plans/` に残さない（plans → specs へ移動）。
2. **`specs/` と `tests/` は同じドメイン切り。** 例: `specs/auth/` と `tests/auth/`。
3. **現行の `specs/` / `tests/` を SemVer フォルダ（`vX.Y.Z/`）で切らない。** 版フォルダは `plans/`（必要なら履歴は `_archived/`）に置く。

ライフサイクルのワンライナー（doc-rule より）:

> 思いつきは wishlist、やることは plans（薄くて可）、動いていることは specs、消えたことは archived、守ることは charter。

---

## 執筆サンプル

コピーして使う短い Markdown スタブ:

| サンプル | 想定パス |
|----------|----------|
| [OKF 索引](./samples/index.md) | `docs/index.md` |
| [pillar（README）](./samples/docs-readme.md) | `docs/README.md` |
| [plans 意図スタブ](./samples/plans-intent-stub.md) | `docs/plans/vX.Y.Z/…` または `docs/plans/unscheduled/…` |
| [specs ドメイン仕様](./samples/specs-domain.md) | `docs/specs/{domain}/…` |
| [tests ドメイン仕様](./samples/tests-domain.md) | `docs/tests/{domain}/…`（specs と同じ domain） |

----

以上
