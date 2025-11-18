#!/bin/bash
set -e

# arguments
GO_OS=${1:-linux}
GO_ARCH=${2:-amd64}

# setup
DIST=./dist
OUT_DIR=${DIST}/${GO_OS}_${GO_ARCH}

go vet ./...

# install and run go linter
# sudo snap install golangci-lint --classic
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run --enable=gosec --modules-download-mode=vendor

echo "working dir: $(pwd)"
echo "build: GOOS=$GO_OS GOARCH=$GO_ARCH OUTDIR=$OUT_DIR ..."
mkdir -p $OUT_DIR 

GOOS=$GO_OS GOARCH=$GO_ARCH go build -o $OUT_DIR -v ./cmd/node-packager