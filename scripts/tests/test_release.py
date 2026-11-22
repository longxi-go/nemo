import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
import release

MODULE = "github.com/longxi-go/nemo"
TAG = "v0.1.0"


class ReleaseTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        self.module = self.root / "go.mod"
        self.module.write_text(
            f'module "{MODULE}"\n\ngo 1.27.0\n\nrequire (\n'
            'github.com/longxi-go/gosr v0.1.0\n'
            'github.com/longxi-go/gosr/tools v0.1.0\n'
            'github.com/longxi-go/golayout v0.1.4\n)\n', encoding="utf-8"
        )
        notes = self.root / "docs/releases" / f"{TAG}.md"
        notes.parent.mkdir(parents=True)
        notes.write_text(f"# nemo {TAG}\n\nConfiguration environments.\n", encoding="utf-8")

    def test_valid_module_and_notes(self):
        self.assertEqual("0.1.0", release.validate_tag(self.root, TAG))

    def test_payload_preserves_notes_and_requests_semantic_latest_policy(self):
        payload = release.release_payload(self.root, TAG)
        self.assertEqual(TAG, payload["tag_name"])
        self.assertEqual(f"nemo {TAG}", payload["name"])
        self.assertEqual("legacy", payload["make_latest"])
        self.assertIs(False, payload["draft"])
        self.assertIs(False, payload["prerelease"])
        self.assertEqual(
            (self.root / "docs/releases" / f"{TAG}.md").read_text(), payload["body"]
        )

    def test_rejects_unstable_or_incompatible_tags(self):
        for tag in ("v01.0.0", "v0.1.0-rc.1", "tools/v0.1.0", "v2.0.0", "../v0.1.0"):
            with self.subTest(tag=tag), self.assertRaises(ValueError):
                release.validate_tag(self.root, tag)

    def test_module_identity_and_syntax_use_go_parser(self):
        for content in ('module example.com/other\n\ngo 1.27.0\n', '/* module '+MODULE+' */\n'):
            with self.subTest(content=content):
                self.module.write_text(content, encoding="utf-8")
                with self.assertRaises(ValueError):
                    release.validate_tag(self.root, TAG)

    def test_rejects_replacement_and_unpublished_dependency(self):
        original = self.module.read_text()
        cases = (
            original + f"\nreplace {MODULE} => ../nemo\n",
            original.replace("gosr v0.1.0", "gosr v0.0.0"),
            original.replace("gosr/tools v0.1.0", "gosr/tools v0.1.0-rc.1"),
        )
        for content in cases:
            with self.subTest(content=content):
                self.module.write_text(content, encoding="utf-8")
                with self.assertRaises(ValueError):
                    release.validate_tag(self.root, TAG)

    def test_rejects_missing_or_mismatched_notes(self):
        notes = self.root / "docs/releases" / f"{TAG}.md"
        for content in ("# nemo v0.2.0\n\nBody\n", f"# nemo {TAG}\n"):
            notes.write_text(content, encoding="utf-8")
            with self.assertRaises(ValueError):
                release.validate_tag(self.root, TAG)
        notes.unlink()
        with self.assertRaises(ValueError):
            release.validate_tag(self.root, TAG)

    def github_responses(self, *, policies=None, pages_type="workflow"):
        return {
            "repos/longxi-go/nemo/pages": {"build_type": pages_type},
            "repos/longxi-go/nemo/environments/github-pages": {
                "deployment_branch_policy": {"custom_branch_policies": True, "protected_branches": False}
            },
            "repos/longxi-go/nemo/environments/github-pages/deployment-branch-policies?per_page=100": {
                "total_count": len(policies or []), "branch_policies": policies or []
            },
        }

    def test_preflight_detects_main_only_pages_environment(self):
        responses = self.github_responses(policies=[{"name": "main", "type": "branch"}])
        with patch.object(release, "github_json", side_effect=responses.__getitem__):
            with self.assertRaisesRegex(ValueError, "does not allow tag"):
                release.check_pages("longxi-go/nemo", TAG)

    def test_preflight_accepts_root_tag_policy_and_rejects_branch_lookalike(self):
        for kind, expected in (("tag", True), ("branch", False)):
            responses = self.github_responses(policies=[{"name": "v*", "type": kind}])
            with self.subTest(kind=kind), patch.object(release, "github_json", side_effect=responses.__getitem__):
                if expected:
                    release.check_pages("longxi-go/nemo", TAG)
                else:
                    with self.assertRaises(ValueError):
                        release.check_pages("longxi-go/nemo", TAG)

    def test_pages_requires_actions_and_no_unknown_policy_shape(self):
        responses = self.github_responses(pages_type="legacy")
        with patch.object(release, "github_json", side_effect=responses.__getitem__):
            with self.assertRaisesRegex(ValueError, "workflow"):
                release.check_pages("longxi-go/nemo", TAG)
        responses = self.github_responses()
        responses["repos/longxi-go/nemo/environments/github-pages"] = {"deployment_branch_policy": None}
        with patch.object(release, "github_json", side_effect=responses.__getitem__):
            release.check_pages("longxi-go/nemo", TAG)

    def test_dependency_release_checks_map_nested_module_tag(self):
        calls = []

        def respond(path):
            calls.append(path)
            version = "v0.1.4" if "golayout" in path else "v0.1.0"
            tag = "tools/" + version if "/tags/tools/" in path else version
            return {"tag_name": tag, "draft": False, "prerelease": False}

        with patch.object(release, "github_json", side_effect=respond):
            release.check_dependencies(self.root)
        self.assertIn("repos/longxi-go/gosr/releases/tags/tools/v0.1.0", calls)

    def test_rejects_unpublished_or_unstable_dependency_release(self):
        for data in ({}, {"tag_name": "v0.1.0", "draft": True, "prerelease": False}):
            with self.subTest(data=data), patch.object(release, "github_json", return_value=data):
                with self.assertRaises(ValueError):
                    release.check_dependencies(self.root)

    def test_github_error_and_malformed_response_are_diagnostic(self):
        for result in (
            subprocess.CompletedProcess([], 1, "", "not authenticated"),
            subprocess.CompletedProcess([], 0, "not json", ""),
        ):
            with self.subTest(result=result), patch.object(release.subprocess, "run", return_value=result):
                with self.assertRaises(ValueError):
                    release.github_json("repos/longxi-go/nemo/pages")

    def test_cli_writes_validated_payload_without_github(self):
        target = self.root / "release.json"
        result = subprocess.run(
            [sys.executable, release.__file__, "--root", str(self.root), "--tag", TAG,
             "--github-payload", str(target)],
            capture_output=True, text=True, check=False,
        )
        self.assertEqual(0, result.returncode, result.stderr)
        self.assertEqual("0.1.0", result.stdout.strip())
        self.assertEqual(release.release_payload(self.root, TAG), json.loads(target.read_text()))

    def test_cli_rejects_invalid_tag_without_github_or_file_writes(self):
        script = Path(release.__file__)
        env = dict(os.environ, GOWORK="off")
        result = subprocess.run(
            [sys.executable, str(script), "--root", str(self.root), "--tag", "v00.1.0"],
            env=env, capture_output=True, text=True, check=False,
        )
        self.assertNotEqual(0, result.returncode)
        self.assertIn("stable", result.stderr)


if __name__ == "__main__":
    unittest.main()
