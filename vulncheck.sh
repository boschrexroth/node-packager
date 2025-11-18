#!/bin/bash
set -e

# install govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

govulncheck ./...