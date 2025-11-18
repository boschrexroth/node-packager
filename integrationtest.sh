#!/bin/bash
set -e

echo Integration Tests ...

TEST_PACKAGE_1=node-red-contrib-data-view
TEST_PACKAGE_2=node-red-contrib-alasql
#TEST_PACKAGE_3=node-red-contrib-modbus
TEST_DIR=./test
COVERAGE_DIR=./coverage

# Clean
rm -rf $TEST_DIR
mkdir -p $TEST_DIR 

rm -rf $COVERAGE_DIR
mkdir -p $COVERAGE_DIR

# Build binary with coverage to test dir
go build -cover -o $TEST_DIR -v ./cmd/node-packager

# Test: Execute binary with command arguments
GOCOVERDIR=$COVERAGE_DIR $TEST_DIR/node-packager $TEST_PACKAGE_1
GOCOVERDIR=$COVERAGE_DIR $TEST_DIR/node-packager --no-audit $TEST_PACKAGE_2
#GOCOVERDIR=$COVERAGE_DIR $TEST_DIR/node-packager --verbose --no-audit $TEST_PACKAGE_3  
# ...

# Print coverage in % format
go tool covdata percent -i $COVERAGE_DIR

# Convert coverage to legacy text format
go tool covdata textfmt -i $COVERAGE_DIR -o $COVERAGE_DIR/coverage.txt

# Convert coverage to HTML format
go tool cover -html $COVERAGE_DIR/coverage.txt -o $COVERAGE_DIR/coverage.html

# Print function coverage in % format
go tool cover -func $COVERAGE_DIR/coverage.txt
