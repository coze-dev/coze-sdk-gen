#!/usr/bin/env bash
set -euo pipefail

gofmt -w $(find . -name '*.go' -not -path './exist-repo/*')
