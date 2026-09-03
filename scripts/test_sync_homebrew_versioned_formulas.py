#!/usr/bin/env python3
"""Unit tests for Homebrew versioned-formula retention (#171)."""

from __future__ import annotations

import importlib.util
import tempfile
import unittest
from pathlib import Path

_SPEC = importlib.util.spec_from_file_location(
    "sync_homebrew_versioned_formulas",
    Path(__file__).with_name("sync-homebrew-versioned-formulas.py"),
)
assert _SPEC and _SPEC.loader
_mod = importlib.util.module_from_spec(_SPEC)
_SPEC.loader.exec_module(_mod)

parse_stable_version = _mod.parse_stable_version
prune_unretained_formulas = _mod.prune_unretained_formulas
select_retained_versions = _mod.select_retained_versions
version_string = _mod.version_string


class RetentionTests(unittest.TestCase):
    def test_parse_stable_version(self) -> None:
        self.assertEqual(parse_stable_version("v0.17.3"), (0, 17, 3))
        self.assertIsNone(parse_stable_version("v0.12.0-rc.1"))
        self.assertIsNone(parse_stable_version("site-v0.16.0-doc.1"))
        self.assertIsNone(parse_stable_version("0.17.3"))

    def test_retain_current_minor_all_patches_and_previous_latest(self) -> None:
        versions = [
            (0, 15, 0),
            (0, 16, 0),
            (0, 16, 1),
            (0, 17, 0),
            (0, 17, 1),
            (0, 17, 2),
            (0, 17, 3),
        ]
        retained = select_retained_versions(versions)
        self.assertEqual(
            {version_string(v) for v in retained},
            {"0.17.0", "0.17.1", "0.17.2", "0.17.3", "0.16.1"},
        )

    def test_previous_minor_without_patches(self) -> None:
        versions = [(0, 16, 0), (0, 17, 0), (0, 17, 1)]
        retained = select_retained_versions(versions)
        self.assertEqual(
            {version_string(v) for v in retained},
            {"0.17.0", "0.17.1", "0.16.0"},
        )

    def test_major_bump_keeps_prior_major_latest_minor_latest_patch(self) -> None:
        versions = [
            (0, 16, 0),
            (0, 17, 2),
            (0, 17, 3),
            (1, 0, 0),
        ]
        retained = select_retained_versions(versions)
        self.assertEqual(
            {version_string(v) for v in retained},
            {"1.0.0", "0.17.3"},
        )

    def test_only_one_minor(self) -> None:
        versions = [(0, 18, 0), (0, 18, 1)]
        retained = select_retained_versions(versions)
        self.assertEqual(
            {version_string(v) for v in retained},
            {"0.18.0", "0.18.1"},
        )

    def test_prune_deletes_unretained(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            tap = Path(tmp)
            (tap / "bwsf@0.17.3.rb").write_text("keep\n", encoding="utf-8")
            (tap / "bwsf@0.15.0.rb").write_text("drop\n", encoding="utf-8")
            (tap / "bwsf.rb").write_text("latest\n", encoding="utf-8")
            deleted = prune_unretained_formulas(tap, {"0.17.3"}, dry_run=False)
            self.assertEqual(deleted, 1)
            self.assertTrue((tap / "bwsf@0.17.3.rb").exists())
            self.assertFalse((tap / "bwsf@0.15.0.rb").exists())
            self.assertTrue((tap / "bwsf.rb").exists())


if __name__ == "__main__":
    unittest.main()
