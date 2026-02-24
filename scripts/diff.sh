#!/usr/bin/env bash
set -euo pipefail

diff -ru --exclude=.git exist-repo/coze-py exist-repo/coze-py-generated
