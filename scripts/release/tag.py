#!/usr/bin/env python3

from __future__ import annotations

import subprocess
import sys


def run_cmd(command: str, *args: str) -> str:
    try:
        return subprocess.check_output(
            [command, *args], stderr=subprocess.STDOUT, text=True
        ).strip()
    except subprocess.CalledProcessError as exc:
        output = exc.output.strip()
        if "No names found" in output:
            return ""
        raise RuntimeError(
            f"failed to run command '{command} {' '.join(args)}':\n{output}"
        ) from exc


def main() -> int:
    run_cmd("git", "fetch", "origin")

    if run_cmd("git", "status", "--porcelain"):
        raise RuntimeError("Uncommitted changes found, please commit or stash them first.")

    local = run_cmd("git", "rev-parse", "@")

    try:
        if local != run_cmd("git", "rev-parse", "@{upstream}"):
            print("Local branch is not up to date with remote, pulling...")
            run_cmd("git", "pull")
    except RuntimeError:
        print("Note: No upstream branch detected. Skipping remote sync check.")

    last_tag = run_cmd("git", "describe", "--tags", "--abbrev=0")
    if last_tag:
        print(f"Latest tag: {last_tag}")
    else:
        print("No tags found.")

    new_tag = input("Enter new tag: ").strip()
    if not new_tag:
        raise RuntimeError("No tag entered, aborting.")

    print(f"Tagging {new_tag}...")
    subprocess.run(["git", "tag", new_tag], check=True)
    print(f"Pushing tag {new_tag}...")
    subprocess.run(["git", "push", "origin", new_tag], check=True)
    print(f"Successfully tagged and pushed {new_tag}.")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as exc:  # noqa: BLE001
        print(exc, file=sys.stderr)
        raise SystemExit(1)
