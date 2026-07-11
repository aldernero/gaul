GOLANGCI_LINT := $(shell command -v golangci-lint 2>/dev/null || command -v golangci-lint-v2 2>/dev/null)

.PHONY: test lint

test:
	go test ./...

lint:
	@if [ -z "$(GOLANGCI_LINT)" ]; then \
		echo "golangci-lint not found; install from https://golangci-lint.run/welcome/install/"; \
		exit 1; \
	fi
	$(GOLANGCI_LINT) run ./...
