#!/bin/bash
result=${PWD##*/}
go build -ldflags "-s -w" -o $result
path/to/upx -3 -q $result