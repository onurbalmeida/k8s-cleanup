APP := k8s-cleanup
PKG := github.com/onurbalmeida/k8s-cleanup
VERSION ?= $(shell git describe --tags --always --dirty || echo v0.0.0)
COMMIT  ?= $(shell git rev-parse --short HEAD || echo none)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X $(PKG)/cmd.Version=$(VERSION) -X $(PKG)/cmd.Commit=$(COMMIT) -X $(PKG)/cmd.Date=$(DATE) -X $(PKG)/cmd.BuiltBy=make

.PHONY: build run test tidy lint e2e clean

build:
	go build -ldflags '$(LDFLAGS)' -o bin/$(APP) .

run:
	go run . run --all-namespaces --older-than 24h

test:
	go test ./...

tidy:
	go mod tidy

lint:
	@command -v golangci-lint >/dev/null || (echo "Install golangci-lint first"; exit 1)
	golangci-lint run

e2e-go:
	go test -v -tags=e2e ./test/e2e

clean:
	rm -rf bin
