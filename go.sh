#!/bin/bash
set -euo pipefail

# Dumps the latest stable Go version

downloads=$(curl --fail --silent --show-error --location \
  --connect-timeout 10 --max-time 30 'https://go.dev/dl/?mode=json')

version=$(jq -er '
  first(
    .[] | select(.stable == true) | .version |
    select(test("^go[0-9]+\\.[0-9]+(\\.[0-9]+)?$")) |
    ltrimstr("go")
  ) // error("No stable Go version found in the official download response.")
' <<< "$downloads")

echo "$version"