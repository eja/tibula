.PHONY: clean test lint tibula release-dry-run release

PACKAGE_NAME := github.com/eja/tibula
GOLANG_CROSS_VERSION := v1.20
GOPATH ?= '$(HOME)/go'

all: lint tibula

clean:
	@rm -f tibula tibula.exe

lint:
	@gofmt -w .

test:
	@go vet ./...
	@go test -v ./test

tibula:
	@go build -ldflags "-s -w" -o tibula cmd/tibula/main.go

