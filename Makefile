# ---- Config ----
BACKEND_DIR := backend
GO := go
LINT := golangci-lint

# ---- Default Target ----
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make lint      - Run golangci-lint"
	@echo "  make test      - Run backend tests"
	@echo "  make build     - Build backend"
	@echo "  make ci        - Run lint + test + build (CI equivalent)"

# ---- Lint ----
.PHONY: lint
lint:
	cd $(BACKEND_DIR) && $(LINT) run

# ---- Test ----
.PHONY: test
test:
	cd $(BACKEND_DIR) && $(GO) test -race ./...

# ---- Build ----
.PHONY: build
build:
	cd $(BACKEND_DIR) && $(GO) build ./...

# ---- CI (local equivalent of GitHub CI) ----
.PHONY: ci
ci: lint test build
