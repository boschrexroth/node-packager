#!/bin/bash
set -e

DIST=./dist

# clean dist
rm -rf $DIST
mkdir -p $DIST 

# build: windows x64
./build.sh windows amd64

# build: windows amd64
./build.sh linux amd64 

# build: windows arm64
./build.sh linux arm64