#!/bin/bash
set -e

silent=${1:-false}

# version
old_version="1.2.2"
new_version="1.2.2"

# go
old_go_version="1.25"
new_go_version="1.26"

echo "go ${old_version} -> ${new_version}" 
echo "go ${old_go_version} -> ${new_go_version}" 
echo

if [[ $silent == false ]]; then
  read -r -p "Press enter to continue"
fi

# main.go
find  ./cmd/node-packager -maxdepth 1 -name main.go -type f -exec sed -i "s/${old_version}/${new_version}/g" {} \;

# .github/workflows/*.yml
find ./.github/workflows -maxdepth 1 -name "*.yml" -type f -exec sed -i "s/go-version: '${old_go_version}'/go-version: '${new_go_version}'/g" {} \;

# go.mod
find . -maxdepth 1 -name go.mod -type f -exec sed -i "s/go ${old_go_version}/go ${new_go_version}/g" {} \;

# update go deps
go get -u ./...
go mod vendor
go mod tidy

./build.sh
./vulncheck.sh