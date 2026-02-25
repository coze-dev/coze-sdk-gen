#!/usr/bin/env bash
set -euo pipefail

go run ./cmd/coze-sdk-gen \
  --config config/generator.yaml \
  --swagger coze-openapi.yaml \
  --language go \
  --output-sdk exist-repo/coze-go-generated "$@"

# Format generated Go SDK source files.
while IFS= read -r -d '' file; do
  gofmt -w "$file"
done < <(find exist-repo/coze-go-generated -type f -name '*.go' -print0)
