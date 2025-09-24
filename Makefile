BINARY_NAME=nas-manager
VERSION=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X github.com/SlashGordon/scripts/cmd.Version=${VERSION} -X github.com/SlashGordon/scripts/cmd.Commit=${COMMIT} -X github.com/SlashGordon/scripts/cmd.Date=${BUILD_TIME}"

.PHONY: build clean release lint test fmt vet

build:
	go build ${LDFLAGS} -o bin/${BINARY_NAME} .

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-linux-amd64 .

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-linux-arm64 .

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-amd64 .

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-arm64 .

build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64

clean:
	rm -rf bin/

release: clean build-all
	@echo "Built binaries:"
	@ls -la bin/

# Go Report Card improvements
fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test -v ./...

lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

check: fmt vet test lint
	@echo "All checks passed!"