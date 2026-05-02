.PHONY: all test build lint lintmax docker-lint golangci-lint-install gosec govulncheck goreleaser tag-major tag-minor tag-patch release bump-glazed install

all: test build

VERSION=v0.1.14

GOLANGCI_LINT_VERSION ?= $(shell cat .golangci-lint-version)
GOLANGCI_LINT_BIN ?= $(CURDIR)/.bin/golangci-lint

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

lint: golangci-lint-install
	mkdir -p $(LINT_GOCACHE) $(LINT_XDG_CACHE_HOME)
	XDG_CACHE_HOME=$(LINT_XDG_CACHE_HOME) GOCACHE=$(LINT_GOCACHE) $(GOLANGCI_LINT_BIN) run -v ./...

lintmax: golangci-lint-install
	mkdir -p $(LINT_GOCACHE) $(LINT_XDG_CACHE_HOME)
	XDG_CACHE_HOME=$(LINT_XDG_CACHE_HOME) GOCACHE=$(LINT_GOCACHE) $(GOLANGCI_LINT_BIN) run -v --max-same-issues=100 ./...

test:
	go test ./...

build:
	go generate ./...
	go build ./...

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
