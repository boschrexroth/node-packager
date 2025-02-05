#!/bin/bash

#go
old_go_version="1.22"
new_go_version="1.23"

#seek & replace
echo "go ${old_go_version} -> ${new_go_version}" 
echo
read -r -p "Press enter to continue"

#.github/workflows/*.yml
find ./.github/workflows -maxdepth 1 -name "*.yml" -type f -exec sed -i "s/go-version: '${old_go_version}'/go-version: '${new_go_version}'/g" {} \;

#go.mod
find . -maxdepth 1 -name go.mod -type f -exec sed -i "s/go ${old_go_version}/go ${new_go_version}/g" {} \;

# update go deps
go get -u ./...
go mod vendor
go mod tidy