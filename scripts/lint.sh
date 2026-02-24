#!/usr/bin/env bash
set -euo pipefail

files=$(find . -name '*.go' -not -path './exist-repo/*')
if [ -n "$files" ]; then
  unformatted=$(gofmt -l $files)
  if [ -n "$unformatted" ]; then
    echo "The following files are not gofmt-formatted:"
    echo "$unformatted"
    exit 1
  fi
fi

go vet ./...
