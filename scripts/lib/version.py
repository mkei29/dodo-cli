import pathlib

VERSION_FILE = "version.txt"


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
