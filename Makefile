SHELL=/bin/bash

export PATH:=/usr/local/go/bin:~/go/bin/:$(PATH)

GOFMT_FILES?=$$(find . -name '*.go')

default: build

test: vet
	go test -v ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...
