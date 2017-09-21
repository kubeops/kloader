#!/bin/bash

set -x
set -eou pipefail

env APPSCODE_ENV=prod ./hack/make.py build
./hack/make.py push
env APPSCODE_ENV=prod ./hack/make.py push
