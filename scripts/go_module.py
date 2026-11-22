"""Read module declarations with Go's parser without resolving dependencies."""

import json
import os
from pathlib import Path
import subprocess

def read_module(path: Path) -> dict:
    """Parse only the supplied file, independent of workspace and user flags."""
    env = dict(os.environ, GOWORK="off", GOFLAGS="", GOENV="off", GOTOOLCHAIN="local")
    try:
        result = subprocess.run(
            [env.get("GO", "go"), "mod", "edit", "-json", str(path.resolve())],
            env=env,
            capture_output=True,
            text=True,
            check=True,
            timeout=30,
        )
        data = json.loads(result.stdout)
        if not isinstance(data, dict):
            raise ValueError("expected a module JSON object")
        return data
    except (OSError, subprocess.SubprocessError, ValueError) as error:
        detail = error.stderr.strip() if isinstance(error, subprocess.CalledProcessError) else str(error)
        raise ValueError(f"Cannot read Go module {path}: {detail}") from error
