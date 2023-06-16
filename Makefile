.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: lint
lint: ## Lint Go source files
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@golangci-lint run

.PHONY: download
download:
	@go mod download

.PHONY: tidy
tidy: ## Run go mod tidy and vendor dependencies
	@go mod tidy
	@go mod vendor

.PHONY: clearcache
clearcache: ## Clear the cache
	@go clean -cache -testcache -modcache

.PHONY: fmt
fmt: ## Format Go source files
	@go fmt ./...

.PHONY: test
test: ## Run unit tests
	@go test -v ./...
