#!/bin/bash

PATH="$PATH:$GOBIN"
protoc --proto_path=../../rms-shared/api/:. --micro_out=. --go_out=. rms-torrent.proto