from rich.prompt import Prompt
import datetime
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

    # Append release draft to release note
    date_str = datetime.date.today().strftime("%Y/%m/%d")
    append_release_note(
        version,
        date_str,
        pathlib.Path("docs/draft/release_draft.md"),
        pathlib.Path("docs/release_note.md"),
        f'To install this update, run `npm install -g "@dodo-doc/cli@{version}"`.',
    )
    append_release_note(
        version,
        date_str,
        pathlib.Path("docs/draft/release_draft.ja.md"),
        pathlib.Path("docs/release_note.ja.md"),
        f'この変更をインストールするには`npm install -g "@dodo-doc/cli@{version}"`を実行してください。',
    )


def append_release_note(
    version: Version,
    date_str: str,
    release_draft_path: pathlib.Path,
    release_note_path: pathlib.Path,
    install_note: str,
):
    if not release_draft_path.exists():
        print(f"No {release_draft_path} found, skipping release note update.")
        return

    draft_content = release_draft_path.read_text().strip()
    if not draft_content:
        print(f"{release_draft_path} is empty, skipping release note update.")
        return

    lines = draft_content.splitlines()
    bullets = []
    for line in lines:
        line = line.strip()
        if not line:
            continue
        if not line.startswith("* "):
            line = f"* {line}"
        bullets.append(line)

    new_section = (
        f"\n# {date_str} - version {version}\n"
        + "\n".join(bullets)
        + f"\n* {install_note}\n"
    )

    content = release_note_path.read_text()
    note_lines = content.split("\n")

    frontmatter_end = -1
    in_frontmatter = False
    for i, note_line in enumerate(note_lines):
        if i == 0 and note_line.strip() == "---":
            in_frontmatter = True
            continue
        if in_frontmatter and note_line.strip() == "---":
            frontmatter_end = i
            break

    if frontmatter_end == -1:
        new_content = new_section + content
    else:
        before = "\n".join(note_lines[: frontmatter_end + 1])
        after = "\n".join(note_lines[frontmatter_end + 1 :])
        new_content = before + new_section + after

    release_note_path.write_text(new_content)
    release_draft_path.write_text("")
    print(f"Updated {release_note_path} with new entry for version {version}")


def is_valid_optional_dependency(name: str) -> bool:
    pattern = r"@dodo-doc/cli-[a-zA-Z0-9_-]+"
    return re.match(pattern, name) is not None


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\nOperation cancelled.")
