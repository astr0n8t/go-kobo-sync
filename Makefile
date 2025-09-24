.PHONY: build

TOOLCHAIN_DIR := /kobo-lib/x-tools/arm-kobo-linux-gnueabihf/bin
CC := $(TOOLCHAIN_DIR)/arm-kobo-linux-gnueabihf-gcc
CXX := $(TOOLCHAIN_DIR)/arm-kobo-linux-gnueabihf-g++

KOBO_PATH := /Volumes/KOBOeReader

default: build
install: certs docker-build push-to-device

clean:
	rm -rf x-tools/
	rm -rf go-kobo-sync/

dev:
	go build -o ./go-kobo-sync/sync_highlights

certs:
	mkdir -p ca-certs && cd ca-certs && [ -f cacert.pem ] || curl -LO https://curl.se/ca/cacert.pem

build: 
	CGO_ENABLED=1 GOARCH=arm GOOS=linux CC=$(CC) CXX=$(CXX) go build -o ./go-kobo-sync/sync_highlights

docker-build:
	docker buildx build --platform linux/amd64 --file ./Dockerfile.build --tag go-kobo-sync:build .
	docker run --platform linux/amd64 --rm -it -v `pwd`:/work go-kobo-sync:build
	cp -r ca-certs/ go-kobo-sync/
	cp install/* go-kobo-sync/
	cp --update=none example/config go-kobo-sync/config
	cp --update=none example/template.md go-kobo-sync/template.md
	cp --update=none example/header_template.md go-kobo-sync/header_template.md

push-to-device:
	mkdir -p $(KOBO_PATH)/.adds/go-kobo-sync/
	cp -r go-kobo-sync/* $(KOBO_PATH)/.adds/go-kobo-sync/

nm-install:
	cat install/nm.config >> $(KOBO_PATH)/.adds/nm/config
