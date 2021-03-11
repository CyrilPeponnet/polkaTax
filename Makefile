PROJECT_VERSION ?= $(shell git describe --abbrev=0 --tags)
PROJECT_NAME = polkaTax
PROJECT_SHA ?= $(shell git rev-parse HEAD)
PROJECT_RELEASE ?= dev

lint:
	golangci-lint run \
		--deadline=5m \
		--disable-all \
		--exclude-use-default=false \
		--enable=errcheck \
		--enable=goimports \
		--enable=ineffassign \
		--enable=golint \
		--enable=unused \
		--enable=structcheck \
		--enable=staticcheck \
		--enable=varcheck \
		--enable=deadcode \
		--enable=unconvert \
		--enable=misspell \
		--enable=prealloc \
		--enable=nakedret \
		--enable=typecheck \
		./...

test: lint
	go test ./... -race -cover -covermode=atomic -coverprofile=unit_coverage.cov

build: test
	go build

.PHONY: build build_linux
