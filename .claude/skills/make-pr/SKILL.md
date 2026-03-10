---
name: make-pr
description: Create a release PR with auto-generated changelog from current branch to main
argument-hint: [base-branch]
disable-model-invocation: true
---

You are a release PR creator that automatically generates a pull request with a comprehensive changelog.

## Release PR Workflow:

1. **Parse Arguments**:
   - `base-branch` (optional): The target branch for the PR (default: `main`)
   - If no arguments provided, use `main` as the base branch

2. **Validate Current State**:
   ```bash
   git status
   git branch --show-current
   ```
   - Ensure you're not on the base branch (main)
   - Check for uncommitted changes and warn the user if any exist
   - Get the current branch name for the PR

3. **Get Commit Range**:
   ```bash
   git fetch origin <base-branch>
   git merge-base origin/<base-branch> HEAD
   ```
   - Find the common ancestor with the base branch
   - This determines the starting point for the changelog

4. **Fetch Commit History**:
   ```bash
   git log <merge-base>..HEAD --oneline --no-merges
   git log <merge-base>..HEAD --format="%s" --no-merges
   ```
   - Get list of all commits in the range
   - Exclude merge commits to avoid duplication

5. **Analyze Changes**:
   ```bash
   git diff origin/<base-branch>...HEAD --stat
   ```
   - Categorize commits by type (feat, fix, docs, refactor, etc.)
   - Identify affected services/components
   - Note breaking changes (BREAKING CHANGE in commit messages)
   - Track file changes and their scope

6. **Generate PR Description**:
   Create a well-formatted PR description with the following sections:

   ```markdown
   ## Summary
   <1-3 sentence summary of the overall changes>

   ## Changes

   ### Features
   - List new features (commits starting with "feat:")

   ### Bug Fixes
   - List bug fixes (commits starting with "fix:")

   ### Documentation
   - List documentation changes (commits starting with "docs:")

   ### Refactoring
   - List refactoring (commits starting with "refactor:")

   ### Other Changes
   - List other commits that don't fit above categories

   ## Breaking Changes
   - Highlight any breaking changes mentioned in commit messages
   - If none, omit this section

   ## Affected Services
   - List services/components that were modified:
     - Frontend
     - Backend
     - Contents Frontend
     - Contents Backend
     - Core
     - etc.

   ## Statistics
   - Commits: <count>
   - Files changed: <count>
   - Insertions: <count>
   - Deletions: <count>

   ## Test Plan
   - [ ] All existing tests pass
   - [ ] New features have been tested
   - [ ] No regressions observed

   ---
   Generated with [Claude Code](https://claude.ai/code)
   ```

7. **Determine PR Title**:
   - Analyze the commits to determine the appropriate title
   - If mostly features: `feat: <main feature description>`
   - If mostly fixes: `fix: <main fix description>`
   - If mixed: `release: <brief description of changes>`
   - Keep the title under 70 characters

8. **Push Branch if Needed**:
   ```bash
   git push -u origin <current-branch>
   ```
   - Ensure the branch is pushed to remote before creating PR

9. **Create the Pull Request**:
   ```bash
   gh pr create --base <base-branch> --title "<pr-title>" --body "<generated-description>"
   ```
   - Use the generated title and description
   - Target the specified base branch

10. **Report Result**:
    - Show the PR URL to the user
    - Summarize the key changes included in the PR
    - Mention any action items or review considerations

**Key Features**:
- Automatic changelog generation from commit history
- Smart PR title based on commit types
- Service-based change grouping for this monorepo
- Statistics on code changes
- Professional formatting for release notes
- Support for custom base branches

**Commit Format Recognition**:
- `feat: add new feature` → Features section
- `fix: resolve bug` → Bug Fixes section
- `docs: update readme` → Documentation section
- `refactor: improve code structure` → Refactoring section
- `perf: optimize query` → Performance section (grouped with Other)
- `feat!: breaking change` or `BREAKING CHANGE:` → Breaking Changes section

**Service Detection**:
Based on file paths, detect affected services:
- `services/frontend/` → Frontend
- `services/contents_frontend/` → Contents Frontend
- `services/backend/` → Backend
- `services/contents_backend/` → Contents Backend
- `services/core/` → Core
- `services/sync_job/` → Sync Job
- `services/ui_component/` → UI Component
- `k8s/` → Infrastructure
- `database/` → Database
- `openapi/` → API Specs

**Usage Examples**:
- `/make-pr` - Create PR from current branch to main
- `/make-pr develop` - Create PR from current branch to develop
- `/make-pr release/v1.0` - Create PR to a release branch

**Error Handling**:
- If on base branch, prompt user to switch to a feature branch
- If no commits to include, inform user and abort
- If branch not pushed, push it automatically
- If PR already exists, show the existing PR URL

**Tips**:
- Ensure all changes are committed before running
- Use conventional commit format for better changelog generation
- Review the generated PR description before finalizing
- Add reviewers manually if needed: `gh pr edit --add-reviewer username`
