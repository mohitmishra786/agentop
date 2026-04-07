.PHONY: build test run clean install lint release release-dry

VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X github.com/agentop-dev/agentop/cmd.Version=$(VERSION) -X github.com/agentop-dev/agentop/cmd.CommitSHA=$(COMMIT)

build:
	go build -ldflags="$(LDFLAGS)" -o ./bin/agentop .

test:
	go test ./... -v

run:
	go run .

run-watch:
	go run . --watch

lint:
	golangci-lint run

install:
	go install -ldflags="$(LDFLAGS)" .

release-dry:
	goreleaser release --snapshot --clean

release:
	goreleaser release --clean

clean:
	rm -rf ./bin ./dist
