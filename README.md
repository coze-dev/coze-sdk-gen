# coze-sdk-gen

A Go-based SDK generator for Coze OpenAPI.

## Status

- Supported language: `python`
- In progress: `go`
- Generation model: Swagger/OpenAPI as source of truth + config-based metadata

## Why Config Is Needed

The legacy SDK shape and the OpenAPI document are not fully 1:1.  
`config/generator.yaml` is used to provide generation metadata, for example:

- API package grouping
- SDK method aliases for the same endpoint
- field alias/type overrides
- legacy-compatible behavior not directly expressible in Swagger

## Quick Start

1. Run Python generator:

```bash
./scripts/genpy.sh
```

Generate directly into a real `coze-py` repository:

```bash
./scripts/genpy.sh --output-sdk /path/to/coze-py
```

Run with CI-parity checks (build, lint/type-check, tests):

```bash
./scripts/genpy.sh --output-sdk /path/to/coze-py --ci-check
```

Run Go generator:

```bash
./scripts/gengo.sh
```

2. Compare generated SDK with baseline SDK:

```bash
./scripts/diff.sh
```

## CLI

```bash
go run ./cmd/coze-sdk-gen \
  --config config/generator.yaml \
  --swagger ./coze-openapi.yaml
```

Output example:

```text
language=python generated_files=57 generated_ops=86 output=<configured-output-dir>
```

Optional overrides:

- `--language`
- `--output-sdk`

## Development Scripts

- format: `./scripts/fmt.sh`
- lint: `./scripts/lint.sh`
- test (coverage gate): `./scripts/test.sh`
- build: `./scripts/build.sh`

Run full checks:

```bash
make check
```

## coze-py CI Parity Check

Use a single entrypoint:

```bash
./scripts/genpy.sh --output-sdk /path/to/coze-py --ci-check
```

or:

```bash
COZE_PY_DIR=/path/to/coze-py make check-coze-py
```

`scripts/genpy.sh` runs:
1. generation
2. `ruff check --fix`
3. `ruff format`
4. optional `coze-py` CI-parity checks (`build`, `ruff check`, `ruff format --check`, `mypy`, `pytest`)
