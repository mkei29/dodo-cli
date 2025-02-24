---
title: Release Flow 
created_at: 2025-02-24T14:28:26+09:00
updated_at: 2025-02-24T14:28:26+09:00
---

# How To Release a CLI
Create a new PR including following changes.

* Bump the `version.txt`. (You can manually edit the file.)
* Bump the version written in `download.sh`
* Update the `release_note.md` in the `docs` directory.

After this PR merged, CI will automatically build a new version and publish it to GitHub.

