#!/bin/bash
result=${PWD##*/}
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o $result
#path/to/upx -3 -q $result