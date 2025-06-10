.PHONY: build

TOOLCHAIN_DIR := $(abspath ./x-tools/arm-kobo-linux-gnueabihf/bin)
CC := $(TOOLCHAIN_DIR)/arm-kobo-linux-gnueabihf-gcc
CXX := $(TOOLCHAIN_DIR)/arm-kobo-linux-gnueabihf-g++

default: build

clean:
	rm -rf x-tools/
	rm -rf go-readwise-kobo-sync/

toolchain:
	tar -xvzf lib/kobo.tar.gz -C .

dev:
	go build -o ./go-readwise-kobo-sync/sync_highlights

build: toolchain
	CGO_ENABLED=1 GOARCH=arm GOOS=linux CC=$(CC) CXX=$(CXX) go build -o ./go-readwise-kobo-sync/sync_highlights
	cp -r ca-certs/ go-readwise-kobo-sync/
	cp sync_highlights.sh go-readwise-kobo-sync/
	touch go-readwise-kobo-sync/token.txt
