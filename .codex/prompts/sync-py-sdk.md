---
description: "Sync generated Python SDK into coze-py and complete PR lifecycle"
---
When asked to sync generated Python SDK output into `coze-py` and complete the PR lifecycle, follow this exact flow.

## Part A: Commit and sync `coze-sdk-gen` repository changes

1. Sync and merge `coze-sdk-gen` branch code when this repo has changes from the task:
   - Run required local gates before commit in `coze-sdk-gen`:
     - formatting
     - lint
     - test
     - build
     - Python diff = 0
     - Go diff = 0
   - Commit with conventional commit style (`feat:`, `fix:`, `refactor:`, etc.).
   - Push current branch to remote.
   - If pushing to `main` is rejected due non-fast-forward, integrate remote first:
     - `git -C /path/to/coze-sdk-gen pull --rebase origin main`
     - then push again.
2. Continue with downstream SDK generation after `coze-sdk-gen` changes are committed/pushed.

## Part B: Generate SDK and complete downstream `coze-py` PR

1. Resolve `<coze-py-path>` first:
   - Use a random temporary path for each run (do not reuse existing local repo paths), for example:
     - `<coze-py-path>=$(mktemp -d /tmp/coze-py-XXXXXX)`
2. Prepare `coze-py` repository:
   - Always clone a fresh repository into `<coze-py-path>` for this run:
     - `git clone https://github.com/coze-dev/coze-py <coze-py-path>`
   - Do not reuse old working trees across runs, to avoid cross-process interference.
3. Ensure git user in `coze-py` (repo-level preferred, fallback to global check) is exactly:
   - `user.name = chyroc`
   - `user.email = chyroc@qq.com`
   - If not set, configure it before branch/commit.
4. Run generation with CI checks against target SDK path:
   - `./scripts/genpy.sh --output-sdk <coze-py-path> --ci-check`
5. Inspect actual `coze-py` git diff and identify changed files/functions.
6. Inspect related `coze-sdk-gen` commits (`git log` + `git show`) and map generator/config changes to SDK output changes.
7. Draft PR title/body from facts above:
   - include SDK behavior change summary
   - include generator commit(s) and what config/generation logic changed
   - include validation command and result summary
   - never include literal `\\n` text in PR body; use real line breaks only
8. Create branch, commit, and push in `coze-py`.
9. Create or update PR.
   - Prefer `gh pr create` / `gh pr edit` when token scopes allow.
   - If `gh` GraphQL fails due org scope restrictions (for example missing `read:org`), use GitHub REST API as fallback:
     - `PATCH /repos/{owner}/{repo}/pulls/{number}` for title/body updates
     - `POST /repos/{owner}/{repo}/issues/{number}/labels` for required labels
10. Ensure required PR labels are present (for `coze-py`: one of `feature`, `enhancement`, `fix`, `bugfix`, `bug`, `chore`, `documentation`) so label checks pass.
11. Poll PR checks until completion; resolve failures before merge.
12. Merge PR only after required checks pass and branch protection conditions are satisfied.
13. Report final results:
   - downstream `coze-py` PR URL/status/merge result
   - `coze-sdk-gen` commit/push result (if updated in this task)
