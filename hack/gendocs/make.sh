#!/usr/bin/env bash

pushd $GOPATH/src/github.com/appscode/kloader/hack/gendocs
go run main.go
popd
