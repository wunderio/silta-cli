VERSION=$(shell cat VERSION 2> /dev/null || echo "custom-`date +%Y-%m-%d-%H-%M`" )

all: build test

build:
	go mod download
	go build -a -gcflags=-trimpath=$(go env GOPATH) -asmflags=-trimpath=$(go env GOPATH) -ldflags "-X github.com/wunderio/silta-cli/internal/common.Version=$(VERSION)" -o silta

build_move:
	make build
	cp silta bintest/

test:
	go test ./tests

install:
	cp ./silta ~/.local/bin/silta

clean:
	go clean -r -x
	-rm -rf silta
