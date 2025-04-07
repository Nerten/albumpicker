
all: lint test build

build:
	goreleaser build --single-target --snapshot --clean -o albumpicker

test:
	go clean -testcache
	go test -race -coverprofile=coverage.out ./...
	grep -v "_mock.go" coverage.out | grep -v mocks > coverage_no_mocks.out
	go tool cover -func=coverage_no_mocks.out
	rm coverage.out coverage_no_mocks.out

update:
	go get -u ./...
	go mod tidy

lint:
	golangci-lint run

.PHONY: build test update lint
