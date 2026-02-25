#!/usr/bin/env bash
set -euo pipefail

OUTPUT_SDK="${OUTPUT_SDK:-exist-repo/coze-go-generated}"
RUN_CI_CHECK="${RUN_CI_CHECK:-0}"
declare -a passthrough_args=()

usage() {
  cat <<'EOF'
Usage: ./scripts/gengo.sh [options] [extra-generator-args...]

Options:
  --output-sdk <dir>    Output SDK directory (default: exist-repo/coze-go-generated)
  --ci-check            Run additional CI checks after generation
  --no-ci-check         Skip additional CI checks (default)
  -h, --help            Show this help message
EOF
}

while [ $# -gt 0 ]; do
  case "$1" in
    --output-sdk)
      if [ $# -lt 2 ]; then
        echo "missing value for --output-sdk"
        exit 1
      fi
      OUTPUT_SDK="$2"
      shift 2
      ;;
    --ci-check)
      RUN_CI_CHECK=1
      shift
      ;;
    --no-ci-check)
      RUN_CI_CHECK=0
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      passthrough_args+=("$1")
      shift
      ;;
  esac
done

go run ./cmd/coze-sdk-gen \
  --config config/generator.yaml \
  --swagger coze-openapi.yaml \
  --language go \
  --output-sdk "$OUTPUT_SDK" \
  "${passthrough_args[@]}"

# Format generated Go SDK source files.
while IFS= read -r -d '' file; do
  gofmt -w "$file"
done < <(find "$OUTPUT_SDK" -type f -name '*.go' -print0)

if [ "$RUN_CI_CHECK" = "1" ]; then
  if [ ! -f "$OUTPUT_SDK/go.mod" ]; then
    echo "CI checks require a full coze-go repository output directory."
    echo "Please set --output-sdk to the coze-go repo path."
    exit 1
  fi

  pushd "$OUTPUT_SDK" >/dev/null
  echo "[coze-go-ci] go test ./... (quiet mode)"
  test_log="$(mktemp -t coze-go-test.XXXXXX.log)"
  if ! go test ./... >"$test_log" 2>&1; then
    echo "[coze-go-ci] go test failed, full output:"
    cat "$test_log"
    rm -f "$test_log"
    popd >/dev/null
    exit 1
  fi
  rm -f "$test_log"
  echo "[coze-go-ci] go test passed"
  if [ -d .git ]; then
    if [ -n "$(git status --porcelain)" ]; then
      echo "[coze-go-ci] generated output has diff:"
      git status --short
      popd >/dev/null
      exit 1
    fi
  fi
  popd >/dev/null
fi
