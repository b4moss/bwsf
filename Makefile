.PHONY: help build up down run shell clean test test-unit test-e2e fmt vet lint deps-outdated

# 変数定義
APP_DIR := app

# デフォルトターゲット
help: ## このヘルプメッセージを表示
	@echo "利用可能なコマンド:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Dockerイメージをビルド
	cd $(APP_DIR) && docker compose build

up: ## コンテナを起動して実行
	cd $(APP_DIR) && docker compose up

down: ## コンテナを停止・削除
	cd $(APP_DIR) && docker compose down

run: ## アプリケーションを実行（サブコマンドを指定可能: make run setup）
	cd $(APP_DIR) && docker compose run --rm golang go run src/main.go $(filter-out $@,$(MAKECMDGOALS))

shell: ## コンテナ内でシェルを起動
	cd $(APP_DIR) && docker compose run --rm golang sh

# ========================================
# テスト
# ========================================

test: ## 全てのテストを実行
	cd $(APP_DIR) && docker compose run --rm golang go test ./...

test-unit: ## ユニットテストのみ実行
	cd $(APP_DIR) && docker compose run --rm golang go test ./src/cmd/... ./src/config/... ./src/core/... ./src/infra/... ./src/utils/...

test-e2e: ## E2Eテストを実行（モックベース）
	cd $(APP_DIR) && docker compose run --rm golang go test -v ./src/e2e/...

test-coverage: ## カバレッジレポートを生成
	cd $(APP_DIR) && docker compose run --rm golang go test -coverprofile=coverage.out ./...
	cd $(APP_DIR) && docker compose run --rm golang go tool cover -html=coverage.out -o coverage.html

# ========================================
# コード品質
# ========================================

fmt: ## コードをフォーマット
	cd $(APP_DIR) && docker compose run --rm golang go fmt ./...

vet: ## コードを静的解析
	cd $(APP_DIR) && docker compose run --rm golang go vet ./...

lint: fmt vet ## フォーマットと静的解析を実行

deps-outdated: ## 古い依存パッケージを確認
	cd $(APP_DIR) && docker compose run --rm golang go list -m -u all

# ========================================
# クリーンアップ
# ========================================

clean: ## コンテナ、イメージ、ボリュームを削除
	cd $(APP_DIR) && docker compose down -v --rmi local

rebuild: clean build ## クリーンビルドを実行

# サブコマンドをターゲットとして認識させないようにする
%:
	@:
# GitHub repository ruleset helpers.
# Real work lives in scripts/; this Makefile is a thin wrapper.
# Target/variable names are prefixed with ruleset- / RULESET_ so this
# file can be vendored via git subtree without colliding with host Makefiles.

SHELL := /bin/bash
.DEFAULT_GOAL := ruleset-help

RULESET_ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
RULESET_SCRIPTS := $(RULESET_ROOT_DIR)/scripts

RULESET_REPO ?=
RULESET_BRANCH ?= main
RULESET_VISIBILITY ?= public
RULESET_CREATE_FLAGS ?=

.PHONY: ruleset-help ruleset-create ruleset-apply ruleset-check

ruleset-help:
	@printf '%s\n' \
		'Targets:' \
		'' \
		'  make ruleset-create RULESET_REPO=OWNER/NAME [RULESET_VISIBILITY=public] [RULESET_CREATE_FLAGS="--clone"]' \
		'      Create a GitHub repo and apply rulesets.' \
		'' \
		'  make ruleset-apply RULESET_REPO=OWNER/NAME' \
		'      Apply/update rulesets on an existing repo.' \
		'' \
		'  make ruleset-check RULESET_REPO=OWNER/NAME [RULESET_BRANCH=main]' \
		'      List rulesets and check which rules apply to RULESET_BRANCH.' \
		'' \
		'Notes:' \
		'  - GitHub Free (org): rulesets work on public repos only.' \
		'  - Requires gh (repo admin) and jq.' \
		'  - Namespaced as ruleset-* / RULESET_* for subtree-safe includes.'

ruleset-create:
	@if [[ -z "$(RULESET_REPO)" ]]; then \
		echo "error: RULESET_REPO=OWNER/NAME is required" >&2; \
		echo "example: make ruleset-create RULESET_REPO=my-org/new-app" >&2; \
		exit 1; \
	fi
	@$(RULESET_SCRIPTS)/create-repo-with-rulesets.sh "$(RULESET_REPO)" "--$(RULESET_VISIBILITY)" $(RULESET_CREATE_FLAGS)

ruleset-apply:
	@if [[ -z "$(RULESET_REPO)" ]]; then \
		echo "error: RULESET_REPO=OWNER/NAME is required" >&2; \
		echo "example: make ruleset-apply RULESET_REPO=my-org/existing-app" >&2; \
		exit 1; \
	fi
	@$(RULESET_SCRIPTS)/apply-rulesets.sh --repo "$(RULESET_REPO)"

ruleset-check:
	@if [[ -z "$(RULESET_REPO)" ]]; then \
		echo "error: RULESET_REPO=OWNER/NAME is required" >&2; \
		echo "example: make ruleset-check RULESET_REPO=my-org/existing-app RULESET_BRANCH=main" >&2; \
		exit 1; \
	fi
	@echo "== ruleset list =="
	@gh ruleset list -R "$(RULESET_REPO)"
	@echo
	@echo "== ruleset check $(RULESET_BRANCH) =="
	@gh ruleset check "$(RULESET_BRANCH)" -R "$(RULESET_REPO)"
