#!/usr/bin/env bash
# Sync retained Homebrew versioned formulas into the tap (see #171).
#
# Retention: all patches of the current minor + latest patch of the previous
# minor present on GitHub Releases. Prereleases are skipped. Out-of-policy
# bwsf@*.rb formulas are deleted from the tap.
#
# Usage:
#   ./scripts/sync-homebrew-versioned-formulas.sh [--push] [--tap-dir DIR]
#
# Without --push, formulas are written to a local tap checkout only.
# With --push, changes are committed and pushed to the tap default branch.
#
# Auth: uses `gh` (set GH_TOKEN / HOMEBREW_TAP_TOKEN for non-interactive CI).
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TAP_REPO="${HOMEBREW_TAP_REPO:-b4moss/homebrew-tap}"
TAP_DIR=""
DO_PUSH=0

# Prefer an explicit tap token in CI so push can reach the tap repo.
if [[ -n "${HOMEBREW_TAP_TOKEN:-}" ]]; then
  export GH_TOKEN="${HOMEBREW_TAP_TOKEN}"
fi

while [[ $# -gt 0 ]]; do
  case "$1" in
    --push)
      DO_PUSH=1
      shift
      ;;
    --tap-dir)
      TAP_DIR="$2"
      shift 2
      ;;
    -h|--help)
      sed -n '2,16p' "$0"
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

if [[ -z "${TAP_DIR}" ]]; then
  TAP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/bwsf-homebrew-tap.XXXXXX")"
  cleanup() { rm -rf "${TAP_DIR}"; }
  # Keep checkout when --push so git ops succeed; still clean ephemeral clones without --push.
  if [[ "${DO_PUSH}" -eq 0 ]]; then
    trap cleanup EXIT
  fi
  echo "cloning ${TAP_REPO} into ${TAP_DIR}"
  gh repo clone "${TAP_REPO}" "${TAP_DIR}" -- --depth 1
else
  TAP_DIR="$(cd "${TAP_DIR}" && pwd)"
fi

python3 "${ROOT_DIR}/scripts/sync-homebrew-versioned-formulas.py" --tap-dir "${TAP_DIR}"

if [[ "${DO_PUSH}" -eq 0 ]]; then
  echo "formulas written under ${TAP_DIR} (pass --push to commit and push)"
  exit 0
fi

git -C "${TAP_DIR}" config user.name "${GIT_AUTHOR_NAME:-bwsf-bot}"
git -C "${TAP_DIR}" config user.email "${GIT_AUTHOR_EMAIL:-bwsf-bot@users.noreply.github.com}"

# Stage adds, updates, and deletions of versioned formulas only.
git -C "${TAP_DIR}" add -A -- 'bwsf@*.rb'

if git -C "${TAP_DIR}" diff --cached --quiet; then
  echo "no tap changes to push"
  exit 0
fi

git -C "${TAP_DIR}" commit -m "Sync retained bwsf versioned Homebrew formulas"

# Ensure git can push over HTTPS with GH_TOKEN / gh credentials.
if command -v gh >/dev/null 2>&1; then
  gh auth setup-git >/dev/null 2>&1 || true
fi
git -C "${TAP_DIR}" push origin HEAD
echo "pushed versioned formulas to ${TAP_REPO}"
