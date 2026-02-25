#!/usr/bin/env bash
set -euo pipefail

go run ./cmd/coze-sdk-gen \
  --config config/generator.yaml \
  --swagger coze-openapi.yaml \
  --language python \
  --output-sdk exist-repo/coze-py-generated "$@"

ruff format \
  --config exist-repo/coze-py/pyproject.toml \
  exist-repo/coze-py-generated/cozepy
