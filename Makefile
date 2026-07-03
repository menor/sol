.PHONY: build install test lint clean release-dry-run

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null | sed 's/^v//' || echo "dev")
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w \
	-X github.com/menor/sol/cmd.version=$(VERSION) \
	-X github.com/menor/sol/cmd.commit=$(COMMIT) \
	-X github.com/menor/sol/cmd.date=$(DATE)

build:
	go build -ldflags '$(LDFLAGS)' -o sol .

install:
	go install -ldflags '$(LDFLAGS)' .

test:
	go test -race ./...

lint:
	golangci-lint run

clean:
	rm -f sol
	rm -rf dist/

release-dry-run:
	goreleaser release --snapshot --clean
