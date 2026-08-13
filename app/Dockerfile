FROM golang:1.25-alpine

WORKDIR /app

# Node.jsとnpmをインストール（bwコマンド用）
RUN apk add --no-cache nodejs npm

# Bitwarden CLIをグローバルにインストール
RUN npm install -g @bitwarden/cli

# 開発に必要なツールをインストール（curl は smoke-ready の HTTPS 疎通確認用）
RUN apk add --no-cache git make curl ca-certificates

# Goモジュールのキャッシュを有効活用するため、go.mod/go.sumを先にコピー
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# 依存関係を整理（新しいパッケージが追加された場合に対応）
RUN go mod tidy

# 開発用のデフォルトコマンド（docker-composeで上書き可能）
CMD ["go", "run", "src/main.go"]

