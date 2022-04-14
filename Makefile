SHELL=/bin/bash

export PATH:=/usr/local/go/bin:~/go/bin/:$(PATH)

test: vet
	go test -v ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...
