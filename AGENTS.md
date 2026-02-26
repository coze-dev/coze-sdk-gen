# Background

This document defines execution rules for `coze-sdk-gen`, targeting automated coding agents (for example Codex/Claude), with the goal of a reproducible and auditable implementation process.

# Style Reference (AGENTS.md / CLAUDE.md)

This file follows common `AGENTS.md` / `CLAUDE.md` best practices:

- Define objective, inputs, outputs, and acceptance criteria first.
- Explicitly define allowed vs. disallowed behavior (especially no direct copy from baseline SDKs).
- Constrain workflow before implementation details.
- Require verifiable checks for every change (`lint` / `test` / `build`, and `diffgo` for Go baseline alignment).
- Priority order: this file > temporary verbal preference > default behavior.

# Inputs and Baseline Repositories

- Swagger input: `coze-openapi.yaml`
- Python baseline repository: `https://github.com/coze-dev/coze-py`
  - Local baseline mirror directory: choose a local path for `coze-py` (for example `<baseline-root>/coze-py`)
- Go baseline repository: `https://github.com/coze-dev/coze-go`
  - Local baseline mirror directory: choose a local path for `coze-go` (for example `<baseline-root>/coze-go`)

# Repository Bootstrap

If baseline repositories are missing locally, they must be cloned before generation/alignment:

- `mkdir -p <baseline-root>`
- `test -d <baseline-root>/coze-py || git clone https://github.com/coze-dev/coze-py <baseline-root>/coze-py`
- `test -d <baseline-root>/coze-go || git clone https://github.com/coze-dev/coze-go <baseline-root>/coze-go`

# Core Implementation Principles

## 1) Swagger-First (Default)

- Non-special files (by default, all files) must be produced via Swagger parsing + generation logic.
- Direct copying/overwriting from local baseline mirrors (`coze-py` / `coze-go`) is forbidden.
- Field/API changes should be handled through Swagger + config convergence first.

## 2) Special-File Whitelist (Exception Mechanism)

Only the following files (runtime infrastructure without direct OpenAPI equivalents) may be template-based:

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

Outside this whitelist, using templates as a replacement for generation logic is forbidden.

# Config Responsibilities

YAML/JSON config can be used to fill Swagger gaps, including:

- Swagger vs SDK historical mismatch (fields/APIs not synchronized)
- Merging multiple Swagger scopes into one SDK package
- API naming mapping (for example `/v3/chat` -> `chat.create` / `chat.stream`)
- Field aliases, type overrides, method signature/call ordering, etc.

Constraint: alignment in the current phase is limited to APIs/fields already implemented in existing Python/Go SDKs.

# Quality Gates

- Language: generator implementation must be Go.
- Scripts: `fmt`, `lint`, `test`, and `build` must exist and be runnable.
- Test coverage: not lower than 80%.
- Test fixtures may be extracted from `coze-openapi.yaml` fragments.
- CI must pass and remain consistent with local quality gates.

# Prohibited Behaviors

- No direct copy/overwrite from baseline SDK code (except whitelist files).
- No commits without verifying required checks and quality gates.
- No masking real generation issues by only expanding ignore rules.
- No regression that breaks existing Go zero-diff baseline.

# Common Commands (Examples)

- Generate Python: `./scripts/genpy.sh`
- Generate Go: `./scripts/gengo.sh`
- Run Go diff: `./scripts/diffgo.sh`
- Format: `./scripts/fmt.sh`
- Lint: `./scripts/lint.sh`
- Test: `./scripts/test.sh`
- Build: `./scripts/build.sh`

# Pre-Delivery Checklist

- [ ] Baseline repositories are prepared (or updated)
- [ ] Python generation checks pass (`./scripts/genpy.sh --output-sdk <coze-py-path> --ci-check`)
- [ ] Go generated vs baseline diff file count = 0
- [ ] `lint` / `test` / `build` all pass
- [ ] Coverage >= 80%
- [ ] Commit messages follow conventions and scope is clear
- [ ] If downstream `coze-py` PR is involved: PR title/body/labels/checks/merge flow is fully completed

# Python SDK Development Workflow

When asked to sync generated Python SDK output to `coze-py` and complete the PR lifecycle, follow the rules first (Rule 1 and Rule 2), then execute by stages.

### Pre-Execution Sync (Mandatory)

Before starting implementation or generation, pull remote `origin/main` into local `main` in both repositories and ensure local is not behind:

1. Pull `origin/main` into local `coze-sdk-gen/main`:
   - `git fetch origin`
   - `git checkout main`
   - `git pull --ff-only origin main`
2. Pull `origin/main` into local `coze-py/main` (in downstream local clone/worktree):
   - `git fetch origin`
   - `git checkout main`
   - `git pull --ff-only origin main`
3. If either repository cannot fast-forward local `main` to `origin/main`, resolve sync first; do not start task execution until both are up to date.

### Rule 1: GitHub PR Rules

1. PRs must be traceable: when creating/updating a PR, title, description, related change scope, and validation info must be complete and auditable.
2. The agent must not auto-merge PRs.
   - Merge actions are allowed only after explicit user merge instructions.
   - When a user explicitly instructs to merge a PR in `coze-py` or `coze-go`, use exactly: `gh pr merge <id> --squash --admin`.
   - For `coze-py` and `coze-go`, do not use other merge methods or options.
   - Do not require pre-approval and do not block on fetching approval first.
3. Human review has priority: continuously read and handle human review comments.
4. Bot handling rule: ignore bot comments/statuses (for example `CodeRabbit`) as review-decision input.
5. After feedback is addressed, you must push and keep updating the same PR.
   - Human approval status can be tracked for visibility, but it is not a prerequisite for merge once explicit user merge instruction is given.
6. Continuously poll PR checks; except explicitly ignored items, all required checks must pass before entering next stages.
7. PR descriptions must not contain literal `\\n`; use real line breaks.
8. PR traceability must be bidirectional:
   - Downstream `coze-py` PR must include the related upstream `coze-sdk-gen` PR link.
   - Upstream `coze-sdk-gen` PR must be backfilled with the downstream `coze-py` PR link after downstream PR creation.
9. For Python SDK workflow tasks, both repositories must have PRs:
   - `coze-sdk-gen` PR is required.
   - `coze-py` PR is required.
   - Do not mark the task complete until both PRs exist and are traceable.
10. PR create/update tool strategy:
   - Prefer `gh pr create` / `gh pr edit`.
   - If `gh` fails due permission issues (for example missing `read:org`), use REST API fallback:
     - `PATCH /repos/{owner}/{repo}/pulls/{number}`
     - `POST /repos/{owner}/{repo}/issues/{number}/labels`
11. After each PR merge, the report must include the PR title in addition to existing merge information.
12. Zero-diff downstream rule:
   - If regenerated `coze-py` has 0 file diff, do not close/abandon the downstream PR due to zero diff.
   - Keep traceability and follow the normal downstream merge flow when user explicitly instructs merge.
   - If a new downstream PR is required but there is no file diff, use a traceability PR (for example an empty commit), then handle it as a normal PR.

### Rule 2: Commit Rules

1. Use small, complete commits (one complete objective per commit).
2. Commit messages must be in English and follow conventional prefixes (`feat`, `fix`, `refactor`, etc.).
3. Commit scope limitation applies only to the `coze-sdk-gen` repository:
   - In `coze-sdk-gen`, commit scope format is mandatory and must follow this mapping:
     - Python SDK changes only: `<type>(py): ...`
     - Go SDK changes only: `<type>(go): ...`
     - Generic codegen changes (not specifically py/go): `<type>(codegen): ...`
     - Changes that impact both py and go simultaneously: `<type>: ...` (no scope)
   - In `coze-py` and `coze-go`, scope is not required.
4. For non-trivial API interface changes (new API, new field, API modification, field modification), `feat` is required. Pure API ordering changes or comment-only changes are excluded.
5. Local gates before commit must pass: formatting, lint, test, build.
6. Validate git user for both `coze-py` and `coze-sdk-gen` (must be repository-level config):
   - `user.name = chyroc`
   - `user.email = chyroc@qq.com`

### Stage 1: Baseline Alignment and Generator Preparation

1. Run this Python SDK workflow in temporary directories to avoid polluting long-lived working directories.
2. Clone each required repository into a random directory under `/tmp` (no historical directory reuse), for example: `mktemp -d /tmp/coze-py-XXXXXX`.
3. Immediately after clone, pull `origin/main` into both local `main` branches using fast-forward only (same requirement as Pre-Execution Sync).
4. Read Python SDK and Swagger.
5. Implement or adjust the generator (implemented in Go).
6. Generate Python SDK and iterate toward the target state.

### Stage 2: Prepare `coze-sdk-gen` Changes

1. Complete `coze-sdk-gen` changes in this repository first:
   - Commit behavior must satisfy Rule 2.
   - Push the working branch.
   - Create or update the `coze-sdk-gen` PR, and keep that PR updated in later changes.
   - `coze-sdk-gen` PR handling must satisfy Rule 1.
   - In this workflow, `coze-sdk-gen` PR is mandatory (paired with downstream `coze-py` PR).
2. Do not merge `coze-sdk-gen` branch into `main` at this stage; only do so after explicit user merge instruction.

### Stage 3: Generate and Update Downstream PR

1. Run generation and checks:
   - `./scripts/genpy.sh --output-sdk <coze-py-path> --ci-check`
   - If codegen or Python checks fail (including environment issues), fix and rerun until all pass.
   - Keep failure reporting concise; avoid unnecessary low-level environment details.
2. Prepare change summary from facts:
   - Read actual `coze-py` changes (file/function-level changes).
   - Read related `coze-sdk-gen` commits (`git log` + `git show`) and map generator/config changes to SDK output changes.
   - PR title/description must include: behavior changes, generator changes, and validation results.
3. Commit and push in `coze-py`, then create or update PR:
   - Execution method must follow Rule 1.
4. Ensure PR has a required label selected from labels that already exist in the target repository.
   - First list available labels (for example via `gh label list`) and choose one matching intent.
   - Preferred semantic set: `feature`, `enhancement`, `fix`, `bugfix`, `bug`, `chore`, `documentation`.
   - If `fix`/`bugfix` does not exist, choose the closest existing equivalent (for example `bug`); do not assume a label exists.

### Stage 4: Post-Merge-Instruction Release Sync and Merge

1. Trigger condition: enter this stage only after explicit user merge instruction.
2. Sync `coze-sdk-gen` mainline:
   - Merge the `coze-sdk-gen` PR.
   - Update local `coze-sdk-gen/main` to the latest remote code.
   - Ensure `main` is pushed to remote.
3. Regenerate and backfill downstream PR based on latest `coze-sdk-gen/main`:
   - Rerun Python SDK generation and checks.
   - Update the same `coze-py` PR with regenerated output and push.
4. Handle downstream merge by user instruction:
   - When user explicitly instructs to merge downstream `coze-py` or `coze-go` PR, use exactly: `gh pr merge <id> --squash --admin`.
   - For `coze-py` and `coze-go`, do not use other merge methods or options.
   - Do not require prior approve state and do not block waiting for approve status before merge.
   - If new commits trigger new checks or comments and user asks for follow-up, return to Stage 3.
5. Final report:
   - `coze-sdk-gen` PR URL / title / status / merge result
   - Downstream `coze-py` PR URL / title / status / merge result
   - `coze-sdk-gen` commit / push result
   - Task completion response must explicitly include both PR links.
