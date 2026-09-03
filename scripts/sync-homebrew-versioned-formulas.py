#!/usr/bin/env python3
"""Sync retained Homebrew versioned formulas into a tap checkout.

Retention policy (from v0.18.0 / #171):
  - Keep every stable patch of the current minor (e.g. 0.18.0 .. 0.18.N)
  - Keep only the latest stable patch of the previous minor that exists on
    GitHub Releases (e.g. 0.17.3). After a major bump, the previous line is
    the latest minor of the prior major (e.g. 1.0.0 → keep 0.17.latest).
  - Exclude drafts, GitHub prereleases, and non-stable tags (rc/beta/…).
  - Physically delete tap formulas (bwsf@*.rb) outside the retain set.

Writes bwsf@<version>.rb so retained releases remain installable via:

  brew tap b4m-oss/tap
  brew install bwsf@0.17.3

Asset naming changed when the project was renamed from bwenv → bwsf:
  - v0.11.1+ (and later RCs) ship bwsf_* archives / bwsf binary
  - earlier releases ship bwenv_* archives / bwenv binary (installed as bwsf)
"""

from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
import urllib.request
from pathlib import Path


REPO = "b4moss/bwsf"
HOMEPAGE = "https://github.com/b4moss/bwsf"
DOWNLOAD_BASE = f"https://github.com/{REPO}/releases/download"

STABLE_TAG_RE = re.compile(r"^v(\d+)\.(\d+)\.(\d+)$")
FORMULA_NAME_RE = re.compile(r"^bwsf@(.+)\.rb$")


def run_gh(*args: str) -> str:
    result = subprocess.run(
        ["gh", *args],
        check=True,
        capture_output=True,
        text=True,
    )
    return result.stdout


def formula_class_name(formula_name: str) -> str:
    """Match Homebrew Formulary.class_s / GoReleaser formulaNameFor."""
    class_name = formula_name[:1].upper() + formula_name[1:]
    class_name = re.sub(
        r"[-_.\s]([a-zA-Z0-9])",
        lambda m: m.group(1).upper(),
        class_name,
    )
    class_name = class_name.replace("+", "x")
    class_name = re.sub(r"(.)@(\d)", r"\1AT\2", class_name)
    return class_name


def parse_checksums(text: str) -> dict[str, str]:
    checksums: dict[str, str] = {}
    for line in text.splitlines():
        line = line.strip()
        if not line:
            continue
        parts = line.split()
        if len(parts) < 2:
            continue
        digest, name = parts[0], parts[-1]
        # strip leading ./ if present
        name = name.lstrip("./")
        checksums[name] = digest
    return checksums


def fetch_checksums(tag: str) -> dict[str, str]:
    url = f"{DOWNLOAD_BASE}/{tag}/checksums.txt"
    try:
        with urllib.request.urlopen(url) as resp:
            return parse_checksums(resp.read().decode())
    except Exception:
        # Fall back to listing release assets and skipping when no checksums
        return {}


def parse_stable_version(tag: str) -> tuple[int, int, int] | None:
    """Return (major, minor, patch) for stable tags only."""
    match = STABLE_TAG_RE.fullmatch(tag)
    if not match:
        return None
    return int(match.group(1)), int(match.group(2)), int(match.group(3))


def version_string(version: tuple[int, int, int]) -> str:
    return f"{version[0]}.{version[1]}.{version[2]}"


def select_retained_versions(
    versions: list[tuple[int, int, int]],
) -> set[tuple[int, int, int]]:
    """Apply the Homebrew versioned-formula retention policy."""
    if not versions:
        return set()

    unique = sorted(set(versions))
    current = unique[-1]
    current_minor = (current[0], current[1])

    retained = {v for v in unique if (v[0], v[1]) == current_minor}

    previous_minor_versions = [v for v in unique if (v[0], v[1]) < current_minor]
    if previous_minor_versions:
        prev_minor = max((v[0], v[1]) for v in previous_minor_versions)
        prev_patches = [v for v in previous_minor_versions if (v[0], v[1]) == prev_minor]
        retained.add(max(prev_patches))

    return retained


def list_releases() -> list[dict]:
    raw = run_gh(
        "release",
        "list",
        "--repo",
        REPO,
        "--limit",
        "200",
        "--json",
        "tagName,isDraft,isPrerelease",
    )
    releases = json.loads(raw)
    return [r for r in releases if not r.get("isDraft") and not r.get("isPrerelease")]


def list_assets(tag: str) -> list[str]:
    raw = run_gh(
        "release",
        "view",
        tag,
        "--repo",
        REPO,
        "--json",
        "assets",
    )
    assets = json.loads(raw).get("assets", [])
    return [a["name"] for a in assets]


PLATFORM_SPECS = [
    # (asset_suffix, os_block, arch_guard)
    ("darwin_amd64.tar.gz", "macos", "intel"),
    ("darwin_arm64.tar.gz", "macos", "arm"),
    ("linux_amd64.tar.gz", "linux", "intel"),
    ("linux_arm64.tar.gz", "linux", "arm"),
]


def detect_prefix(asset_names: set[str]) -> str | None:
    if any(n.startswith("bwsf_") for n in asset_names):
        return "bwsf"
    if any(n.startswith("bwenv_") for n in asset_names):
        return "bwenv"
    return None


def render_install(prefix: str, indent: str) -> str:
    if prefix == "bwsf":
        return f'{indent}bin.install "bwsf"'
    # Legacy bwenv binary — expose as bwsf for a consistent command name
    return f'{indent}bin.install "bwenv" => "bwsf"'


def render_macos_pkg(url: str, sha: str, prefix: str, arch: str, base_indent: int = 2) -> str:
    guard = "Hardware::CPU.intel?" if arch == "intel" else "Hardware::CPU.arm?"
    pad = " " * base_indent
    inner = " " * (base_indent + 2)
    install = render_install(prefix, " " * (base_indent + 4))
    return f"""{pad}if {guard}
{inner}url "{url}"
{inner}sha256 "{sha}"

{inner}define_method(:install) do
{install}
{inner}end
{pad}end"""


def render_linux_pkg(url: str, sha: str, prefix: str, arch: str, base_indent: int = 2) -> str:
    if arch == "intel":
        guard = "Hardware::CPU.intel? && Hardware::CPU.is_64_bit?"
    else:
        guard = "Hardware::CPU.arm? && Hardware::CPU.is_64_bit?"
    pad = " " * base_indent
    inner = " " * (base_indent + 2)
    install = render_install(prefix, " " * (base_indent + 4))
    return f"""{pad}if {guard}
{inner}url "{url}"
{inner}sha256 "{sha}"
{inner}define_method(:install) do
{install}
{inner}end
{pad}end"""


def generate_formula(tag: str, version: str, checksums: dict[str, str], assets: list[str]) -> str | None:
    asset_set = set(assets)
    prefix = detect_prefix(asset_set)
    if not prefix:
        print(f"skip {tag}: no bwsf_/bwenv_ archives", file=sys.stderr)
        return None

    macos_meta: list[tuple[str, str, str]] = []
    linux_meta: list[tuple[str, str, str]] = []

    for suffix, os_name, arch in PLATFORM_SPECS:
        asset = f"{prefix}_{suffix}"
        if asset not in asset_set:
            continue
        sha = checksums.get(asset)
        if not sha:
            print(f"skip {tag}/{asset}: missing checksum", file=sys.stderr)
            continue
        url = f"{DOWNLOAD_BASE}/{tag}/{asset}"
        if os_name == "macos":
            macos_meta.append((url, sha, arch))
        else:
            linux_meta.append((url, sha, arch))

    if not macos_meta and not linux_meta:
        print(f"skip {tag}: no usable platform archives", file=sys.stderr)
        return None

    formula_name = f"bwsf@{version}"
    class_name = formula_class_name(formula_name)

    nested = bool(macos_meta) and bool(linux_meta)
    pkg_indent = 4 if nested else 2

    macos_pkgs = [
        render_macos_pkg(url, sha, prefix, arch, base_indent=pkg_indent)
        for url, sha, arch in macos_meta
    ]
    linux_pkgs = [
        render_linux_pkg(url, sha, prefix, arch, base_indent=pkg_indent)
        for url, sha, arch in linux_meta
    ]

    body_parts: list[str] = []
    if nested:
        body_parts.append("  on_macos do")
        body_parts.append("\n".join(macos_pkgs))
        body_parts.append("  end")
        body_parts.append("")
        body_parts.append("  on_linux do")
        body_parts.append("\n".join(linux_pkgs))
        body_parts.append("  end")
    elif macos_pkgs:
        body_parts.append("  depends_on :macos")
        body_parts.append("")
        body_parts.append("\n".join(macos_pkgs))
    else:
        body_parts.append("  depends_on :linux")
        body_parts.append("")
        body_parts.append("\n".join(linux_pkgs))

    body = "\n".join(body_parts)

    return f"""# typed: false
# frozen_string_literal: true

# Generated by scripts/sync-homebrew-versioned-formulas.py — DO NOT EDIT.
class {class_name} < Formula
  desc "CLI tool to manage .env files with Bitwarden"
  homepage "{HOMEPAGE}"
  version "{version}"
  license "MIT"
  keg_only :versioned_formula

{body}

  test do
    system "#{{bin}}/bwsf", "--help"
  end
end
"""


def prune_unretained_formulas(
    tap_dir: Path,
    retained_versions: set[str],
    *,
    dry_run: bool,
) -> int:
    """Delete bwsf@*.rb formulas whose version is outside the retain set."""
    deleted = 0
    for path in sorted(tap_dir.glob("bwsf@*.rb")):
        match = FORMULA_NAME_RE.fullmatch(path.name)
        if not match:
            continue
        version = match.group(1)
        if version in retained_versions:
            continue
        if dry_run:
            print(f"would delete {path}")
        else:
            path.unlink()
            print(f"deleted {path}")
        deleted += 1
    return deleted


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--tap-dir",
        type=Path,
        required=True,
        help="Path to a checkout of the Homebrew tap repository",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Print formulas that would be written/deleted without changing files",
    )
    args = parser.parse_args()

    tap_dir: Path = args.tap_dir
    if not tap_dir.is_dir():
        print(f"error: tap dir not found: {tap_dir}", file=sys.stderr)
        return 1

    stable_releases: list[tuple[tuple[int, int, int], str]] = []
    for release in list_releases():
        tag = release["tagName"]
        version = parse_stable_version(tag)
        if version is None:
            print(f"skip {tag}: not a stable vMAJOR.MINOR.PATCH tag", file=sys.stderr)
            continue
        stable_releases.append((version, tag))

    retained = select_retained_versions([v for v, _ in stable_releases])
    retained_strings = {version_string(v) for v in retained}
    print(
        "retain: "
        + (", ".join(f"bwsf@{v}" for v in sorted(retained_strings)) or "(none)"),
        file=sys.stderr,
    )

    written = 0
    for version, tag in sorted(stable_releases):
        if version not in retained:
            continue
        version_s = version_string(version)
        assets = list_assets(tag)
        checksums = fetch_checksums(tag)
        if not checksums:
            print(f"skip {tag}: checksums.txt unavailable", file=sys.stderr)
            continue

        content = generate_formula(tag, version_s, checksums, assets)
        if content is None:
            continue

        filename = f"bwsf@{version_s}.rb"
        dest = tap_dir / filename
        if args.dry_run:
            print(f"would write {dest}")
        else:
            dest.write_text(content, encoding="utf-8")
            print(f"wrote {dest}")
        written += 1

    deleted = prune_unretained_formulas(
        tap_dir,
        retained_strings,
        dry_run=args.dry_run,
    )

    print(f"done: {written} versioned formula(s) kept/written, {deleted} pruned")
    return 0 if written or deleted else 1


if __name__ == "__main__":
    raise SystemExit(main())
