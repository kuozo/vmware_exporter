#!/usr/bin/env bash

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

export COMMIT_SHA1=$(git rev-parse --short HEAD)
export VERSION=$(git describe --tags `git rev-list --tags --max-count=1`)

go build -ldflags "-extldflags \"-static\" -X main.VERSION=${VERSION} -X main.COMMIT_SHA1=${COMMIT_SHA1} -X main.BUILD_DATE=$(date +%F-%T)" -o dist/vmware_exporter cmd/main.go
