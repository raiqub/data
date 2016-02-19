#!/bin/bash

GOPKG_PATH=$GOPATH/src/gopkg.in/raiqub/data.v0
GITHUB_PATH=$GOPATH/src/github.com/raiqub/data

# trap ctrl-c and call ctrl_c()
trap ctrl_c SIGINT

function prepare() {
	if [ ! -d "$GOPKG_PATH" ]; then
		echo "Directory '$GOPKG_PATH' not found" >&2
		exit 1
	fi
	if [ ! -d "$GITHUB_PATH" ]; then
		echo "Directory '$GITHUB_PATH' not found" >&2
		exit 1
	fi
	
	mv "$GOPKG_PATH" "$GOPKG_PATH.TMP"
	ln -s "$GITHUB_PATH" "$GOPKG_PATH"
}
function finalize() {
	if [ -d "$GOPKG_PATH.TMP" ]; then
		rm "$GOPKG_PATH"
		mv "$GOPKG_PATH.TMP" "$GOPKG_PATH"
	fi
}

function ctrl_c() {
	echo -en "\nExiting...\n"
	finalize
}

prepare

go test -v --race ./...
test -z "$(gofmt -s -l -w . | tee /dev/stderr)"
test -z "$(golint ./... | tee /dev/stderr)"
go vet ./...
go test -bench . -benchmem ./... | grep "Benchmark" > bench_result.txt

finalize
