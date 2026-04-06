.PHONY: build test run clean install lint release-dry

build:
	go build -ldflags="-s -w" -o ./bin/agentop .

test:
	go test ./... -v

run:
	go run .

run-watch:
	go run . --watch

lint:
	golangci-lint run

install:
	go install .

release-dry:
	goreleaser release --snapshot --clean

clean:
	rm -rf ./bin ./dist
