# coze-sdk-gen

A Go-based SDK generator for Coze OpenAPI.

Current status:
- Language support: `python` (implemented), `go` (planned)
- Source of truth: OpenAPI Swagger + generator config
- Compatibility target: `exist-repo/coze-py`
- Generator behavior: generate Python sources from Swagger (not direct SDK copy)

## Why config is required

The existing Python SDK and Swagger are not perfectly aligned. The generator uses `config/generator.yaml` to fill missing metadata, including:
- API package grouping rules
- SDK method aliases for the same HTTP endpoint
- Field alias hints
- Legacy operations/packages that are missing in Swagger

## Quick start

1. Generate Python SDK:

```bash
./scripts/generate.sh
```

This writes output to `exist-repo/coze-py-generated`.

2. Compare generated SDK with legacy SDK:

```bash
./scripts/diff.sh
```

The script prints file-level differences. The current implementation is Swagger-driven and still iterating toward full legacy parity.

## Development scripts

- Format: `./scripts/fmt.sh`
- Lint: `./scripts/lint.sh`
- Test (with coverage > 80% gate): `./scripts/test.sh`
- Build: `./scripts/build.sh`

Or run all checks:

```bash
make check
```

## CLI usage

```bash
go run ./cmd/coze-sdk-gen \
  --config config/generator.yaml \
  --swagger exist-repo/coze-openapi-swagger.yaml
```

Output example:

```text
language=python generated_files=23 generated_ops=69 output=exist-repo/coze-py-generated
```

Optional overrides:
- `--language`
- `--source-sdk`
- `--output-sdk`
