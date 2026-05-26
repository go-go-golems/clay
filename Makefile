.PHONY: all test build lint lintmax docker-lint golangci-lint-install glazed-lint-build glazed-lint gosec govulncheck goreleaser tag-major tag-minor tag-patch release bump-glazed install logcopter-generate logcopter-check

all: test build

VERSION=v0.1.14

GOLANGCI_LINT_VERSION ?= $(shell cat .golangci-lint-version)
GOLANGCI_LINT_BIN ?= $(CURDIR)/.bin/golangci-lint
GLAZED_LINT_BIN ?= /tmp/glazed-lint
GLAZED_LINT_PKG ?= github.com/go-go-golems/glazed/cmd/tools/glazed-lint
GLAZED_VERSION ?= $(shell GOWORK=off go list -m -f '{{.Version}}' github.com/go-go-golems/glazed 2>/dev/null)
GLAZED_LINT_FLAGS ?= -glazedclilint.allow-paths=pkg/analysis/,pkg/cli/,pkg/cmds/fields/,pkg/cmds/logging/,pkg/cmds/sources/,pkg/help/,pkg/cmds/profiles/,pkg/cmds/commandmeta/edit.go

CACHE_DIR := $(CURDIR)/.cache
LINT_GOCACHE := $(CACHE_DIR)/go-build
LINT_XDG_CACHE_HOME := $(CACHE_DIR)/xdg

TAPES := $(wildcard doc/vhs/*.tape)
gifs: $(TAPES)
	for i in $(TAPES); do vhs < $$i; done

docker-lint:
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) golangci-lint run -v ./...

golangci-lint-install:
	mkdir -p $(dir $(GOLANGCI_LINT_BIN))
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(dir $(GOLANGCI_LINT_BIN)) $(GOLANGCI_LINT_VERSION)

glazed-lint-build:
	@echo "Building glazed-lint from Glazed module..."
	@if [ -n "$(GLAZED_VERSION)" ] && [ "$(GLAZED_VERSION)" != "(devel)" ]; then \
		echo "Installing $(GLAZED_LINT_PKG)@$(GLAZED_VERSION)"; \
		GOBIN=$(dir $(GLAZED_LINT_BIN)) GOWORK=off go install $(GLAZED_LINT_PKG)@$(GLAZED_VERSION); \
	else \
		echo "Installing $(GLAZED_LINT_PKG) from workspace/module"; \
		GOBIN=$(dir $(GLAZED_LINT_BIN)) go install $(GLAZED_LINT_PKG); \
	fi

glazed-lint: glazed-lint-build
	go vet -vettool=$(GLAZED_LINT_BIN) $(GLAZED_LINT_FLAGS) ./pkg/...

lint: glazed-lint-build golangci-lint-install
	mkdir -p $(LINT_GOCACHE) $(LINT_XDG_CACHE_HOME)
	XDG_CACHE_HOME=$(LINT_XDG_CACHE_HOME) GOCACHE=$(LINT_GOCACHE) $(GOLANGCI_LINT_BIN) run -v ./...
	go vet -vettool=$(GLAZED_LINT_BIN) $(GLAZED_LINT_FLAGS) ./pkg/...

lintmax: glazed-lint-build golangci-lint-install
	mkdir -p $(LINT_GOCACHE) $(LINT_XDG_CACHE_HOME)
	XDG_CACHE_HOME=$(LINT_XDG_CACHE_HOME) GOCACHE=$(LINT_GOCACHE) $(GOLANGCI_LINT_BIN) run -v --max-same-issues=100 ./...
	go vet -vettool=$(GLAZED_LINT_BIN) $(GLAZED_LINT_FLAGS) ./pkg/...

test:
	go test ./...

build:
	go generate ./...
	go build ./...

logcopter-generate:
	go generate ./...

logcopter-check:
	go tool logcopter-gen -area-prefix go-go-golems.clay -strip-prefix github.com/go-go-golems/clay -check ./pkg/...

tag-major:
	git tag $(shell svu major)

tag-minor:
	git tag $(shell svu minor)

tag-patch:
	git tag $(shell svu patch)

release:
	git push --tags origin
	GOPROXY=proxy.golang.org go list -m github.com/go-go-golems/clay@$(shell svu current)

bump-glazed:
	go get github.com/go-go-golems/glazed@latest
	go get github.com/go-go-golems/logcopter@latest
	go mod tidy

install:
	go build -o ./dist/clay ./cmd/clay && \
		cp ./dist/clay $(shell which clay)

gosec:
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	gosec -exclude=G101,G304,G301,G306,G204 -exclude-dir=.history ./...

govulncheck:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...
