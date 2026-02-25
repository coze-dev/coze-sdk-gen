#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONFIG_PATH="${CONFIG_PATH:-$ROOT_DIR/config/generator.yaml}"
DIFF_LANGUAGE="${DIFF_LANGUAGE:-python}"

mapfile -t DIFF_EXCLUDES < <(
  python3 - "$CONFIG_PATH" "$DIFF_LANGUAGE" <<'PY'
import sys
import yaml

config_path = sys.argv[1]
language = sys.argv[2].strip().lower()

with open(config_path, "r", encoding="utf-8") as f:
    cfg = yaml.safe_load(f) or {}

paths = (
    ((cfg.get("diff") or {}).get("ignore_paths_by_language") or {}).get(language)
    or []
)
for p in paths:
    text = str(p).strip()
    if text:
        print(text)
PY
)

if [[ ${#DIFF_EXCLUDES[@]} -eq 0 ]]; then
  DIFF_EXCLUDES=(.git)
fi

DIFF_CMD=(diff -ru)
for exclude in "${DIFF_EXCLUDES[@]}"; do
  DIFF_CMD+=(--exclude="$exclude")
done
DIFF_CMD+=("$ROOT_DIR/exist-repo/coze-py" "$ROOT_DIR/exist-repo/coze-py-generated")
"${DIFF_CMD[@]}"
