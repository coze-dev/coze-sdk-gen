# Scope

This document defines execution rules for `coze-sdk-gen`, targeted at automated coding agents (e.g., Codex/Claude).  
The objective is reproducible, auditable implementation that continuously converges to **zero diff for both Python and Go SDKs**.

# Style Reference (AGENTS.md / CLAUDE.md)

This file follows common `AGENTS.md` / `CLAUDE.md` best practices:

- Define objective, inputs, outputs, and acceptance criteria first.
- Explicitly define allowed vs disallowed behaviors (especially no direct copy from baseline SDKs).
- Constrain workflow before implementation details.
- Require verifiable checks for every change (`lint` / `test` / `build` / `diff`).
- Priority order: this file > temporary verbal preference > default behavior.

# Inputs and Baseline Repositories

- Swagger input: `coze-openapi.yaml`
- Python baseline repository: `https://github.com/coze-dev/coze-py`
  - Local alignment directory: `exist-repo/coze-py`
- Go baseline repository: `https://github.com/coze-dev/coze-go`
  - Local alignment directory: `exist-repo/coze-go`

# Repository Bootstrap

If baseline repositories are missing locally, they must be cloned before generation/alignment:

- `mkdir -p exist-repo`
- `test -d exist-repo/coze-py || git clone https://github.com/coze-dev/coze-py exist-repo/coze-py`
- `test -d exist-repo/coze-go || git clone https://github.com/coze-dev/coze-go exist-repo/coze-go`

# Output Directories and Acceptance Targets

- Python output directory: `exist-repo/coze-py-generated`
- Go output directory: `exist-repo/coze-go-generated`

Final acceptance must satisfy both:

- Diff file count between `exist-repo/coze-py-generated` and `exist-repo/coze-py` is **0**
- Diff file count between `exist-repo/coze-go-generated` and `exist-repo/coze-go` is **0**

# Core Implementation Principles

## 1) Swagger-First (Default)

- Non-special files (default: all files) must be produced via Swagger parsing + generation logic.
- Direct copying/overwriting from `exist-repo/coze-py` or `exist-repo/coze-go` is forbidden.
- Field/API changes should be handled through Swagger + config convergence first.

## 2) Special-File Whitelist (Exception Mechanism)

Only the following files, which are runtime infrastructure without direct OpenAPI equivalents, may be template-based:

- `cozepy/__init__.py`
- `cozepy/auth/__init__.py`
- `cozepy/websockets/**`
- `request.py.tpl`
- `config.py.tpl`
- `util.py.tpl`
- `model.py.tpl`
- `log.py.tpl`
- `exception.py.tpl`
- `version.py.tpl`
- `pyproject.toml.tpl`

Outside this whitelist, template substitution as a replacement for generation is forbidden.

## 3) No New Diff Regression

- Current PySDK diff baseline is 0.
- Any new logic must not introduce new PySDK diff.
- Before and after every `test`, `lint`, and `commit`, both Python and Go diff file counts must remain 0.

# Config Responsibilities

YAML/JSON config may be used to fill Swagger gaps, including:

- Swagger vs SDK historical mismatch (fields/APIs not synchronized);
- Merging multiple Swagger scopes into one SDK package;
- API name mapping (e.g., `/v3/chat` -> `chat.create` / `chat.stream`);
- Field aliases, type overrides, method signature/call ordering, etc.

Constraint: in the current phase, alignment is based on APIs/fields already implemented in existing Python/Go SDKs.

# Standard Workflow

1. Prepare baseline repositories (clone if missing).
2. Read Python SDK, Go SDK, and Swagger.
3. Implement or adjust the generator (Go language).
4. Generate Python and Go SDKs into the designated output directories.
5. Compare generated outputs with baselines and iteratively reduce diff.
6. Repeat until both Python and Go reach zero diff.

# Downstream coze-py PR Workflow

When asked to sync generated Python SDK output into `coze-py` and complete the PR lifecycle, follow this exact flow:

1. Run generation with CI checks against the target SDK path:
   - `./scripts/genpy.sh --output-sdk <coze-py-path> --ci-check`
2. Inspect actual `coze-py` git diff and identify changed files/functions.
3. Inspect related `coze-sdk-gen` commits (`git log` + `git show`) and map generator/config changes to SDK output changes.
4. Draft PR title/body from facts above:
   - include SDK behavior change summary
   - include generator commit(s) and what config/generation logic changed
   - include validation command and result summary
   - never include literal `\n` text in PR body; use real line breaks only
5. Create branch, commit, and push in `coze-py`.
6. Create or update PR.
   - Prefer `gh pr create` / `gh pr edit` when token scopes allow.
   - If `gh` GraphQL fails due org scope restrictions (for example missing `read:org`), use GitHub REST API as fallback:
     - `PATCH /repos/{owner}/{repo}/pulls/{number}` for title/body updates
     - `POST /repos/{owner}/{repo}/issues/{number}/labels` for required labels
7. Ensure required PR labels are present (for `coze-py`: one of `feature`, `enhancement`, `fix`, `bugfix`, `bug`, `chore`, `documentation`) so label checks pass.
8. Poll PR checks until completion; resolve failures before merge.
9. Merge PR only after required checks pass and branch protection conditions are satisfied.
10. Report final PR URL, status, and merge result.

# Quality Gates

- Language: generator implementation must be Go.
- Scripts: `fmt`, `lint`, `test`, and `build` must exist and be runnable.
- Test coverage: not lower than 80%.
- Test fixtures may be extracted from `coze-openapi.yaml` fragments.
- CI must pass and remain consistent with local quality gates.

# Commit Workflow and Conventions

- Use small, complete commits (one complete objective per commit).
- Commit messages must be in English and follow conventional style (`feat:`, `fix:`, `refactor:`, etc.).
- Use `feat:` for non-trivial API surface changes: added APIs, added fields, modified APIs, or modified fields. Pure API ordering-only changes and comment-only changes are excluded.
- Before each commit, all of the following are required:
  - formatting
  - lint
  - test
  - build
  - Python diff = 0
  - Go diff = 0

# Prohibited Behaviors

- No direct copy/overwrite from baseline SDK code (except whitelist files).
- No commits without verifying diff and quality gates.
- No masking real generation issues by expanding ignore rules only.
- No regression that breaks existing zero-diff baseline.

# Common Commands (Examples)

- Generate Python: `./scripts/genpy.sh`
- Generate Go: `./scripts/gengo.sh`
- Run diff: `./scripts/diff.sh`
- Format: `./scripts/fmt.sh`
- Lint: `./scripts/lint.sh`
- Test: `./scripts/test.sh`
- Build: `./scripts/build.sh`

# Pre-Delivery Checklist

- [ ] Baseline repositories are prepared (or updated)
- [ ] Python generated vs baseline diff file count = 0
- [ ] Go generated vs baseline diff file count = 0
- [ ] `lint` / `test` / `build` all pass
- [ ] Coverage >= 80%
- [ ] Commit message follows conventions and scope is clear
- [ ] If downstream `coze-py` PR is required: PR title/body/labels/checks/merge all completed
