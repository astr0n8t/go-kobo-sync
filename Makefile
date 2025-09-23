.PHONY: build

TOOLCHAIN_DIR := /kobo-lib/x-tools/arm-kobo-linux-gnueabihf/bin
CC := $(TOOLCHAIN_DIR)/arm-kobo-linux-gnueabihf-gcc
CXX := $(TOOLCHAIN_DIR)/arm-kobo-linux-gnueabihf-g++

default: build

clean:
	rm -rf x-tools/
	rm -rf go-kobo-sync/

dev:
	go build -o ./go-kobo-sync/sync_highlights

certs:
	mkdir -p ca-certs && cd ca-certs && curl -LO https://curl.se/ca/cacert.pem

build: 
	CGO_ENABLED=1 GOARCH=arm GOOS=linux CC=$(CC) CXX=$(CXX) go build -o ./go-kobo-sync/sync_highlights
	cp -r ca-certs/ go-kobo-sync/
	cp sync_highlights.sh go-kobo-sync/
	touch go-kobo-sync/config.txt
	touch go-kobo-sync/template.md

docker-build:
	docker buildx build --platform linux/amd64 --file ./Dockerfile.build --tag go-kobo-sync:build .
	docker run --platform linux/amd64 --rm -it -v `pwd`:/work go-kobo-sync:build

