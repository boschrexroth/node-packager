#!/bin/bash
set -e

COVERAGE_DIR=./coverage

rm -rf $COVERAGE_DIR
mkdir -p $COVERAGE_DIR

go test -v ./... -cover -coverprofile $COVERAGE_DIR/coverage.out -p=1 -failfast -timeout 100s
go tool cover -html $COVERAGE_DIR/coverage.out -o $COVERAGE_DIR/coverage.html