#!/usr/bin/env bash
set -euo pipefail

go run ./cmd/coze-sdk-gen \
  --config config/generator.yaml \
  --swagger exist-repo/coze-openapi-swagger.yaml "$@"
