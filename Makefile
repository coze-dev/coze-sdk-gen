.PHONY: fmt lint test build check check-coze-py

fmt:
	./scripts/fmt.sh

lint:
	./scripts/lint.sh

test:
	./scripts/test.sh

build:
	./scripts/build.sh

check-coze-py:
	./scripts/genpy.sh --output-sdk $${COZE_PY_DIR:-exist-repo/coze-py} --ci-check

check: lint test build check-coze-py
