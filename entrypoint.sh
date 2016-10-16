#!/bin/sh

set -e

export PATH=$PATH:/gopath/bin:/go/src/github.com/uber/go-torch/FlameGraph
export GOPATH=$GOPATH:/gopath
export HOME=/gopath/src/github.com/coccyx/gogen

cd $HOME

bash
