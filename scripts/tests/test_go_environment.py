import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from go_environment import verification_environment


class GoEnvironmentTests(unittest.TestCase):
    def test_rejects_process_module_override(self):
        with patch.dict(os.environ, {"GOFLAGS": "-modfile=/tmp/alternate.mod"}):
            with self.assertRaisesRegex(ValueError, "GOFLAGS"):
                verification_environment()

    def test_rejects_persistent_module_override_using_real_go(self):
        with tempfile.TemporaryDirectory() as directory:
            settings = Path(directory) / "goenv"
            settings.write_text("GOFLAGS=-modfile=/tmp/alternate.mod\n", encoding="utf-8")
            with patch.dict(os.environ, {"GOENV": str(settings), "GOFLAGS": ""}):
                with self.assertRaisesRegex(ValueError, "persistent"):
                    verification_environment()

    def test_isolates_workspace_and_toolchain_without_changing_process(self):
        with patch.dict(os.environ, {"GOENV": "off", "GOFLAGS": "", "GOWORK": "/tmp/other.work", "GOTOOLCHAIN": "auto"}):
            env = verification_environment()
            self.assertEqual("off", env["GOWORK"])
            self.assertEqual("local", env["GOTOOLCHAIN"])
            self.assertEqual("/tmp/other.work", os.environ["GOWORK"])
            result = subprocess.run(
                [env.get("GO", "go"), "env", "GOWORK", "GOFLAGS"],
                env=env, capture_output=True, text=True, check=True,
            )
            self.assertEqual(["off", ""], result.stdout.splitlines())


if __name__ == "__main__":
    unittest.main()
