.PHONY: help build up down run shell clean rebuild test test-unit test-e2e test-coverage fmt vet lint deps-outdated

# 変数定義
APP_DIR := app
SHELL := /bin/bash
.DEFAULT_GOAL := help

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

# ========================================
# GitHub ruleset helpers
# ========================================

include ruleset.mk

# サブコマンドをターゲットとして認識させないようにする（末尾に置く）
%:
	@:
