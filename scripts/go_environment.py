"""Require the checked-in dependency graph for repository verification."""

import os
import subprocess


def verification_environment() -> dict:
    """Reject hidden module flags while retaining configured network settings."""
    if os.environ.get("GOFLAGS"):
        raise ValueError("GOFLAGS must be empty for repository verification")
    env = dict(os.environ, GOWORK="off", GOFLAGS="", GOTOOLCHAIN="local")
    result = subprocess.run(
        [env.get("GO", "go"), "env", "GOFLAGS"],
        env=env, capture_output=True, text=True, check=True, timeout=30,
    )
    if result.stdout.strip():
        raise ValueError("GOFLAGS in persistent Go configuration must be empty for repository verification")
    return env


if __name__ == "__main__":
    try:
        verification_environment()
    except (ValueError, OSError, subprocess.SubprocessError) as error:
        raise SystemExit(str(error)) from error
