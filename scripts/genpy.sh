#!/usr/bin/env bash
set -euo pipefail

OUTPUT_SDK="${OUTPUT_SDK:-exist-repo/coze-py-generated}"
RUN_CI_CHECK="${RUN_CI_CHECK:-0}"
RUN_TESTS="${RUN_TESTS:-1}"
AUTO_INSTALL_POETRY="${AUTO_INSTALL_POETRY:-1}"
declare -a passthrough_args=()

select_poetry_python_with_sqlite3() {
  local -a candidates=(
    "${POETRY_PYTHON:-}"
    python3
    python3.12
    python3.11
    python3.10
    /usr/bin/python3
    /usr/bin/python3.12
    /usr/bin/python3.11
    /usr/bin/python3.10
    /usr/local/bin/python3
    /usr/local/bin/python3.12
    /usr/local/bin/python3.11
    /usr/local/bin/python3.10
  )
  local candidate=""
  local candidate_path=""

  for candidate in "${candidates[@]}"; do
    if [ -z "$candidate" ] || ! command -v "$candidate" >/dev/null 2>&1; then
      continue
    fi
    candidate_path="$(command -v "$candidate")"
    if ! "$candidate_path" -c "import sqlite3" >/dev/null 2>&1; then
      echo "[coze-py-ci] skip python without sqlite3: $candidate_path"
      continue
    fi
    echo "[coze-py-ci] poetry env use $candidate_path"
    if poetry env use "$candidate_path" >/dev/null 2>&1; then
      echo "[coze-py-ci] selected python: $candidate_path"
      return 0
    fi
    echo "[coze-py-ci] skip incompatible python: $candidate_path"
  done

  echo "[coze-py-ci] no usable python interpreter with sqlite3 support was found."
  return 1
}

usage() {
  cat <<'EOF'
Usage: ./scripts/genpy.sh [options] [extra-generator-args...]

Options:
  --output-sdk <dir>         Output SDK directory (default: exist-repo/coze-py-generated)
  --ci-check                 Run coze-py CI parity checks after generation
  --no-ci-check              Do not run coze-py CI parity checks (default)
  --run-tests                Run pytest in CI parity checks (default)
  --skip-tests               Skip pytest in CI parity checks
  --auto-install-poetry      Auto install poetry when missing (default)
  --no-auto-install-poetry   Do not auto install poetry
  -h, --help                 Show this help message
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
    --run-tests)
      RUN_TESTS=1
      shift
      ;;
    --skip-tests)
      RUN_TESTS=0
      shift
      ;;
    --auto-install-poetry)
      AUTO_INSTALL_POETRY=1
      shift
      ;;
    --no-auto-install-poetry)
      AUTO_INSTALL_POETRY=0
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
  --language python \
  --output-sdk "$OUTPUT_SDK" \
  "${passthrough_args[@]}"

PYPROJECT_CONFIG=""
if [ -f "$OUTPUT_SDK/pyproject.toml" ]; then
  PYPROJECT_CONFIG="$OUTPUT_SDK/pyproject.toml"
elif [ -f "exist-repo/coze-py/pyproject.toml" ]; then
  PYPROJECT_CONFIG="exist-repo/coze-py/pyproject.toml"
fi

if [ -z "$PYPROJECT_CONFIG" ]; then
  echo "pyproject.toml not found for formatter config."
  echo "Expected one of:"
  echo "  - $OUTPUT_SDK/pyproject.toml"
  echo "  - exist-repo/coze-py/pyproject.toml"
  exit 1
fi

if [ ! -d "$OUTPUT_SDK/cozepy" ]; then
  echo "generated package directory not found: $OUTPUT_SDK/cozepy"
  exit 1
fi

ruff check --fix \
  --config "$PYPROJECT_CONFIG" \
  "$OUTPUT_SDK/cozepy"

ruff format \
  --config "$PYPROJECT_CONFIG" \
  "$OUTPUT_SDK/cozepy"

if [ "$RUN_CI_CHECK" = "1" ]; then
  if [ ! -f "$OUTPUT_SDK/pyproject.toml" ]; then
    echo "CI parity checks require a full coze-py repository output directory."
    echo "Please set --output-sdk to the coze-py repo path."
    exit 1
  fi

  if ! command -v poetry >/dev/null 2>&1; then
    if [ "$AUTO_INSTALL_POETRY" = "1" ]; then
      if command -v pipx >/dev/null 2>&1; then
        echo "poetry not found, installing via pipx ..."
        pipx install poetry || pipx upgrade poetry || true
        export PATH="$HOME/.local/bin:$PATH"
      else
        echo "poetry not found, installing via pip --user ..."
        python3 -m pip install --user --break-system-packages poetry
        export PATH="$HOME/.local/bin:$PATH"
      fi
    fi
  fi

  if ! command -v poetry >/dev/null 2>&1; then
    echo "poetry not found in PATH after installation attempt."
    echo "Install poetry manually, or use --auto-install-poetry."
    exit 1
  fi

  pushd "$OUTPUT_SDK" >/dev/null
  if ! select_poetry_python_with_sqlite3; then
    echo "Install a python interpreter with sqlite3 support and retry."
    exit 1
  fi
  echo "[coze-py-ci] poetry install"
  poetry install --no-interaction
  echo "[coze-py-ci] poetry build"
  poetry build
  echo "[coze-py-ci] ruff check"
  poetry run ruff check cozepy
  echo "[coze-py-ci] ruff format --check"
  poetry run ruff format --check
  echo "[coze-py-ci] mypy"
  poetry run mypy .
  if [ "$RUN_TESTS" = "1" ]; then
    echo "[coze-py-ci] pytest"
    poetry run pytest --cov --cov-report=xml
  else
    echo "[coze-py-ci] skip pytest (RUN_TESTS=$RUN_TESTS)"
  fi
  popd >/dev/null
fi
