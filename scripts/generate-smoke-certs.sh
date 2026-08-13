#!/usr/bin/env bash
# Regenerate self-signed TLS material for the Vaultwarden smoke stack.
# Usage: ./scripts/generate-smoke-certs.sh
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
OUT_DIR="${ROOT_DIR}/app/smoke/certs"
DAYS="${SMOKE_CERT_DAYS:-3650}"

mkdir -p "${OUT_DIR}"
TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout "${TMP}/ca.key" -out "${TMP}/ca.crt" -days "${DAYS}" \
  -subj "/CN=bwsf-smoke-CA"

openssl req -newkey rsa:2048 -nodes \
  -keyout "${TMP}/vaultwarden.key" -out "${TMP}/vaultwarden.csr" \
  -subj "/CN=vaultwarden"

openssl x509 -req -in "${TMP}/vaultwarden.csr" \
  -CA "${TMP}/ca.crt" -CAkey "${TMP}/ca.key" -CAcreateserial \
  -out "${TMP}/vaultwarden.crt" -days "${DAYS}" \
  -extfile <(printf "subjectAltName=DNS:vaultwarden,DNS:localhost,IP:127.0.0.1")

install -m 0644 "${TMP}/ca.crt" "${OUT_DIR}/ca.crt"
install -m 0600 "${TMP}/ca.key" "${OUT_DIR}/ca.key"
install -m 0644 "${TMP}/vaultwarden.crt" "${OUT_DIR}/vaultwarden.crt"
install -m 0600 "${TMP}/vaultwarden.key" "${OUT_DIR}/vaultwarden.key"

echo "Wrote smoke certs to ${OUT_DIR}"
openssl x509 -in "${OUT_DIR}/vaultwarden.crt" -noout -subject -dates
