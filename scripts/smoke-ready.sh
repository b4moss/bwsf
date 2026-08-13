#!/bin/sh
# Wait until Vaultwarden HTTPS is reachable with the smoke CA.
# Intended to run inside the golang compose service (or any container with the CA mounted).
set -eu

URL="${BWSF_SMOKE_VW_URL:-https://vaultwarden}"
CA="${BWSF_SMOKE_CA:-/smoke-certs/ca.crt}"
RETRIES="${BWSF_SMOKE_READY_RETRIES:-60}"
SLEEP_SECS="${BWSF_SMOKE_READY_SLEEP:-2}"

if [ ! -f "${CA}" ]; then
  echo "smoke CA not found: ${CA}" >&2
  exit 1
fi

if ! command -v curl >/dev/null 2>&1; then
  echo "curl is required for smoke-ready" >&2
  exit 1
fi

i=1
while [ "${i}" -le "${RETRIES}" ]; do
  if curl -fsS --cacert "${CA}" "${URL}/alive" >/dev/null 2>&1; then
    echo "Vaultwarden is ready at ${URL} (attempt ${i}/${RETRIES})"
    exit 0
  fi
  sleep "${SLEEP_SECS}"
  i=$((i + 1))
done

echo "Vaultwarden not ready at ${URL} after ${RETRIES} attempts" >&2
exit 1
