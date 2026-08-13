#!/bin/sh
# Orchestrate bwsf command smoke against Vaultwarden (issue #110).
# Intended to run inside the golang compose service.
set -eu

CMD="${CMD:-all}"
TARGET="${TARGET:-vaultwarden}"
BACKEND="${BACKEND:-bw}"

SMOKE_SERVER_URL="${BWSF_SMOKE_VW_URL:-https://vaultwarden:80}"
SMOKE_EMAIL="${BWSF_SMOKE_EMAIL:-smoke@bwsf.local}"
SMOKE_PASSWORD="${BWSF_SMOKE_PASSWORD:-SmokePassw0rd!}"
SMOKE_NAME="${BWSF_SMOKE_NAME:-Smoke User}"
PROJECT_NAME="${BWSF_SMOKE_PROJECT:-bwsf-smoke}"
REPO_ROOT="${BWSF_SMOKE_REPO_ROOT:-/project-root}"
APP_DIR="${BWSF_SMOKE_APP_DIR:-/app}"
FIXTURE_ENV="${BWSF_SMOKE_FIXTURE:-${APP_DIR}/smoke/fixtures/sample.env}"
KEEP_TMP="${BWSF_SMOKE_KEEP_TMP:-0}"

export BWSF_PASSWORD="${BWSF_PASSWORD:-$SMOKE_PASSWORD}"
export NODE_EXTRA_CA_CERTS="${NODE_EXTRA_CA_CERTS:-${BWSF_SMOKE_CA:-/smoke-certs/ca.crt}}"

BWSF_BIN=""

build_bwsf() {
  BWSF_BIN="${SMOKE_ROOT}/bwsf"
  echo "==> build bwsf binary"
  go -C "${APP_DIR}" build -o "${BWSF_BIN}" ./src
}

bwsf() {
  # Use a prebuilt binary so cwd stays the smoke project dir (project name = basename).
  "${BWSF_BIN}" "$@"
}

die() {
  echo "smoke error: $*" >&2
  exit 1
}

assert_target() {
  case "${TARGET}" in
    vaultwarden) ;;
    *) die "unsupported TARGET=${TARGET} (only vaultwarden in #110)" ;;
  esac
}

assert_backend() {
  case "${BACKEND}" in
    bw) ;;
    api) die "BACKEND=api is reserved (not implemented in #110)" ;;
    *) die "unsupported BACKEND=${BACKEND}" ;;
  esac
}

prepare_tmp() {
  RUN_ID="$(date +%Y%m%d%H%M%S)-$$"
  SMOKE_ROOT="${REPO_ROOT}/.smoke-tmp/${RUN_ID}"
  PROJECT_DIR="${SMOKE_ROOT}/${PROJECT_NAME}"
  HOME_DIR="${SMOKE_ROOT}/home"
  PULL_DIR="${SMOKE_ROOT}/pulled"

  mkdir -p "${PROJECT_DIR}" "${HOME_DIR}" "${PULL_DIR}"
  export HOME="${HOME_DIR}"

  cp "${FIXTURE_ENV}" "${PROJECT_DIR}/.env"
  echo "smoke tmp: ${SMOKE_ROOT}"
}

bootstrap_user() {
  echo "==> bootstrap Vaultwarden user (${SMOKE_EMAIL})"
  node "${REPO_ROOT}/scripts/smoke-register.mjs" \
    "${SMOKE_SERVER_URL}" "${SMOKE_EMAIL}" "${SMOKE_PASSWORD}" "${SMOKE_NAME}"
}

step_setup() {
  echo "==> setup"
  cd "${PROJECT_DIR}"
  bwsf setup \
    --host-type selfhosted \
    --url "${SMOKE_SERVER_URL}" \
    --email "${SMOKE_EMAIL}" \
    --password "${SMOKE_PASSWORD}" \
    --yes
}

step_push() {
  echo "==> push"
  cd "${PROJECT_DIR}"
  bwsf push --from .
}

step_pull() {
  echo "==> pull"
  cd "${PROJECT_DIR}"
  bwsf pull --output "${PULL_DIR}"
  grep -q "SMOKE_MARKER=bwsf-smoke-ok" "${PULL_DIR}/.env" \
    || die "pull assertion failed: marker missing in ${PULL_DIR}/.env"
}

step_list() {
  echo "==> list"
  cd "${PROJECT_DIR}"
  out="$(bwsf list)"
  echo "${out}"
  echo "${out}" | grep -qx "${PROJECT_NAME}" \
    || die "list assertion failed: expected project name ${PROJECT_NAME}"
}

run_cmd() {
  case "$1" in
    setup) step_setup ;;
    push) step_push ;;
    pull) step_pull ;;
    list) step_list ;;
    all)
      step_setup
      step_push
      step_pull
      step_list
      ;;
    *) die "unsupported CMD=${1} (setup|push|pull|list|all)" ;;
  esac
}

cleanup_success() {
  if [ "${KEEP_TMP}" = "1" ]; then
    echo "smoke keep tmp: ${SMOKE_ROOT}"
    return
  fi
  rm -rf "${SMOKE_ROOT}"
  echo "smoke tmp removed"
}

main() {
  assert_target
  assert_backend

  # Ensure VW is reachable (caller usually ran smoke-ready already)
  sh "${REPO_ROOT}/scripts/smoke-ready.sh"

  prepare_tmp
  build_bwsf
  bootstrap_user

  # For single-command runs that need auth state, always setup first except CMD=setup
  case "${CMD}" in
    all|setup) ;;
    push|pull|list)
      step_setup
      ;;
  esac

  set +e
  run_cmd "${CMD}"
  status=$?
  set -e

  if [ "${status}" -ne 0 ]; then
    echo "smoke FAILED; tmp kept at: ${SMOKE_ROOT}" >&2
    exit 1
  fi

  cleanup_success
  echo "smoke OK (CMD=${CMD} TARGET=${TARGET} BACKEND=${BACKEND})"
}

main "$@"
