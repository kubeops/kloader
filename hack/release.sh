#!/bin/bash

set -x
set -eou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT="$GOPATH/src/github.com/appscode/kloader"
rm -rf $REPO_ROOT/dist

env APPSCODE_ENV=prod ./hack/make.py build
./hack/make.py push
env APPSCODE_ENV=prod ./hack/make.py push

rm -rf $REPO_ROOT/dist/.tag
