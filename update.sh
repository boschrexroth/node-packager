#!/bin/bash
set -e

#version

old_version="1.2.1"
new_version="1.2.1"

#go
old_go_version="1.24"
new_go_version="1.24"

#seek & replace
echo "go ${old_version} -> ${new_version}" 
echo "go ${old_go_version} -> ${new_go_version}" 
echo
read -r -p "Press enter to continue"

#build.json
find  . -maxdepth 1 -name build.json -type f -exec sed -i "s/${old_version}/${new_version}/g" {} \;

#main.go
find  ./cmd/node-packager -maxdepth 1 -name main.go -type f -exec sed -i "s/${old_version}/${new_version}/g" {} \;

#.github/workflows/*.yml
find ./.github/workflows -maxdepth 1 -name "*.yml" -type f -exec sed -i "s/go-version: '${old_go_version}'/go-version: '${new_go_version}'/g" {} \;

#go.mod
find . -maxdepth 1 -name go.mod -type f -exec sed -i "s/go ${old_go_version}/go ${new_go_version}/g" {} \;

# update go deps
go get -u ./...
go mod vendor
go mod tidy