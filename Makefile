# Taken from github.com/RAttab/gonfork
all: build verify test
verify: vet lint
test: test-cover test-race
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
	@GOMAXPROCS=4 go test -v -test.short -race ./...
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
	go test -v -bench=./...
.PHONY: test-bench

# https://github.com/dominikh/go-tools#tools
test-unused:
	@echo -- run unused code checker
	unused ./...
.PHONY: test-unused

.PHONY:test-all
