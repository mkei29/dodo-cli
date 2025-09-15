---
title: "Release Flow"
path: "for_contributors_release_flow"
description: ""
created_at: "2025-09-15T19:27:45+09:00"
updated_at: "2025-09-15T19:27:45+09:00"
---

:::message alert
Releasing the new version of dodo-cli is only allowed for the administrator.
:::

# How To Release a CLI
To create a new PR including following changes, please run the following command.
After this PR merged, CI will automatically build a new version and publish it to GitHub.

```bash
uv run -m scripts.bump
```


