# Taken from github.com/RAttab/gonfork
all: build verify test examples

EXAMPLES_GO := $(wildcard examples/*.go)
EXAMPLES := $(patsubst %.go,%,$(EXAMPLES_GO))

verify: vet lint
test: test-cover test-race test-unused test-bench
.PHONY: all verify test

fmt:
	@echo -- format source code
	@go fmt ./...
.PHONY: fmt

sec:
	@echo -- security check
	@gosec ./...
.PHONY: sec

build: fmt
	@echo -- build all packages
	@go install ./...
.PHONY: build

vet: build
	@echo -- static analysis
	@go vet ./...
.PHONY: vet

lint: vet
	@echo -- report coding style issues
	@find . -type f -name "*.go" -exec golint {} \;
.PHONY: lint

test-cover: vet
	@echo -- build and run tests
	@go test -cover -test.short ./...
.PHONY: test-cover

test-race: vet
	@echo -- rerun all tests with race detector
	@GOMAXPROCS=4 go test -test.short -race ./...
.PHONY: test-race

test-all: vet
	@echo -- build and run all tests
	@GOMAXPROCS=4 go test -race ./...

test-cover-anal:
	@echo -- run cover analysis
	go test -coverprofile cover.out ./...
	go tool cover -html=cover.out -o cover.html
.PHONY: test-cover-anal

test-bench:
	@echo -- run benchmarks
	go test -v -bench=.
.PHONY: test-bench

examples_echo:
	@echo -- building examples
examples: examples_echo $(EXAMPLES)

.PHONY: examples examples_echo

%: %.go
	@echo -- build $<
	@go build -o $@ $<

# https://github.com/dominikh/go-tools#tools
test-unused:
	@echo -- run unused code checker
	staticcheck ./...
.PHONY: test-unused

.PHONY:test-all
