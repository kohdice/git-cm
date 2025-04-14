BIN := ./bin/git-cm
ifeq ($(OS),Windows_NT)
BIN := $(BIN).exe
endif
CURRENT_COMMIT := $(shell git rev-parse --short HEAD)
GIT_DIRTY := $(shell if [ -n "$$(git status --porcelain)" ]; then echo "-dirty"; fi)
CURRENT_REVISION := $(CURRENT_COMMIT)$(GIT_DIRTY)
BUILD_LDFLAGS := "-s -w -X main.version=v0.0.0 -X main.revision=$(CURRENT_REVISION)"

.PHONY: all
all: clean tidy check build

.PHONY: build
build:
	CGO_ENABLED=0 go build -trimpath -ldflags=$(BUILD_LDFLAGS) -o $(BIN) .

.PHONY: clean
clean:
	go clean

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: fmt
fmt:
	golangci-lint fmt

.PHONY: lint
lint:
	golangci-lint run

.PHONY: check
check: fmt lint

.PHONY: test
test:
	go test -v -cover ./...

.PHONY: coverage
coverage:
	go test -v -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
