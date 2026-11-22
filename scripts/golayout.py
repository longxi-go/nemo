"""Delegate formatting and source notices to the declared golayout dependency."""

import json
import os
from pathlib import Path
import subprocess
import sys

from go_module import read_module
from go_environment import verification_environment

OPERATIONS = {
    "setup": ("format.sh", "setup"),
    "write": ("format.sh", "write"),
    "check": ("format.sh", "check"),
    "test": ("format_test.sh", None),
    "headers-write": ("license_headers.sh", "write"),
    "headers-check": ("license_headers.sh", "check"),
}


def run(root: Path, operation: str) -> int:
    """Resolve a checksum-verified module without relying on sibling checkouts."""
    go = os.environ.get("GO", "go")
    env = verification_environment()
    # Download metadata may write go.sum; require the complete tool graph first.
    verified = subprocess.run(
        [go, "-C", str(root), "list", "-mod=readonly", "-deps",
         "github.com/longxi-go/golayout/cmd/golayout"],
        env=env, capture_output=True, text=True, check=False,
    )
    if verified.returncode:
        raise ValueError("Cannot verify pinned golayout dependencies without changes:\n" + verified.stderr)
    download = subprocess.run(
        [go, "-C", str(root), "mod", "download", "-json", "github.com/longxi-go/golayout"],
        env=env, capture_output=True, text=True, check=False,
    )
    if download.returncode:
        raise ValueError("Cannot resolve pinned golayout module:\n" + download.stderr + download.stdout)
    try:
        module = json.loads(download.stdout)
        directory = module.get("Dir") if isinstance(module, dict) else None
        if not isinstance(directory, str) or not directory or not Path(directory).is_absolute():
            raise ValueError("expected an absolute module directory")
    except ValueError as error:
        raise ValueError(f"Invalid golayout download metadata: {error}") from error
    script_name, mode = OPERATIONS[operation]
    script = Path(directory) / "scripts" / script_name
    if not script.is_file():
        raise ValueError(f"Pinned golayout script is missing: {script}")
    language = read_module(root / "go.mod").get("Go")
    if not language:
        raise ValueError("A go directive is required in go.mod")
    command = ["bash", str(script)]
    if mode is not None:
        command.append(mode)
    command.extend([str(root), str(root), "go" + language])
    return subprocess.run(command, env=env, check=False).returncode


def main() -> None:
    if len(sys.argv) != 2 or sys.argv[1] not in OPERATIONS:
        raise SystemExit("expected a golayout operation: " + ", ".join(OPERATIONS))
    try:
        result = run(Path(__file__).resolve().parents[1], sys.argv[1])
    except (ValueError, OSError, subprocess.SubprocessError) as error:
        raise SystemExit(str(error)) from error
    raise SystemExit(result)


if __name__ == "__main__":
    main()
