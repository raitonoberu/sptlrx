.PHONY: build test man

build:
	go build -ldflags '-w -s'

test:
	go test ./...

man:
	go-md2man -in man/sptlrx.5.md -out man/sptlrx.5
