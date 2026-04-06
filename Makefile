.PHONY: all build test lint lint-fix clean check update-readme check-config release-dry-run

all: build test lint

build:
	go build -o vanity-ssh-keygen ./cmd/vanity-ssh-keygen

test:
	go test -v ./...

lint:
	golangci-lint run

lint-fix:
	go fix ./...
	golangci-lint run --fix

update-readme:
	./update-readme.sh

check-config:
	KO_DOCKER_REPO=ko.local goreleaser check

release-dry-run:
	KO_DOCKER_REPO=ko.local goreleaser release --snapshot --clean

clean:
	rm -f vanity-ssh-keygen
	rm -f pprof
	rm -rf dist

check: lint test
	go mod tidy -diff
	go fix -diff ./...
