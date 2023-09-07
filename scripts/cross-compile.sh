#!/usr/bin/env bash

set -euxo pipefail

rm -rf ./bin

build_commands=('
    apk add make curl git \
    ; GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o bin/linux-arm64/notdone \
    ; GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/linux-amd64/notdone \
    ; GOOS=linux GOARCH=arm go build -ldflags="-s -w" -o bin/linux-arm/notdone \
    ; GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/darwin-amd64/notdone \
    ; GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o bin/darwin-arm64/notdone 
')

# run a docker container with osxcross and cross compile everything
docker run -it --rm -v $(pwd):/usr/local/src -w /usr/local/src \
	golang:alpine3.16 \
	sh -c "$build_commands"

# create archives
cd bin
for dir in $(ls -d *);
do
    cp ../README.md $dir
    cp ../LICENSE $dir
    # mkdir -p $dir/docs
    # cp -r ../docs/pages/* $dir/docs
    tar cfzv "$dir".tgz $dir
    rm -rf $dir
done
cd ..