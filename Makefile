VERSION=$(shell cat VERSION 2> /dev/null || echo "custom-`date +%Y-%m-%d-%H-%M`" )

all: build test doc install

build:
	go mod download
	go build -a -gcflags=-trimpath=$(go env GOPATH) -asmflags=-trimpath=$(go env GOPATH) -ldflags "-X github.com/wunderio/silta-cli/internal/common.Version=$(VERSION)" -o silta

test:
	go test ./tests

install:
	cp ./silta ~/.local/bin/silta

doc:
	silta doc

clean:
	go clean -r -x
	-rm -rf silta
