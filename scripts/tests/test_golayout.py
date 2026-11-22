import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
import golayout


class GolayoutTests(unittest.TestCase):
    def setUp(self):
        temporary = tempfile.TemporaryDirectory(prefix="nemo scripts ")
        self.addCleanup(temporary.cleanup)
        self.root = Path(temporary.name)
        guard = patch.object(golayout, "verification_environment", return_value={"GOWORK": "off"})
        guard.start()
        self.addCleanup(guard.stop)
        self.directory = self.root / "tool cache"
        (self.directory / "scripts").mkdir(parents=True)
        for script, _ in golayout.OPERATIONS.values():
            (self.directory / "scripts" / script).touch()

    def test_missing_checksums_do_not_change_module_files(self):
        # This fixed release is a missing-checksum fixture, not the repository pin.
        module = """module example.test/readonly

go 1.27.0

require github.com/longxi-go/golayout v0.1.4

tool github.com/longxi-go/golayout/cmd/golayout
"""
        (self.root / "go.mod").write_text(module)
        subprocess.run(["git", "init", "--quiet", str(self.root)], check=True)
        env = dict(os.environ, GOWORK="off", GOTOOLCHAIN="local", GOFLAGS="", GOENV="off")
        with patch.object(golayout, "verification_environment", return_value=env):
            with self.assertRaisesRegex(ValueError, "missing go.sum"):
                golayout.run(self.root, "check")
        self.assertEqual(module, (self.root / "go.mod").read_text())
        self.assertFalse((self.root / "go.sum").exists())
        partial_sum = "github.com/longxi-go/golayout v0.1.4/go.mod h1:YB1IaFZ+0UtB3XuCHfSFci8aY7yJyDXFhexoqgUzFQc=\n"
        (self.root / "go.sum").write_text(partial_sum)
        with patch.object(golayout, "verification_environment", return_value=env):
            with self.assertRaisesRegex(ValueError, "missing go.sum"):
                golayout.run(self.root, "setup")
        self.assertEqual(module, (self.root / "go.mod").read_text())
        self.assertEqual(partial_sum, (self.root / "go.sum").read_text())

    def test_delegates_each_operation_and_module_language(self):
        for operation, (script, mode) in golayout.OPERATIONS.items():
            result = subprocess.CompletedProcess([], 0, json.dumps({"Dir": str(self.directory)}), "")
            with self.subTest(operation=operation), patch.object(golayout, "read_module", return_value={"Go": "1.28.1"}):
                with patch.object(golayout.subprocess, "run", side_effect=[subprocess.CompletedProcess([], 0), result, subprocess.CompletedProcess([], 7)]) as execute:
                    self.assertEqual(7, golayout.run(self.root, operation))
            self.assertEqual(
                [os.environ.get("GO", "go"), "-C", str(self.root), "list", "-mod=readonly", "-deps",
                 "github.com/longxi-go/golayout/cmd/golayout"],
                execute.call_args_list[0].args[0],
            )
            call = execute.call_args_list[-1]
            expected = ["bash", str(self.directory / "scripts" / script)]
            if mode is not None:
                expected.append(mode)
            expected.extend([str(self.root), str(self.root), "go1.28.1"])
            self.assertEqual(expected, call.args[0])
            self.assertEqual("off", call.kwargs["env"]["GOWORK"])

    def test_rejects_failed_download_and_invalid_metadata(self):
        cases = (
            subprocess.CompletedProcess([], 1, "", "download failed"),
            subprocess.CompletedProcess([], 0, "not json", ""),
            subprocess.CompletedProcess([], 0, "[]", ""),
            subprocess.CompletedProcess([], 0, '{"Dir":"relative"}', ""),
            subprocess.CompletedProcess([], 0, '{"Dir":null}', ""),
        )
        for result in cases:
            with self.subTest(result=result), patch.object(golayout.subprocess, "run", side_effect=[subprocess.CompletedProcess([], 0), result]):
                with self.assertRaises(ValueError):
                    golayout.run(self.root, "check")

    def test_requires_shared_script_and_language_directive(self):
        result = subprocess.CompletedProcess([], 0, json.dumps({"Dir": str(self.directory)}), "")
        with patch.object(golayout.subprocess, "run", return_value=result), patch.object(golayout, "read_module", return_value={}):
            with self.assertRaisesRegex(ValueError, "go directive"):
                golayout.run(self.root, "check")
        (self.directory / "scripts/format.sh").unlink()
        with patch.object(golayout.subprocess, "run", return_value=result):
            with self.assertRaisesRegex(ValueError, "script is missing"):
                golayout.run(self.root, "check")


if __name__ == "__main__":
    unittest.main()
