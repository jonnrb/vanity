#!/bin/bash
set -e
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" .
upx vanityserver
docker build -t jonnrb/vanity .
rm vanityserver
