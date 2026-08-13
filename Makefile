.PHONY: help build up down run shell clean rebuild test test-unit test-e2e test-coverage fmt vet lint deps-outdated smoke-up smoke-down smoke-ready smoke

# 変数定義
APP_DIR := app
SHELL := /bin/bash
.DEFAULT_GOAL := help

# Smoke / real-server targets (Vaultwarden). CMD/TARGET/BACKEND are reserved for #110.
CMD ?=
TARGET ?= vaultwarden
BACKEND ?= bw

# デフォルトターゲット
help: ## このヘルプメッセージを表示
	@echo "利用可能なコマンド:"
	@grep -hE '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

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
# テスト（サーバ不要）
# ========================================

test: ## サーバ不要の全テスト（unit + mock e2e。スモークは含まない）
	cd $(APP_DIR) && docker compose run --rm golang go test ./...

test-unit: ## ユニット／結合テストのみ（サーバ不要）
	cd $(APP_DIR) && docker compose run --rm golang go test ./src/cmd/... ./src/config/... ./src/core/... ./src/infra/... ./src/utils/...

test-e2e: ## モック E2E（サーバ不要。実 Vaultwarden は make smoke*）
	cd $(APP_DIR) && docker compose run --rm golang go test -v ./src/e2e/...

test-coverage: ## カバレッジレポートを生成（サーバ不要）
	cd $(APP_DIR) && docker compose run --rm golang go test -coverprofile=coverage.out ./...
	cd $(APP_DIR) && docker compose run --rm golang go tool cover -html=coverage.out -o coverage.html

# ========================================
# スモーク土台（実 Vaultwarden / #109）
# ========================================

smoke-up: ## Vaultwarden（smoke profile）を起動
	cd $(APP_DIR) && docker compose --profile smoke up -d vaultwarden

smoke-down: ## スモーク用サービスを停止
	cd $(APP_DIR) && docker compose --profile smoke down

smoke-ready: ## テスト用コンテナから VW HTTPS 疎通を確認
	cd $(APP_DIR) && docker compose --profile smoke run --rm --no-deps golang sh /project-root/scripts/smoke-ready.sh

smoke: ## 実コマンド通しスモーク（#110 で実装。現時点はスタブ）
	@echo "make smoke is not implemented yet (see #110)." >&2
	@echo "Reserved vars: CMD=$(CMD) TARGET=$(TARGET) BACKEND=$(BACKEND)" >&2
	@echo "Use: make smoke-up && make smoke-ready && make smoke-down" >&2
	@exit 1

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

clean: ## コンテナ、イメージ、ボリュームを削除（smoke profile 含む）
	cd $(APP_DIR) && docker compose --profile smoke down -v --rmi local

rebuild: clean build ## クリーンビルドを実行

# ========================================
# GitHub ruleset helpers
# ========================================

include ruleset.mk

# サブコマンドをターゲットとして認識させないようにする（末尾に置く）
%:
	@:
