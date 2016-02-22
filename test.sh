#!/bin/bash

GOPKG_PATH=$GOPATH/src/gopkg.in/raiqub/data.v0
GITHUB_PATH=$GOPATH/src/github.com/raiqub/data
TMP_PATH=${GOPKG_PATH}.TMP

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
	if [ -d "$TMP_PATH" ]; then
		echo "Directory '$TMP_PATH' exists" >&2
		exit 1
	fi
	
	mv "$GOPKG_PATH" "$TMP_PATH"
	ln -s "$GITHUB_PATH" "$GOPKG_PATH"
}
function finalize() {
	if [ ! -d "$TMP_PATH" ]; then
		echo "Directory '$TMP_PATH' not found" >&2
		exit 1
	fi
	if [ ! -h "$GOPKG_PATH" ]; then
		echo "'$GOPKG_PATH' is not a symbolic link" >&2
		exit 1
	fi
	if [ -d "$TMP_PATH" ]; then
		rm "$GOPKG_PATH"
		mv "$TMP_PATH" "$GOPKG_PATH"
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
