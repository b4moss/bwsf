#!/usr/bin/env python3
"""Generate versioned Homebrew formulas for every GitHub release of bwsf.

Writes bwsf@<version>.rb files into a tap checkout so past releases remain
installable via:

  brew tap b4m-oss/tap
  brew install bwsf@0.15.0

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
HOMEPAGE = "https://github.com/b4m-oss/bwsf"
DOWNLOAD_BASE = f"https://github.com/{REPO}/releases/download"


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
    return [r for r in releases if not r.get("isDraft")]


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
        help="Print formulas that would be written without writing files",
    )
    args = parser.parse_args()

    tap_dir: Path = args.tap_dir
    if not tap_dir.is_dir():
        print(f"error: tap dir not found: {tap_dir}", file=sys.stderr)
        return 1

    written = 0
    for release in list_releases():
        tag = release["tagName"]
        if not tag.startswith("v"):
            print(f"skip {tag}: unexpected tag format", file=sys.stderr)
            continue
        version = tag[1:]
        assets = list_assets(tag)
        checksums = fetch_checksums(tag)
        if not checksums:
            # Try to continue only if we somehow have digests elsewhere — skip safely
            print(f"skip {tag}: checksums.txt unavailable", file=sys.stderr)
            continue

        content = generate_formula(tag, version, checksums, assets)
        if content is None:
            continue

        filename = f"bwsf@{version}.rb"
        dest = tap_dir / filename
        if args.dry_run:
            print(f"would write {dest}")
        else:
            dest.write_text(content, encoding="utf-8")
            print(f"wrote {dest}")
        written += 1

    print(f"done: {written} versioned formula(s)")
    return 0 if written else 1


if __name__ == "__main__":
    raise SystemExit(main())
