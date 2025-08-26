---
title: Release Flow 
created_at: 2025-02-24T14:28:26+09:00
updated_at: 2025-02-24T14:28:26+09:00
---

# How To Release a CLI
To create a new PR including following changes, please run the following command.
After this PR merged, CI will automatically build a new version and publish it to GitHub.

```bash
uv run scripts/bump.py
```


