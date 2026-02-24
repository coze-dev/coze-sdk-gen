.PHONY: fmt lint test build check

fmt:
	./scripts/fmt.sh

lint:
	./scripts/lint.sh

test:
	./scripts/test.sh

build:
	./scripts/build.sh

check: lint test build
