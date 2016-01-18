#!/bin/bash

go test -v --race ./...
test -z "$(gofmt -s -l -w . | tee /dev/stderr)"
test -z "$(golint ./... | tee /dev/stderr)"
go vet ./...
go test -bench . -benchmem ./... | grep "Benchmark" > bench_result.txt
