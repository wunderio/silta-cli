VERSION=$(shell cat VERSION 2> /dev/null || echo "custom")

all: build test upload

build:
	go mod download
	go build -a -gcflags=-trimpath=$(go env GOPATH) -asmflags=-trimpath=$(go env GOPATH) -ldflags "-X github.com/wunderio/silta-cli/internal/common.Version=$(VERSION)" -o silta

test:
	go test ./tests

upload:
	gsutil cp ./silta gs://silta-cli-test/
	gsutil setmeta -h "Cache-Control: no-store, max-age=10" "gs://silta-cli-test/silta"

clean:
	go clean -r -x
	-rm -rf silta
