GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)

staticcheck:
	@command -v staticcheck > /dev/null 2>&1 || go get honnef.co/go/tools/cmd/staticcheck
	@staticcheck -checks="all" -tests $(GOFMT_FILES)
.PHONY: staticcheck

fmtcheck:
	@command -v goimports > /dev/null 2>&1 || go get golang.org/x/tools/cmd/goimports
	@echo "==> Checking that code complies with gofmt requirements..."
	@gofmt_files=$$(gofmt -l $(GOFMT_FILES)); \
	if [ -n "$$gofmt_files" ]; then \
		echo "gofmt needs running on the following files:"; \
		echo "$$gofmt_files"; \
		echo "You can use the command: \`make fmt\` to reformat code."; \
		exit 1; \
	fi
.PHONY: fmtcheck

fmt:
	@command -v goimports > /dev/null 2>&1 || go get golang.org/x/tools/cmd/goimports
	@echo "==> Fixing source code with gofmt..."
	@goimports -w $(GOFMT_FILES)
.PHONY: fmt

lint-ci:
	@command -v golangci-lint > /dev/null 2>&1 || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@golangci-lint run
.PHONY: lint-ci

tidy:
	@go mod tidy
	@go mod vendor
.PHONY: tidy

deps-list:
	@go list -m -u -mod=readonly all
.PHONY: deps-list

deps-update:
	@go get -u -v ./...
	@go mod tidy
	@go mod vendor
.PHONY: deps-update

deps-cleancache:
	@go clean -modcache
.PHONY: deps-cleancache
