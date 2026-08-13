#!/usr/bin/env bash
# Minimal bootstrap for Codespaces / Dev Containers (#111).
set -euo pipefail

echo "[devcontainer] Installing Bitwarden CLI (bw)..."
npm install -g @bitwarden/cli

echo "[devcontainer] Ensuring make is available..."
if ! command -v make >/dev/null 2>&1; then
  sudo apt-get update
  sudo apt-get install -y --no-install-recommends make
fi

echo "[devcontainer] Downloading Go modules..."
(cd app && go mod download)

echo "[devcontainer] Versions:"
go version
node --version
npm --version
bw --version || true
docker version --format 'docker {{.Server.Version}}' 2>/dev/null || docker version || true
docker compose version || true

echo "[devcontainer] Ready. Try: make test"
echo "[devcontainer] Optional smoke: make smoke-up && make smoke-ready && make smoke"
