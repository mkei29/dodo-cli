from rich.prompt import Prompt
import pathlib
import subprocess
from contextlib import chdir
import json
import re

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


class Version:
    @classmethod
    def from_string(cls, version: str):
        splitted = version.split(".")
        if len(splitted) != 3:
            raise ValueError("Version must be in the format 'major.minor.patch'")
        major, minor, patch = splitted
        if not major.isdigit():
            raise ValueError("Major version cannot be zero")
        if not minor.isdigit():
            raise ValueError("Minor version cannot be zero")
        if not patch.isdigit():
            raise ValueError("Patch version cannot be zero")
        return cls(int(major), int(minor), int(patch))

    def __init__(self, major: int, minor: int, patch: int):
        self.major = major
        self.minor = minor
        self.patch = patch

    def next_patch(self):
        return Version(self.major, self.minor, self.patch + 1)

    def __str__(self):
        return f"{self.major}.{self.minor}.{self.patch}"

    def __repr__(self):
        return f"Version({self.major}, {self.minor}, {self.patch})"


def write_version(version: str) -> Version:
    with open(VERSION_FILE, "w") as f:
        f.write(str(version))


def read_current_version() -> Version:
    repository_root = pathlib.Path(__file__).parent

    version_path = None
    while repository_root != repository_root.parent:
        p = repository_root / VERSION_FILE
        if p.exists():
            version_path = p
        repository_root = repository_root.parent

    with open(version_path, "r") as f:
        current_version = f.read().strip()
    return Version.from_string(current_version)


def is_valid_optional_dependency(name: str) -> bool:
    pattern = r"@dodo-doc/cli-[a-zA-Z0-9_-]+"
    return re.match(pattern, name) is not None


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\nOperation cancelled.")
