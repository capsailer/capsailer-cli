.PHONY: build clean test install

BINARY_NAME=capsailer
VERSION=0.1.0
BUILD_TIME=$(shell date +%FT%T%z)
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

build:
	go build ${LDFLAGS} -o ${BINARY_NAME} cmd/capsailer/main.go

clean:
	go clean
	rm -f ${BINARY_NAME}
	rm -rf capsailer-build
	rm -f *.tar.gz

test:
	go test -v ./...

install: build
	cp ${BINARY_NAME} /usr/local/bin/

run:
	./${BINARY_NAME}

# Example targets for convenience
example-manifest:
	cp example-manifest.yaml manifest.yaml

example-build: example-manifest
	./${BINARY_NAME} build --manifest manifest.yaml --output capsailer-bundle.tar.gz

# Docker build target (future enhancement)
docker-build:
	docker build -t capsailer:${VERSION} . 