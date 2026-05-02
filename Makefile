# crap4go Makefile
# Provides convenience targets for building, testing, and linting.
# All commands assume Go is on the PATH.

.PHONY: build test test-race cover lint vet fmt tidy clean install

# --------------------------------------------------------------------------
# Build
# --------------------------------------------------------------------------

## build: compile the crap binary into ./bin/crap
build:
	go build -o bin/crap ./cmd/crap

## install: install crap to $GOPATH/bin (or $GOBIN)
install:
	go install ./cmd/crap

# --------------------------------------------------------------------------
# Testing
# --------------------------------------------------------------------------

## test: run the full test suite
test:
	go test ./...

## test-race: run the full test suite with the race detector
test-race:
	go test -race ./...

## cover: run tests and open an HTML coverage report
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report written to coverage.html"

# --------------------------------------------------------------------------
# Code quality
# --------------------------------------------------------------------------

## vet: run go vet on all packages
vet:
	go vet ./...

## fmt: format all Go source files in-place
fmt:
	go fmt ./...

## lint: run golangci-lint (requires golangci-lint on PATH)
lint:
	golangci-lint run ./...

# --------------------------------------------------------------------------
# Module management
# --------------------------------------------------------------------------

## tidy: tidy the module and verify the module graph
tidy:
	go mod tidy
	go mod verify

# --------------------------------------------------------------------------
# Clean
# --------------------------------------------------------------------------

## clean: remove build artefacts
clean:
	rm -rf bin/ coverage.out coverage.html

# --------------------------------------------------------------------------
# Help
# --------------------------------------------------------------------------

## help: print this help message
help:
	@grep -E '^##' Makefile | sed 's/^## //'
