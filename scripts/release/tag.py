#!/usr/bin/env python3

from __future__ import annotations

import subprocess
import sys


def run_cmd(command: str, *args: str) -> str:
    proc = subprocess.run(
        [command, *args],
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        text=True,
        check=False,
    )

    output = proc.stdout.strip()
    if proc.returncode != 0:
        if "No names found" in output:
            return ""
        raise RuntimeError(
            f"failed to run command '{command} {' '.join(args)}':\n{output}"
        )

    return output


def execute_step(description: str, command: str, *args: str) -> None:
    print(description)
    subprocess.run([command, *args], check=True)


def main() -> int:
    run_cmd("git", "fetch", "origin")

    status = run_cmd("git", "status", "--porcelain")
    if status:
        raise RuntimeError("Uncommitted changes found, please commit or stash them first.")

    local = run_cmd("git", "rev-parse", "@")
    remote = run_cmd("git", "rev-parse", "@{u}")

    if local != remote:
        print("Local branch is not up to date with remote, pulling...")
        run_cmd("git", "pull")

    last_tag = run_cmd("git", "describe", "--tags", "--abbrev=0")
    if last_tag:
        print(f"Latest tag: {last_tag}")
    else:
        print("No tags found.")

    new_tag = input("Enter new tag: ").strip()
    if not new_tag:
        raise RuntimeError("No tag entered, aborting.")

    execute_step(f"Tagging {new_tag}...", "git", "tag", new_tag)
    execute_step(f"Pushing tag {new_tag}...", "git", "push", "origin", new_tag)
    print(f"Successfully tagged and pushed {new_tag}.")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as exc:  # noqa: BLE001
        print(exc, file=sys.stderr)
        raise SystemExit(1)
