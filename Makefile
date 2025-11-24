SHELL := bash
.ONESHELL:
.SHELLFLAGS := -u -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules
.SILENT:

GO_BUILD_FLAGS := -buildvcs=true

help:
	echo '  make clean         remove generated files'
	echo '  make build         build the project from scratch'
	echo '  make prunejuice    build executable if not already built'
	echo '  make check         execute tests and checks'
.PHONY: help

# go builds are fast enough that we can just build on demand instead of trying to do any fancy
# change detection
build: clean prunejuice
.PHONY: build

prunejuice:
	go build ${GO_BUILD_FLAGS} ./cmd/prunejuice

clean:
	rm -f ./prunejuice
.PHONY: clean

check:
	go test ${GO_BUILD_FLAGS} ./...
	go vet ./...
.PHONY: check
