all: build verify test
verify: vet lint
test: test-cover test-race

fmt:
	@echo -- format source code
	@go fmt ./...

build: fmt
	@echo -- build all packages
	@go install ./...

vet: build
	@echo -- static analysis
	@go vet ./...

lint: vet
	@echo -- report coding style issues
	@find . -type f -name *.go -exec golint {} \;

test-cover: vet
	@echo -- build and run tests
	@go test -cover -test.short ./...

test-race: vet
	@echo -- rerun all tests with race detector
	@GOMAXPROCS=4 go test -test.short -race ./...

test-all: vet
	@echo -- build and run all tests
	@GOMAXPROCS=4 go test -race ./...


.PHONY: all fmt build vet lint verify test-cover test-race test