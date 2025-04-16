BIN := ./bin/git-cm
ifeq ($(OS),Windows_NT)
BIN := $(BIN).exe
endif
BUILD_LDFLAGS := "-s -w"

.PHONY: all
all: clean tidy check build

.PHONY: build
build:
	CGO_ENABLED=0 go build -trimpath -ldflags=$(BUILD_LDFLAGS) -o $(BIN) .

.PHONY: install
install:
	CGO_ENABLED=0 go install -trimpath -ldflags=$(BUILD_LDFLAGS) .

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
