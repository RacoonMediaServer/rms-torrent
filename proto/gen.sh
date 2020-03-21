#!/bin/bash

PATH="$PATH:$GOBIN"
protoc --proto_path=$GOPATH/src:. --micro_out=. --go_out=. rtorrent.proto