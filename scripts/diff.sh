#!/usr/bin/env bash
set -euo pipefail

diff -ru \
  --exclude=.git \
  --exclude=.github \
  --exclude=.gitignore \
  --exclude=.pre-commit-config.yaml \
  --exclude=.vscode \
  --exclude=CONTRIBUTING.md \
  --exclude=LICENSE \
  --exclude=README.md \
  --exclude=codecov.yml \
  --exclude=examples \
  --exclude=poetry.lock \
  --exclude=__pycache__ \
  --exclude=tests \
  exist-repo/coze-py \
  exist-repo/coze-py-generated
