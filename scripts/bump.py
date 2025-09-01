from rich.prompt import Prompt
import pathlib
import subprocess
from contextlib import chdir
import json
import re

from .lib.version import Version, write_version, read_current_version

VERSION_FILE = "version.txt"


def main():
    current_version = read_current_version()
    default_version = current_version.next_patch()
    version_str = Prompt.ask(
        "[bold]Enter the new version[/bold]", default=str(default_version)
    )
    version = Version.from_string(version_str)
    write_version(version)

    # Run go generate
    subprocess.run(["go", "generate", "./src"], check=True)
    print(f"Bumped version to {version}")

    # Update versions in package.json files
    package_dir = pathlib.Path("packages")
    for package in package_dir.iterdir():
        if not package.is_dir():
            continue
        with chdir(package):
            version_query = f"version={version}"
            subprocess.run(["pnpm", "pkg", "set", version_query], check=True)
            print(f"Updated package.json in {package} to version {version}")

    # Update the root package.json version
    version_query = f"version={version}"
    subprocess.run(["pnpm", "pkg", "set", version_query], check=True)

    # Update optionalDependencies versions in the root package.json
    with open("package.json", "r") as f:
        package_json = json.load(f)
    optional_dependencies = package_json.get("optionalDependencies", {})
    package_names = [
        pkg for pkg in optional_dependencies.keys() if is_valid_optional_dependency(pkg)
    ]
    for pkg in package_names:
        dep_query = f"optionalDependencies.{pkg}={version}"
        subprocess.run(["pnpm", "pkg", "set", dep_query], check=True)
    print(f"Updated root package.json to version {version}")


def is_valid_optional_dependency(name: str) -> bool:
    pattern = r"@dodo-doc/cli-[a-zA-Z0-9_-]+"
    return re.match(pattern, name) is not None


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\nOperation cancelled.")
