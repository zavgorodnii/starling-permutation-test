#!/usr/bin/env make

all: deps build test

deps:
	go get github.com/pkg/errors
	go get github.com/tealeg/xlsx
	go get gonum.org/v1/plot/...
	go get github.com/gonum/stat/distuv
	go get github.com/stretchr/testify/assert

build:
	GOOS=windows GOARCH=amd64 go build -o ./bin/spt_win_x86-64.exe main.go
	GOOS=darwin GOARCH=amd64 go build -o ./bin/spt_darwin_x86-64 main.go
	GOOS=linux GOARCH=amd64 go build -o ./bin/spt_linux_x86-64 main.go

test:
	go test github.com/starling-permutation-test/src
