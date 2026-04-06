.PHONY: all build test lint lint-fix clean check check-readme smoke-test check-config release-dry-run

all: build test smoke-test lint-fix

build:
	go build -o vanity-ssh-keygen ./cmd/vanity-ssh-keygen

test:
	go test ./...

lint:
	golangci-lint run

lint-fix:
	go fix ./...
	golangci-lint run --fix

README.md: build update-readme.sh
	./update-readme.sh

check-readme: README.md
	git diff --exit-code README.md

smoke-test: build
	./vanity-ssh-keygen --metrics x
	grep -i x x.pub
	rm -f x x.pub

check-config:
	KO_DOCKER_REPO=ko.local goreleaser check

release-dry-run:
	KO_DOCKER_REPO=ko.local goreleaser release --snapshot --clean

clean:
	rm -f vanity-ssh-keygen
	rm -f pprof
	rm -rf dist
	rm -f x x.pub

check: lint test check-readme
	go mod tidy -diff
	go fix -diff ./...
